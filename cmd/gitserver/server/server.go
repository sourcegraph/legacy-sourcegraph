package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/honey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

var runCommand = func(cmd *exec.Cmd) (error, int) { // mocked by tests
	err := cmd.Run()
	exitStatus := -10810
	if cmd.ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return err, exitStatus
}
var skipCloneForTests = false // set by tests

// Server is a gitserver server.
type Server struct {
	// ReposDir is the path to the base directory for gitserver storage.
	ReposDir string

	// GithubAccessToken if set will be used for all fetches and clones to
	// Github. This allows gitserver to access private code. Should only
	// be used on private servers, not Sourcegraph.com
	GithubAccessToken string

	// InsecureSkipCheckVerifySSH controls whether the client verifies the
	// SSH server's certificate or host key. If InsecureSkipCheckVerifySSH
	// is true, the program is susceptible to a man-in-the-middle
	// attack. This should only be used for testing.
	InsecureSkipCheckVerifySSH bool

	// cloning tracks repositories (key is '/'-separated path) that are
	// in the process of being cloned.
	cloningMu sync.Mutex
	cloning   map[string]struct{}

	updateRepo        chan<- updateRepoRequest
	repoUpdateLocksMu sync.Mutex // protects the map below and also updates to locks.once
	repoUpdateLocks   map[string]*locks
}

type locks struct {
	once *sync.Once  // consolidates multiple waiting updates
	mu   *sync.Mutex // prevents updates running in parallel
}

type updateRepoRequest struct {
	repo string
}

// shortGitCommandTimeout returns the timeout for git commands that should not
// take a long time. Some commands such as "git archive" are allowed more time
// than "git rev-parse", so this will return an appropriate timeout given the
// command.
func shortGitCommandTimeout(args []string) time.Duration {
	if len(args) < 1 {
		return time.Minute
	}
	switch args[0] {
	case "archive":
		// 10 minutes is a long time, but this never blocks a user request for
		// this long. Even repos that are not that large can take a long time,
		// for example a search over all repos in an organization may have
		// several large repos. All of those repos will be competing for IO =>
		// we need a larger timeout.
		return 10 * time.Minute

	default:
		return time.Minute
	}
}

// This is a timeout for long git commands like clone or remote update.
// that may take a while for large repos. These types of commands should
// be run in the background.
var longGitCommandTimeout = time.Hour

// Handler returns the http.Handler that should be used to serve requests.
func (s *Server) Handler() http.Handler {
	s.cloning = make(map[string]struct{})
	s.updateRepo = s.repoUpdateLoop()
	s.repoUpdateLocks = make(map[string]*locks)

	mux := http.NewServeMux()
	mux.HandleFunc("/exec", s.handleExec)
	mux.HandleFunc("/list", s.handleList)
	mux.HandleFunc("/is-repo-cloneable", s.handleIsRepoCloneable)
	mux.HandleFunc("/repo-from-remote-url", s.handleRepoFromRemoteURL)
	return mux
}

func cloneCmd(ctx context.Context, origin, dir string) *exec.Cmd {
	return exec.CommandContext(ctx, "git", "clone", "--mirror", origin, dir)
}

func (s *Server) setCloneLock(dir string) {
	s.cloningMu.Lock()
	s.cloning[dir] = struct{}{}
	s.cloningMu.Unlock()
}

func (s *Server) releaseCloneLock(dir string) {
	s.cloningMu.Lock()
	delete(s.cloning, dir)
	s.cloningMu.Unlock()
}

