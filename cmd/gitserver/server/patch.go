package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/unpack"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var patchID uint64

func (s *Server) handleCreateCommitFromPatchBinary(w http.ResponseWriter, r *http.Request) {
	var req protocol.CreateCommitFromPatchRequest
	var resp protocol.CreateCommitFromPatchResponse
	var status int

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := new(protocol.CreateCommitFromPatchResponse)
		resp.SetError("", "", "", errors.Wrap(err, "decoding CreateCommitFromPatchRequest"))
		status = http.StatusBadRequest
	} else {
		status, resp = s.createCommitFromPatch(r.Context(), req)
	}

	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) createCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (int, protocol.CreateCommitFromPatchResponse) {
	logger := s.Logger.Scoped("createCommitFromPatch", "").
		With(
			log.String("repo", string(req.Repo)),
			log.String("baseCommit", string(req.BaseCommit)),
			log.String("targetRef", req.TargetRef),
		)

	var resp protocol.CreateCommitFromPatchResponse

	repo := string(protocol.NormalizeRepo(req.Repo))
	repoGitDir := filepath.Join(s.ReposDir, repo, ".git")
	if _, err := os.Stat(repoGitDir); os.IsNotExist(err) {
		repoGitDir = filepath.Join(s.ReposDir, repo)
		if _, err := os.Stat(repoGitDir); os.IsNotExist(err) {
			resp.SetError(repo, "", "", errors.Wrap(err, "gitserver: repo does not exist"))
			return http.StatusInternalServerError, resp
		}
	}

	var (
		remoteURL *vcs.URL
		err       error
	)

	if req.Push != nil && req.Push.RemoteURL != "" {
		remoteURL, err = vcs.ParseURL(req.Push.RemoteURL)
	} else {
		remoteURL, err = s.getRemoteURL(ctx, req.Repo)
	}

	ref := req.TargetRef
	// If the push is to a Gerrit project,we need to push to a magic ref.
	if req.PushRef != nil && *req.PushRef != "" {
		ref = *req.PushRef
	}
	if req.UniqueRef {
		refs, err := repoRemoteRefs(ctx, remoteURL, ref)
		if err != nil {
			logger.Error("Failed to get remote refs", log.Error(err))
			resp.SetError(repo, "", "", errors.Wrap(err, "repoRemoteRefs"))
			return http.StatusInternalServerError, resp
		}

		retry := 1
		tmp := ref
		for {
			if _, ok := refs[tmp]; !ok {
				break
			}
			tmp = ref + "-" + strconv.Itoa(retry)
			retry++
		}
		ref = tmp
	}

	if req.Push != nil && req.PushRef == nil {
		ref = ensureRefPrefix(ref)
	}

	if err != nil {
		logger.Error("Failed to get remote URL", log.Error(err))
		resp.SetError(repo, "", "", errors.Wrap(err, "repoRemoteURL"))
		return http.StatusInternalServerError, resp
	}

	redactor := newURLRedactor(remoteURL)
	defer func() {
		if resp.Error != nil {
			resp.Error.Command = redactor.redact(resp.Error.Command)
			resp.Error.CombinedOutput = redactor.redact(resp.Error.CombinedOutput)
			if resp.Error.InternalError != "" {
				resp.Error.InternalError = redactor.redact(resp.Error.InternalError)
			}
		}
	}()

	// Ensure tmp directory exists
	tmpRepoDir, err := s.tempDir("patch-repo-")
	if err != nil {
		resp.SetError(repo, "", "", errors.Wrap(err, "gitserver: make tmp repo"))
		return http.StatusInternalServerError, resp
	}
	//defer cleanUpTmpRepo(logger, tmpRepoDir)

	argsToString := func(args []string) string {
		return strings.Join(args, " ")
	}

	// Temporary logging command wrapper
	prefix := fmt.Sprintf("%d %s ", atomic.AddUint64(&patchID, 1), repo)
	run := func(cmd *exec.Cmd, reason string) ([]byte, error) {
		if !gitdomain.IsAllowedGitCmd(logger, cmd.Args[1:]) {
			return nil, errors.New("command not on allow list")
		}

		t := time.Now()
		// runRemoteGitCommand since one of our commands could be git push
		out, err := runRemoteGitCommand(ctx, s.recordingCommandFactory.Wrap(ctx, s.Logger, cmd), true, nil)

		logger := logger.With(
			log.String("prefix", prefix),
			log.String("command", argsToString(cmd.Args)),
			log.Duration("duration", time.Since(t)),
			log.String("output", string(out)),
		)

		if err != nil {
			resp.SetError(repo, argsToString(cmd.Args), string(out), errors.Wrap(err, "gitserver: "+reason))
			logger.Warn("command failed", log.Error(err))
		} else {
			logger.Info("command ran successfully")
		}
		return out, err
	}

	tmpGitPathEnv := "GIT_DIR=" + filepath.Join(tmpRepoDir, ".git")

	tmpObjectsDir := filepath.Join(tmpRepoDir, ".git", "objects")
	repoObjectsDir := filepath.Join(repoGitDir, "objects")

	altObjectsEnv := "GIT_ALTERNATE_OBJECT_DIRECTORIES=" + repoObjectsDir

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = tmpRepoDir
	cmd.Env = append(os.Environ(), tmpGitPathEnv)

	if _, err := run(cmd, "init tmp repo"); err != nil {
		return http.StatusInternalServerError, resp
	}

	cmd = exec.CommandContext(ctx, "git", "reset", "-q", string(req.BaseCommit))
	cmd.Dir = tmpRepoDir
	cmd.Env = append(os.Environ(), tmpGitPathEnv, altObjectsEnv)

	if out, err := run(cmd, "basing staging on base rev"); err != nil {
		logger.Error("Failed to base the temporary repo on the base revision",
			log.String("output", string(out)),
		)
		return http.StatusInternalServerError, resp
	}

	applyArgs := append([]string{"apply", "--cached"}, req.GitApplyArgs...)

	cmd = exec.CommandContext(ctx, "git", applyArgs...)
	cmd.Dir = tmpRepoDir
	cmd.Env = append(os.Environ(), tmpGitPathEnv, altObjectsEnv)
	cmd.Stdin = bytes.NewReader(req.Patch)

	if out, err := run(cmd, "applying patch"); err != nil {
		logger.Error("Failed to apply patch", log.String("output", string(out)))
		return http.StatusBadRequest, resp
	}

	messages := req.CommitInfo.Messages
	if len(messages) == 0 {
		messages = []string{"<Sourcegraph> Creating commit from patch"}
	}
	authorName := req.CommitInfo.AuthorName
	if authorName == "" {
		authorName = "Sourcegraph"
	}
	authorEmail := req.CommitInfo.AuthorEmail
	if authorEmail == "" {
		authorEmail = "support@sourcegraph.com"
	}
	committerName := req.CommitInfo.CommitterName
	if committerName == "" {
		committerName = authorName
	}
	committerEmail := req.CommitInfo.CommitterEmail
	if committerEmail == "" {
		committerEmail = authorEmail
	}

	gitCommitArgs := []string{"commit"}
	for _, m := range messages {
		gitCommitArgs = append(gitCommitArgs, "-m", stylizeCommitMessage(m))
	}
	cmd = exec.CommandContext(ctx, "git", gitCommitArgs...)

	cmd.Dir = tmpRepoDir
	cmd.Env = append(os.Environ(), []string{
		tmpGitPathEnv,
		altObjectsEnv,
		fmt.Sprintf("GIT_COMMITTER_NAME=%s", committerName),
		fmt.Sprintf("GIT_COMMITTER_EMAIL=%s", committerEmail),
		fmt.Sprintf("GIT_AUTHOR_NAME=%s", authorName),
		fmt.Sprintf("GIT_AUTHOR_EMAIL=%s", authorEmail),
		fmt.Sprintf("GIT_COMMITTER_DATE=%v", req.CommitInfo.Date),
		fmt.Sprintf("GIT_AUTHOR_DATE=%v", req.CommitInfo.Date),
	}...)

	if out, err := run(cmd, "committing patch"); err != nil {
		logger.Error("Failed to commit patch.", log.String("output", string(out)))
		return http.StatusInternalServerError, resp
	}

	cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = tmpRepoDir
	cmd.Env = append(os.Environ(), tmpGitPathEnv, altObjectsEnv)

	// We don't use 'run' here as we only want stdout
	out, err := cmd.Output()
	if err != nil {
		resp.SetError(repo, argsToString(cmd.Args), string(out), errors.Wrap(err, "gitserver: retrieving new commit id"))
		return http.StatusInternalServerError, resp
	}
	cmtHash := strings.TrimSpace(string(out))

	// Move objects from tmpObjectsDir to repoObjectsDir.
	err = filepath.Walk(tmpObjectsDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(tmpObjectsDir, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(repoObjectsDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
			return err
		}
		// do the actual move. If dst exists we can ignore the error since it
		// will contain the same content (content addressable FTW).
		if err := os.Rename(path, dst); err != nil && !os.IsExist(err) {
			return err
		}
		return nil
	})
	if err != nil {
		resp.SetError(repo, "", "", errors.Wrap(err, "copying git objects"))
		return http.StatusInternalServerError, resp
	}

	if req.Push != nil {
		if req.Push.P4Credentials != nil {
			// Perforce credentials are filled out, so push to a Perforce server instead of a Git host
			cid, err := s.shelveChangelist(ctx, req, tmpGitPathEnv, altObjectsEnv)
			if err != nil {
				resp.SetError(repo, "", "", err)
				return http.StatusInternalServerError, resp
			}
			ncid, err := strconv.ParseUint(cid, 10, 64)
			if err != nil {
				resp.SetError(repo, "", "", errors.Wrap(err, "invalid changelist id: "+cid))
				return http.StatusInternalServerError, resp
			}

			resp.ChangelistId = &ncid
		} else {
			cmd = exec.CommandContext(ctx, "git", "push", "--force", remoteURL.String(), fmt.Sprintf("%s:%s", cmtHash, ref))
			cmd.Dir = repoGitDir

			// If the protocol is SSH and a private key was given, we want to
			// use it for communication with the code host.
			if remoteURL.IsSSH() && req.Push.PrivateKey != "" && req.Push.Passphrase != "" {
				// We set up an agent here, which sets up a socket that can be provided to
				// SSH via the $SSH_AUTH_SOCK environment variable and the goroutine to drive
				// it in the background.
				// This is used to pass the private key to be used when pushing to the remote,
				// without the need to store it on the disk.
				agent, err := newSSHAgent(logger, []byte(req.Push.PrivateKey), []byte(req.Push.Passphrase))
				if err != nil {
					resp.SetError(repo, "", "", errors.Wrap(err, "gitserver: error creating ssh-agent"))
					return http.StatusInternalServerError, resp
				}
				go agent.Listen()
				// Make sure we shut this down once we're done.
				defer agent.Close()

				cmd.Env = append(
					os.Environ(),
					[]string{
						fmt.Sprintf("SSH_AUTH_SOCK=%s", agent.Socket()),
					}...,
				)
			}

			if out, err = run(cmd, "pushing ref"); err != nil {
				logger.Error("Failed to push", log.String("commit", cmtHash), log.String("output", string(out)))
				return http.StatusInternalServerError, resp
			}
		}
	}
	resp.Rev = "refs/" + strings.TrimPrefix(ref, "refs/")

	if req.PushRef == nil {
		cmd = exec.CommandContext(ctx, "git", "update-ref", "--", ref, cmtHash)
		cmd.Dir = repoGitDir

		if out, err = run(cmd, "creating ref"); err != nil {
			logger.Error("Failed to create ref for commit.", log.String("commit", cmtHash), log.String("output", string(out)))
			return http.StatusInternalServerError, resp
		}
	}

	return http.StatusOK, resp
}

