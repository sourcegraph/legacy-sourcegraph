package main

import (
	"context"
	"flag"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// Options are configurables for Combine.
type Options struct {
	Logger *log.Logger
}

func (o *Options) SetDefaults() {
	if o.Logger == nil {
		o.Logger = log.Default()
	}
}

// Combine opens the git repository at path and transforms commits from all
// non-origin remotes into commits onto HEAD.
func Combine(path string, opt Options) error {
	opt.SetDefaults()

	log := opt.Logger

	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	readmeHash, err := readmeObject(r.Storer)
	if err != nil {
		return err
	}

	conf, err := r.Config()
	if err != nil {
		return err
	}

	lastLog := time.Now()

	headRef, _ := r.Head()
	var head *object.Commit
	if headRef != nil {
		head, err = r.CommitObject(headRef.Hash())
		if err != nil {
			return err
		}
	}

	log.Println("Determining the tree hashes of subdirectories...")
	remoteToTree := map[string]plumbing.Hash{}
	if head != nil {
		tree, err := head.Tree()
		if err != nil {
			return err
		}
		for _, entry := range tree.Entries {
			remoteToTree[entry.Name] = entry.Hash
		}
	}

	log.Println("Collecting new commits...")
	lastLog = time.Now()
	remoteToCommits := map[string][]*object.Commit{}
	for remote := range conf.Remotes {
		if remote == "origin" {
			continue
		}

		commit, err := remoteHead(r, remote)
		if err != nil {
			return err
		}
		if commit == nil {
			// No known default branch on this remote, ignore it.
			continue
		}

		for depth := 0; ; depth++ {
			if time.Since(lastLog) > time.Second {
				log.Printf("Collecting new commits... (remotes %s, commit depth %d)", remote, depth)
				lastLog = time.Now()
			}

			if commit.TreeHash == remoteToTree[remote] {
				break
			}

			remoteToCommits[remote] = append(remoteToCommits[remote], commit)

			if commit.NumParents() == 0 {
				remoteToTree[remote] = commit.TreeHash
				break
			}
			commit, err = commit.Parent(0)
			if err != nil {
				return err
			}
		}
	}

	applyCommit := func(remote string, commit *object.Commit) error {
		remoteToTree[remote] = commit.TreeHash

		// Add tree entries for each remote in matching directories.
		var entries []object.TreeEntry
		for thisRemote, tree := range remoteToTree {
			entries = append(entries, object.TreeEntry{
				Name: thisRemote,
				Mode: filemode.Dir,
				Hash: tree,
			})
		}

		// Add README.md.
		entries = append(entries, object.TreeEntry{
			Name: "README.md",
			Mode: filemode.Regular,
			Hash: readmeHash,
		})

		// TODO is this necessary?
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Name < entries[j].Name
		})

		// Construct the root tree.
		treeHash, err := storeObject(r.Storer, &object.Tree{
			Entries: entries,
		})
		if err != nil {
			return err
		}

		// TODO break links so we don't appear in upstream analytics. IE
		// remove links from message, scrub author and committer, etc.
		newCommit := &object.Commit{
			Author: sanitizeSignature(commit.Author),
			Committer: object.Signature{
				Name:  "sourcegraph-bot",
				Email: "no-reply@sourcegraph.com",
				When:  commit.Committer.When,
			},
			Message:  sanitizeMessage(remote, commit),
			TreeHash: treeHash,
		}

		// We just create a linear history. parentHash is zero if this is the
		// first commit to HEAD.
		if head != nil {
			newCommit.ParentHashes = []plumbing.Hash{head.Hash}
		}

		headHash, err := storeObject(r.Storer, newCommit)
		if err != nil {
			return err
		}

		if err := setHEAD(r.Storer, headHash); err != nil {
			return err
		}

		headRef, _ := r.Head()
		if headRef != nil {
			head, err = r.CommitObject(headRef.Hash())
			if err != nil {
				return err
			}
		}

		return nil
	}

	log.Println("Applying new commits...")
	total := 0
	for _, commits := range remoteToCommits {
		total += len(commits)
	}
	for height := 0; len(remoteToCommits) > 0; {
		// Loop over keys so we can delete entries from the map.
		remotes := []string{}
		for remote := range remoteToCommits {
			remotes = append(remotes, remote)
		}

		// Pop 1 commit per remote and put each tree in a directory by the same name as the remote.
		for _, remote := range remotes {
			deepestCommit := remoteToCommits[remote][len(remoteToCommits[remote])-1]

			err = applyCommit(remote, deepestCommit)
			if err != nil {
				return err
			}
			height++

			// Pop the deepest commit.
			remoteToCommits[remote] = remoteToCommits[remote][:len(remoteToCommits[remote])-1]

			// Delete this remote once we applied all of its new commits.
			if len(remoteToCommits[remote]) == 0 {
				delete(remoteToCommits, remote)
			}

			if time.Since(lastLog) > time.Second {
				progress := float64(height) / float64(total)
				log.Printf("%.2f%% done (applied %d commits out of %d total)", progress*100, height+1, total)
				lastLog = time.Now()
			}
		}
	}

	return nil
}

func readmeObject(storer storer.EncodedObjectStorer) (plumbing.Hash, error) {
	readme := []byte(`# megarepo

This is a synthetic monorepo created by continuously applying commits from upstream projects into respective sub directories.

See https://github.com/sourcegraph/sourcegraph/tree/main/internal/cmd/git-combine
`)
	obj := storer.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	obj.SetSize(int64(len(readme)))

	w, err := obj.Writer()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err := w.Write(readme); err != nil {
		_ = w.Close()
		return plumbing.ZeroHash, err
	}

	if err := w.Close(); err != nil {
		return plumbing.ZeroHash, err
	}

	return storer.SetEncodedObject(obj)
}