func (s *Server) handleIsRepoCloneable(w http.ResponseWriter, r *http.Request) {
	var req protocol.IsRepoCloneableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cloneable := s.isCloneable(r.Context(), req.Repo)
	if err := json.NewEncoder(w).Encode(cloneable); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleRepoFromRemoteURL(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoFromRemoteURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repo := reverse(req.RemoteURL)
	if err := json.NewEncoder(w).Encode(repo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleExec(w http.ResponseWriter, r *http.Request) {
	var req protocol.ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Flush writes more aggressively than standard net/http so that clients
	// with a context deadline see as much partial response body as possible.
	fw := newFlushingResponseWriter(w)
	w = fw
	defer fw.Close()

	ctx, cancel := context.WithTimeout(r.Context(), shortGitCommandTimeout(req.Args))
	defer cancel()

	start := time.Now()
	exitStatus := -10810 // sentinel value to indicate not set
	var stdoutN, stderrN int64
	var status string
	var errStr string

	req.Repo = protocol.NormalizeRepo(req.Repo)

	// Instrumentation
	{
		repo := repotrackutil.GetTrackedRepo(req.Repo)
		cmd := ""
		if len(req.Args) > 0 {
			cmd = req.Args[0]
		}
		execRunning.WithLabelValues(cmd, repo).Inc()
		defer func() {
			duration := time.Since(start)
			execRunning.WithLabelValues(cmd, repo).Dec()
			execDuration.WithLabelValues(cmd, repo, status).Observe(duration.Seconds())
			if honey.Enabled() {
				ev := honey.Event("gitserver-exec")
				ev.AddField("repo", req.Repo)
				ev.AddField("cmd", cmd)
				ev.AddField("args", strings.Join(req.Args, " "))
				ev.AddField("duration_ms", duration.Seconds()*1000)
				ev.AddField("stdout_size", stdoutN)
				ev.AddField("stderr_size", stderrN)
				ev.AddField("exit_status", exitStatus)
				ev.AddField("status", status)
				if errStr != "" {
					ev.AddField("error", errStr)
				}
				ev.Send()
			}
		}()
	}

	dir := path.Join(s.ReposDir, req.Repo)
	s.cloningMu.Lock()
	_, cloneInProgress := s.cloning[dir]
	s.cloningMu.Unlock()
	if cloneInProgress || strings.ToLower(req.Repo) == "github.com/sourcegraphtest/alwayscloningtest" {
		status = "clone-in-progress"
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(&protocol.NotFoundPayload{CloneInProgress: true})
		return
	}
	if !repoCloned(dir) {
		if origin := OriginMap(req.Repo); origin != "" && s.isCloneable(ctx, req.Repo) {
			// Mark this repo as currently being cloned. We have to check again if someone else isn't already
			// cloning since we released the lock. We released the lock since isCloneable is a potentially
			// slow operation.
			s.cloningMu.Lock()
			_, cloneInProgress := s.cloning[dir]
			if !cloneInProgress {
				s.cloning[dir] = struct{}{} // Mark this repo as currently being cloned.
			}
			s.cloningMu.Unlock()

			if skipCloneForTests {
				s.releaseCloneLock(dir)
			} else if !cloneInProgress {
				go func() {
					// Create a new context because this is in a background goroutine.
					ctx, cancel := context.WithTimeout(context.Background(), longGitCommandTimeout)
					defer func() {
						cancel()
						s.releaseCloneLock(dir)
					}()

					cmd := cloneCmd(ctx, origin, dir)
					if output, err := s.runWithRemoteOpts(cmd, req.Repo); err != nil {
						log15.Error("clone failed", "error", err, "output", string(output))
						return
					}
				}()
			}

			status = "clone-in-progress"
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&protocol.NotFoundPayload{CloneInProgress: true})
			return
		}

		status = "repo-not-found"
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(&protocol.NotFoundPayload{CloneInProgress: false})
		return
	}

	if req.EnsureRevision != "" {
		cmd := exec.CommandContext(ctx, "git", "rev-parse", req.EnsureRevision)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			// Revision not found, update before running actual command.
			s.doRepoUpdate(ctx, req.Repo)
		}
	}

	w.Header().Set("Trailer", "X-Exec-Error, X-Exec-Exit-Status, X-Exec-Stderr")
	w.WriteHeader(http.StatusOK)

	var stderrBuf bytes.Buffer
	stdoutW := &writeCounter{w: w}
	stderrW := &writeCounter{w: &stderrBuf}

	cmd := exec.CommandContext(ctx, "git", req.Args...)
	cmd.Dir = dir
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	var err error
	err, exitStatus = runCommand(cmd)
	if err != nil {
		errStr = err.Error()
	}

	status = strconv.Itoa(exitStatus)
	stdoutN = stdoutW.n
	stderrN = stderrW.n

	stderr := stderrBuf.String()
	if len(stderr) > 1024 {
		stderr = stderr[:1024]
	}

	// write trailer
	w.Header().Set("X-Exec-Error", errStr)
	w.Header().Set("X-Exec-Exit-Status", status)
	w.Header().Set("X-Exec-Stderr", string(stderr))

	if !skipCloneForTests {
		s.updateRepo <- updateRepoRequest{req.Repo}
	}
}

// testRepoExists is a test fixture that overrides the return value
// for isCloneable when it is set.
var testRepoExists func(ctx context.Context, origin string, repoURI string) bool

// isCloneable checks to see if the repo is cloneable.
func (s *Server) isCloneable(ctx context.Context, repo string) bool {
	ctx, cancel := context.WithTimeout(ctx, shortGitCommandTimeout(nil))
	defer cancel()

	if strings.ToLower(repo) == "github.com/sourcegraphtest/alwayscloningtest" {
		return true
	}

	repo = protocol.NormalizeRepo(repo)
	origin := OriginMap(repo)
	if origin == "" {
		return false
	}

	if testRepoExists != nil {
		return testRepoExists(ctx, origin, repo)
	}
	cmd := exec.CommandContext(ctx, "git", "ls-remote", origin, "HEAD")
	_, err := s.runWithRemoteOpts(cmd, repo)
	return err == nil
}

var execRunning = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "exec_running",
	Help:      "number of gitserver.Command running concurrently.",
}, []string{"cmd", "repo"})
var execDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "exec_duration_seconds",
	Help:      "gitserver.Command latencies in seconds.",
	Buckets:   traceutil.UserLatencyBuckets,
}, []string{"cmd", "repo", "status"})

