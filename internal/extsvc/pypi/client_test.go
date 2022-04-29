package pypi

import (
	"context"
	"flag"
	"path"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}

func TestDownload(t *testing.T) {
	ctx := context.Background()
	cli := newTestClient(t, "Download", update(t.Name()))

	b, err := cli.Project(ctx, "requests")
	if err != nil {
		t.Fatal(err)
	}

	fs, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}

	// Pick the oldest tarball.
	j := -1
	for i, f := range fs {
		if path.Ext(f.Name) == ".gz" {
			j = i
			break
		}
	}

	p, err := cli.Download(ctx, fs[j].URL)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/requests", update(t.Name()), p)
}

func TestParse(t *testing.T) {
	cli := newTestClient(t, "Parse", update(t.Name()))
	b, err := cli.Project(context.Background(), "tiny")
	if err != nil {
		t.Fatal(err)
	}
	files, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, "testdata/golden/tiny", update(t.Name()), files)
}

func TestParse_empty(t *testing.T) {
	b := []byte(`
<!DOCTYPE html>
<html>
  <body>
  </body>
</html>
`)

	_, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParse_PEP503(t *testing.T) {
	// There may be any other HTML elements on the API pages as long as the required
	// anchor elements exist.
	b := []byte(`
<!DOCTYPE html>
<html>
  <head>
    <meta name="pypi:repository-version" content="1.0">
    <title>Links for frob</title>
  </head>
  <body>
	<h1>Links for frob</h1>
    <a href="/frob-1.0.0.tar.gz/" data-requires-python="&gt;=3">frob-1.0.0</a>
	<h2>More links for frob</h1>
	<div>
	    <a href="/frob-2.0.0.tar.gz/" data-gpg-sig="true">frob-2.0.0</a>
	</div>
  </body>
</html>
`)

	got, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}

	tr := true
	want := []File{
		{
			Name:               "frob-1.0.0",
			URL:                "/frob-1.0.0.tar.gz/",
			DataRequiresPython: ">=3",
		},
		{
			Name:       "frob-2.0.0",
			URL:        "/frob-2.0.0.tar.gz/",
			DataGPGSig: &tr,
		},
	}

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("-want, +got\n%s", d)
	}
}

// goos: darwin
// goarch: arm64
// pkg: github.com/sourcegraph/sourcegraph/internal/extsvc/pypi
// BenchmarkParse-10           5781            207931 ns/op
func BenchmarkParse(b *testing.B) {
	cli := newTestClient(b, "Download", update("TestDownload"))
	data, err := cli.Project(context.Background(), "requests")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestToWheel(t *testing.T) {
	have := []string{
		"requests-2.16.2-py2.py3-none-any.whl",
		"grpcio-1.46.0rc2-cp39-cp39-win_amd64.whl",
	}
	want := []Wheel{
		{
			Distribution: "requests",
			Version:      "2.16.2",
			BuildTag:     "",
			PythonTag:    "py2.py3",
			ABITag:       "none",
			PlatformTag:  "any",
		},
		{
			Distribution: "grpcio",
			Version:      "1.46.0rc2",
			BuildTag:     "",
			PythonTag:    "cp39",
			ABITag:       "cp39",
			PlatformTag:  "win_amd64",
		},
	}

	var got []Wheel
	for _, h := range have {
		g, err := ToWheel(h)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, *g)
	}

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("-want, +got\n%s", d)
	}
}

// newTestClient returns a pypi Client that records its interactions
// to testdata/vcr/.
func newTestClient(t testing.TB, name string, update bool) *Client {
	cassete := filepath.Join("testdata/vcr/", normalize(name))
	rec, err := httptestutil.NewRecorder(cassete, update)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	})

	doer, err := httpcli.NewFactory(nil, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fatal(err)
	}

	c := NewClient("urn", []string{"https://pypi.org/simple"})
	c.cli = doer
	return c
}