func stylizeCommitMessage(message string) string {
	if styleMessage(message) {
		return fmt.Sprintf("%q", message)
	}
	return message
}

func styleMessage(message string) bool {
	return !strings.HasPrefix(message, "Change-Id: I")
}

var cidPattern = lazyregexp.New(`Change (\d+) files shelved`)

func (s *Server) shelveChangelist(ctx context.Context, req protocol.CreateCommitFromPatchRequest, tmpGitPathEnv, altObjectsEnv string) (string, error) {
	logger := s.Logger.Scoped("createCommitFromPatch", "").
		With(
			log.String("repo", string(req.Repo)),
			log.String("baseCommit", string(req.BaseCommit)),
			log.String("targetRef", req.TargetRef),
		)

	// use the name of the target branch as the perforce client name
	p4client := strings.TrimPrefix(req.TargetRef, "refs/heads/")

	// do all work in (another) temporary directory
	tmpClientDir, err := s.tempDir("perforce-client-")
	if err != nil {
		return "", errors.Wrap(err, "gitserver: make tmp repo for Perforce client")
	}
	defer cleanUpTmpRepo(logger, tmpClientDir)

	// we'll need these environment variables for subsequent commands
	commonEnv := append(os.Environ(), []string{
		tmpGitPathEnv,
		altObjectsEnv,
		fmt.Sprintf("P4PORT=%s", req.Push.RemoteURL),
		fmt.Sprintf("P4USER=%s", req.Push.P4Credentials.P4User),
		fmt.Sprintf("P4PASSWD=%s", req.Push.P4Credentials.P4Passwd),
		fmt.Sprintf("P4CLIENT=%s", p4client),
	}...)

	// get the commit message from the base commit so that we can extract the changelist id from it
	cmd := exec.CommandContext(ctx, "git", "show", "--no-patch", "--pretty=format:%B", string(req.BaseCommit))
	cmd.Dir = tmpClientDir
	cmd.Env = commonEnv

	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "gitserver: retrieving base commit message")
	}

	// extract the base changelist id from the commit message
	baseCID, err := perforce.GetP4ChangelistID(string(out))
	if err != nil {
		return "", errors.Wrap(err, "gitserver: retrieving base changelist id")
	}

	// get the list of files involved in the commit
	cmd = exec.CommandContext(ctx, "git", "diff-tree", "--no-commit-id", "--name-only", "-r", string(req.BaseCommit))
	cmd.Dir = tmpClientDir
	cmd.Env = commonEnv
	out, err = cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "gitserver: retrieving files in base commit")
	}
	fileList := strings.Split(string(out), "\n")
	if len(fileList) <= 0 {
		return "", errors.New("gitserver: no files in base commit")
	}

	// format a description for the client spec and the changelist
	// from the commit message(s)
	// be sure to indent lines so that it fits the Perforce form format
	desc := "batch change"
	if len(req.CommitInfo.Messages) > 0 {
		desc = strings.ReplaceAll(strings.Join(req.CommitInfo.Messages, "\n"), "\n", "\n\t")
	}

	// create a Perforce client spec to use for creating the changelist
	clientSpec := fmt.Sprintf(
		`Client:	%s
Owner:	%s
Description:
	%s
Root:	%s
Options:	noallwrite noclobber nocompress unlocked nomodtime normdir
SubmitOptions:	submitunchanged
LineEnd:	local
View:	%s... //%s/...
`,
		p4client,
		req.Push.RemoteURL,
		desc,
		tmpClientDir,
		req.Repo,
		p4client,
	)
	cmd = exec.CommandContext(ctx, "p4", "client", "-i")
	cmd.Dir = tmpClientDir
	cmd.Env = commonEnv
	cmd.Stdin = bytes.NewReader([]byte(clientSpec))
	out, err = cmd.CombinedOutput()
	if err != nil {
		logger.Error("p4 client failed", log.String("output", string(out)))
		return "", errors.Wrap(err, "gitserver: creating a Perforce client spec")
	}

	// get the files from the Perforce server
	// want to specify the file at the base changelist revision
	// build a slice of file names with the changelist id appended
	files_with_cid := append([]string(nil), fileList...)
	for i := 0; i < len(files_with_cid); i++ {
		files_with_cid[i] = files_with_cid[i] + "@" + baseCID
	}
	cmd = exec.CommandContext(ctx, "p4", "sync")
	cmd.Args = append(cmd.Args, files_with_cid...)
	cmd.Dir = tmpClientDir
	cmd.Env = commonEnv
	out, err = cmd.CombinedOutput()
	if err != nil {
		logger.Error("p4 sync failed", log.String("output", string(out)))
		return "", errors.Wrap(err, "gitserver: p4 sync")
	}

	// "checkout" the files by opening them for edit
	cmd = exec.CommandContext(ctx, "p4", "edit")
	cmd.Args = append(cmd.Args, fileList...)
	cmd.Dir = tmpClientDir
	cmd.Env = commonEnv
	cmd.Stdin = bytes.NewReader([]byte(clientSpec))
	out, err = cmd.CombinedOutput()
	if err != nil {
		logger.Error("p4 edit failed", log.String("output", string(out)))
		return "", errors.Wrap(err, "gitserver: p4 edit")
	}

	// delete the files involved with the batch change in case the batch change involves a file deletion
	// ignore all errors for now; just assume that it's going to work
	for _, fileName := range fileList {
		os.RemoveAll(fileName)
	}

	// overlay with files from the commit
	// 1. create an archive from the commit
	// 2. pipe the archive to `tar -x` to extract it into the temp dir

	// archive the commit
	archiveCmd := exec.CommandContext(ctx, "git", "archive", "--format=tar", "--verbose", string(req.BaseCommit))
	archiveCmd.Dir = tmpClientDir
	archiveCmd.Env = commonEnv

	// connect the archive to the untar process
	stdout, err := archiveCmd.StdoutPipe()
	if err != nil {
		return "", errors.Wrap(err, "gitserver: git archive stdout")
	}

	reader := bufio.NewReader(stdout)

	// start the archive; it'll send stdout (the tar archive) to `unpack.Tar` via the `io.Reader`
	if err := archiveCmd.Start(); err != nil {
		return "", errors.Wrap(err, "gitserver: git archive")
	}

	unpack.Tar(reader, tmpClientDir, unpack.Opts{SkipDuplicates: true})

	// make sure the untar process completes before moving on
	if err := archiveCmd.Wait(); err != nil {
		return "", errors.Wrap(err, "gitserver: overlay files with the git archive")
	}

	// ensure that there are changes to shelve

	// use p4 diff to list the changes
	diffCmd := exec.CommandContext(ctx, "p4", "diff", "-f", "-sa")
	diffCmd.Dir = tmpClientDir
	diffCmd.Env = commonEnv
	// use `wc`` to count the files instead of capturing the output of `p4 diff` in Go and counting it
	// because using `wc` saves having to cache the output of `p4 diff` in memory
	wcCmd := exec.CommandContext(ctx, "wc", "-l")
	wcCmd.Dir = tmpClientDir
	wcCmd.Env = commonEnv
	wcCmd.Stdin, _ = diffCmd.StdoutPipe()
	var wcStdout bytes.Buffer
	var wcStderr bytes.Buffer
	wcCmd.Stdout = &wcStdout
	wcCmd.Stderr = &wcStderr
	if err := wcCmd.Start(); err != nil {
		return "", errors.Wrap(err, "gitserver: counting changed files")
	}
	if err := diffCmd.Run(); err != nil {
		return "", errors.Wrap(err, "gitserver: p4 diff")
	}
	if err := wcCmd.Wait(); err != nil {
		return "", errors.Wrap(err, "gitserver: wait for counting changed files")
	}
	if diffFileCount, err := strconv.Atoi(strings.TrimSpace(wcStdout.String())); err != nil {
		return "", errors.Wrap(err, "gitserver: count file diffs")
	} else if diffFileCount <= 0 {
		return "", errors.Wrap(err, "gitserver: no changes to shelve")
	}

	// submit the changes as a shelved changelist

	// create a changelist form
	cmd = exec.CommandContext(ctx, "p4", "change", "-o")
	cmd.Dir = tmpClientDir
	cmd.Env = commonEnv
	out, err = cmd.CombinedOutput()
	if err != nil {
		logger.Error("p4 change failed", log.String("output", string(out)))
		return "", errors.Wrap(err, "gitserver: p4 change")
	}
	// add the commit message to the change form
	changeForm := strings.Replace(string(out), "<enter description here>", desc, 1)

	// feed the changelist form into `p4 shelve`
	// capture the output to parse for a changelist id
	cmd = exec.CommandContext(ctx, "p4", "shelve", "-i")
	cmd.Dir = tmpClientDir
	cmd.Env = commonEnv
	changeBuffer := bytes.Buffer{}
	changeBuffer.Write([]byte(changeForm))
	cmd.Stdin = &changeBuffer
	out, err = cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "gitserver: p4 change")
	}

	matches := cidPattern.FindStringSubmatch(string(out))
	if len(matches) != 2 {
		logger.Error("p4 shelve output does not contain a changelist id", log.String("output", string(out)))
		return "", errors.New("gitserver: p4 shelve output does not contain a changelist id")
	}
	return matches[1], nil
}

func cleanUpTmpRepo(logger log.Logger, path string) {
	err := os.RemoveAll(path)
	if err != nil {
		logger.Warn("unable to clean up tmp repo", log.String("path", path), log.Error(err))
	}
}

// ensureRefPrefix checks whether the ref is a full ref and contains the
// "refs/heads" prefix (i.e. "refs/heads/master") or just an abbreviated ref
// (i.e. "master") and adds the "refs/heads/" prefix if the latter is the case.
//
// Copied from git package to avoid cycle import when testing git package.
func ensureRefPrefix(ref string) string {
	return "refs/heads/" + strings.TrimPrefix(ref, "refs/heads/")
}