func init() {
	prometheus.MustRegister(execRunning)
	prometheus.MustRegister(execDuration)
}

func (s *Server) repoUpdateLoop() chan<- updateRepoRequest {
	updateRepo := make(chan updateRepoRequest, 10)
	lastCheckAt := make(map[string]time.Time)

	go func() {
		for req := range updateRepo {
			if t, ok := lastCheckAt[req.repo]; ok && time.Now().Before(t.Add(10*time.Second)) {
				continue // git data still fresh
			}
			lastCheckAt[req.repo] = time.Now()
			go func(req updateRepoRequest) {
				// Create a new context with a new timeout (instead of passing one through updateRepoRequest)
				// because the ctx of the updateRepoRequest sender will get cancelled before this goroutine runs.
				ctx, cancel := context.WithTimeout(context.Background(), longGitCommandTimeout)
				defer cancel()
				s.doRepoUpdate(ctx, req.repo)
			}(req)
		}
	}()

	return updateRepo
}

var headBranchPattern = regexp.MustCompile("HEAD branch: (.+?)\\n")

func (s *Server) doRepoUpdate(ctx context.Context, repo string) {
	s.repoUpdateLocksMu.Lock()
	l, ok := s.repoUpdateLocks[repo]
	if !ok {
		l = &locks{
			once: new(sync.Once),
			mu:   new(sync.Mutex),
		}
		s.repoUpdateLocks[repo] = l
	}
	once := l.once
	mu := l.mu
	s.repoUpdateLocksMu.Unlock()

	once.Do(func() {
		mu.Lock() // Prevent multiple updates in parallel. It works fine, but it wastes resources.
		defer mu.Unlock()

		s.repoUpdateLocksMu.Lock()
		l.once = new(sync.Once) // Make new requests wait for next update.
		s.repoUpdateLocksMu.Unlock()

		s.doRepoUpdate2(ctx, repo)
	})
}

func (s *Server) doRepoUpdate2(ctx context.Context, repo string) {
	cmd := exec.CommandContext(ctx, "git", "remote", "update", "--prune")
	cmd.Dir = path.Join(s.ReposDir, repo)
	if output, err := s.runWithRemoteOpts(cmd, repo); err != nil {
		log15.Error("Failed to update", "repo", repo, "error", err, "output", string(output))
		return
	}

	headBranch := "master"

	// try to fetch HEAD from origin
	cmd = exec.CommandContext(ctx, "git", "remote", "show", "origin")
	cmd.Dir = path.Join(s.ReposDir, repo)
	output, err := s.runWithRemoteOpts(cmd, repo)
	if err != nil {
		log15.Error("Failed to fetch remote info", "repo", repo, "error", err, "output", string(output))
		return
	}
	submatches := headBranchPattern.FindSubmatch(output)
	if len(submatches) == 2 {
		submatch := string(submatches[1])
		if submatch != "(unknown)" {
			headBranch = string(submatch)
		}
	}

	// check if branch pointed to by HEAD exists
	cmd = exec.CommandContext(ctx, "git", "rev-parse", headBranch)
	cmd.Dir = path.Join(s.ReposDir, repo)
	if err := cmd.Run(); err != nil {
		// branch does not exist, pick first branch
		cmd := exec.CommandContext(ctx, "git", "branch")
		cmd.Dir = path.Join(s.ReposDir, repo)
		list, err := cmd.Output()
		if err != nil {
			log15.Error("Failed to list branches", "repo", repo, "error", err, "output", string(output))
			return
		}
		lines := strings.Split(string(list), "\n")
		branch := strings.TrimPrefix(strings.TrimPrefix(lines[0], "* "), "  ")
		if branch != "" {
			headBranch = branch
		}
	}

	// set HEAD
	cmd = exec.CommandContext(ctx, "git", "symbolic-ref", "HEAD", "refs/heads/"+headBranch)
	cmd.Dir = path.Join(s.ReposDir, repo)
	if output, err := cmd.CombinedOutput(); err != nil {
		log15.Error("Failed to set HEAD", "repo", repo, "error", err, "output", string(output))
		return
	}
}