func storeObject(storer storer.EncodedObjectStorer, obj interface {
	Encode(plumbing.EncodedObject) error
}) (plumbing.Hash, error) {
	o := storer.NewEncodedObject()
	if err := obj.Encode(o); err != nil {
		return plumbing.ZeroHash, err
	}

	hash := o.Hash()
	if storer.HasEncodedObject(hash) == nil {
		return hash, nil
	}

	if _, err := storer.SetEncodedObject(o); err != nil {
		return plumbing.ZeroHash, err
	}

	return hash, nil
}

func setHEAD(storer storer.ReferenceStorer, hash plumbing.Hash) error {
	head, err := storer.Reference(plumbing.HEAD)
	if err != nil {
		return err
	}

	name := plumbing.HEAD
	if head.Type() != plumbing.HashReference {
		name = head.Target()
	}

	return storer.SetReference(plumbing.NewHashReference(name, hash))
}

func sanitizeSignature(sig object.Signature) object.Signature {
	// We sanitize the email since that is how github connects up commits to
	// authors. We intentionally break this connection since these are
	// synthetic commits.
	prefix := "no-reply"
	if idx := strings.Index(sig.Email, "@"); idx > 0 {
		prefix = sig.Email[:idx]
	}
	email := fmt.Sprintf("%s@%X.example.com", prefix, crc32.ChecksumIEEE([]byte(sig.Email)))

	return object.Signature{
		Name:  sig.Name,
		Email: email,
		When:  sig.When,
	}
}

func sanitizeMessage(dir string, commit *object.Commit) string {
	// There are lots of things that could link to other artificats in the
	// commit message. So we play it safe and just remove the message.
	title := commitTitle(commit)

	// vscode seems to often include URLs to issues and ping users in commit
	// titles. I am guessing this is due to its tiny box for creating commit
	// messages. This leads to github crosslinking to megarepo. Lets naively
	// sanitize.
	for _, bad := range []string{"@", "http://", "https://"} {
		if i := strings.Index(title, bad); i >= 0 {
			title = title[:i]
		}
	}

	title = strings.TrimSpace(title)

	return fmt.Sprintf("%s: %s\n\nCommit: %s\n", dir, title, commit.Hash)
}

func commitTitle(commit *object.Commit) string {
	title := commit.Message
	if idx := strings.IndexByte(title, '\n'); idx > 0 {
		title = title[:idx]
	}
	return strings.TrimSpace(title)
}

func hasRemote(path, remote string) (bool, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return false, err
	}

	conf, err := r.Config()
	if err != nil {
		return false, err
	}

	_, ok := conf.Remotes[remote]
	return ok, nil
}

func getGitDir() (string, error) {
	dir := os.Getenv("GIT_DIR")
	if dir == "" {
		return os.Getwd()
	}
	return dir, nil
}

func runGit(dir string, args ...string) error {
	// they should be _much_ faster than this, but we set this incase git gets blocked.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	log.Printf("starting git %s", strings.Join(args, " "))
	err := cmd.Run()
	log.Printf("finished git in %s", time.Since(start))
	return err
}

func doDaemon(dir string, done <-chan struct{}, opt Options) error {
	isDone := func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}

	opt.SetDefaults()

	for {
		// convenient way to stop the daemon to do manual operations like add
		// more upstreams.
		if b, err := os.ReadFile(filepath.Join(dir, "PAUSE")); err == nil {
			opt.Logger.Printf("PAUSE file present: %s", string(b))
			select {
			case <-time.After(time.Minute):
			case <-done:
				return nil
			}
			continue
		}

		if err := runGit(dir, "fetch", "--all", "--no-tags"); err != nil {
			return err
		}

		if isDone() {
			return nil
		}

		if err := Combine(dir, opt); err != nil {
			return err
		}

		if isDone() {
			return nil
		}

		if hasOrigin, err := hasRemote(dir, "origin"); err != nil {
			return err
		} else if !hasOrigin {
			opt.Logger.Printf("skipping push since remote origin is missing")
		} else if err := runGit(dir, "push", "origin"); err != nil {
			return err
		}

		select {
		case <-time.After(time.Minute):
		case <-done:
			return nil
		}
	}
}

func main() {
	daemon := flag.Bool("daemon", false, "run in daemon mode. This mode loops on fetch, combine, push.")

	flag.Parse()

	opt := Options{}

	gitDir, err := getGitDir()
	if err != nil {
		log.Fatal(err)
	}

	if *daemon {
		done := make(chan struct{}, 1)

		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c
			done <- struct{}{}
		}()

		err := doDaemon(gitDir, done, opt)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err = Combine(gitDir, opt)
	if err != nil {
		log.Fatal(err)
	}
}

// remoteHead returns the HEAD commit for the given remote.
func remoteHead(r *git.Repository, remote string) (*object.Commit, error) {
	// We don't know what the remote HEAD is, so we hardcode the usual options and test if they exist.
	commonDefaultBranches := []string{"main", "master", "trunk", "development"}
	for _, name := range commonDefaultBranches {
		ref, err := storer.ResolveReference(r.Storer, plumbing.NewRemoteReferenceName(remote, name))
		if err == nil {
			return r.CommitObject(ref.Hash())
		}
	}

	log.Printf("ignoring remote %q because it doesn't have any of the common default branches %v", remote, commonDefaultBranches)
	return nil, nil
}
