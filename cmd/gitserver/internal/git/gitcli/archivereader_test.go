package gitcli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/require"
)

func readFileContentsFromTar(t *testing.T, tr *tar.Reader, name string) string {
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		if h.Name == name {
			contents, err := io.ReadAll(tr)
			require.NoError(t, err)
			return string(contents)
		}
	}

	t.Fatalf("File %q not found in tar archive", name)
	return ""
}

func readFileContentsFromZip(t *testing.T, zr *zip.Reader, name string) string {
	f, err := zr.Open(name)
	if err != nil {
		t.Fatalf("File %q not found in zip archive", name)
	}

	contents, err := io.ReadAll(f)
	require.NoError(t, err)
	return string(contents)
}

func TestGitCLIBackend_buildArchiveArgs(t *testing.T) {
	t.Run("no pathspecs", func(t *testing.T) {
		args := buildArchiveArgs(gitserver.ArchiveFormatTar, "HEAD", nil)
		require.Equal(t, []string{"archive", "--worktree-attributes", "--format=tar", "HEAD", "--"}, args)
	})

	t.Run("with pathspecs", func(t *testing.T) {
		args := buildArchiveArgs(gitserver.ArchiveFormatTar, "HEAD", []string{"file1", "file2"})
		require.Equal(t, []string{"archive", "--worktree-attributes", "--format=tar", "HEAD", "--", "file1", "file2"}, args)
	})

	t.Run("zip adds -0", func(t *testing.T) {
		args := buildArchiveArgs(gitserver.ArchiveFormatZip, "HEAD", nil)
		require.Equal(t, []string{"archive", "--worktree-attributes", "--format=zip", "-0", "HEAD", "--"}, args)
	})
}

func TestGitCLIBackend_ArchiveReader(t *testing.T) {
	ctx := context.Background()

	backend := BackendWithRepoCommands(t,
		"echo abcd > file1",
		"mkdir dir1",
		"echo efgh > dir1/file2",
		"git add file1",
		"git add dir1",
		"git commit -m commit --author='Foo Author <foo@sourcegraph.com>'",
	)

	commitID, err := backend.RevParseHead(ctx)
	require.NoError(t, err)

	t.Run("read simple tar archive", func(t *testing.T) {
		r, err := backend.ArchiveReader(ctx, "tar", string(commitID), nil)
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		tr := tar.NewReader(r)
		contents := readFileContentsFromTar(t, tr, "file1")
		require.Equal(t, "abcd\n", contents)
	})

	t.Run("read simple zip archive", func(t *testing.T) {
		r, err := backend.ArchiveReader(ctx, "zip", string(commitID), nil)
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		contents, err := io.ReadAll(r)
		require.NoError(t, err)
		zr, err := zip.NewReader(bytes.NewReader([]byte(contents)), int64(len(contents)))
		require.NoError(t, err)
		fileContents := readFileContentsFromZip(t, zr, "file1")
		require.Equal(t, "abcd\n", fileContents)
	})

	t.Run("read multiple files from tar archive using PathspecLiteral", func(t *testing.T) {
		r, err := backend.ArchiveReader(ctx, "tar", string(commitID), []string{string(gitdomain.PathspecLiteral("file1")), string(gitdomain.PathspecLiteral("dir1/file2"))})
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		tr := tar.NewReader(r)
		contents := readFileContentsFromTar(t, tr, "dir1/file2")
		require.Equal(t, "efgh\n", contents)
	})

	t.Run("read file in directory", func(t *testing.T) {
		r, err := backend.ArchiveReader(ctx, "tar", string(commitID), nil)
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		tr := tar.NewReader(r)
		contents := readFileContentsFromTar(t, tr, "dir1/file2")
		require.Equal(t, "efgh\n", contents)
	})

	t.Run("non existent commit", func(t *testing.T) {
		_, err := backend.ArchiveReader(ctx, "tar", "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", nil)
		require.Error(t, err)
		require.True(t, errors.HasType(err, &gitdomain.RevisionNotFoundError{}))
	})

	t.Run("non existent file", func(t *testing.T) {
		_, err := backend.ArchiveReader(ctx, "tar", string(commitID), []string{"no-file"})
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})
}
