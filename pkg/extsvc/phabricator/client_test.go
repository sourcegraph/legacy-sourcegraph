package phabricator_test

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/phabricator"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/httptestutil"
)

var update = flag.Bool("update", false, "update testdata")

func TestClient_GetRawDiff(t *testing.T) {
	cli, save := newClient(t, "GetRawDiff")
	defer save()

	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	for _, tc := range []struct {
		name string
		ctx  context.Context
		id   int
		err  string
	}{{
		name: "diff not found",
		id:   0xdeadbeef,
		err:  "ERR_NOT_FOUND: Diff not found.",
	}, {
		name: "diff found",
		id:   20455,
	}, {
		name: "timeout",
		ctx:  timeout,
		err:  "Post https://secure.phabricator.com/api/differential.getrawdiff: context deadline exceeded",
	}} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Background()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			diff, err := cli.GetRawDiff(tc.ctx, tc.id)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.id == 0 {
				return
			}

			path := "testdata/golden/get-raw-diff-" + strconv.Itoa(tc.id)
			if *update {
				if err = ioutil.WriteFile(path, []byte(diff), 0640); err != nil {
					t.Fatalf("failed to update golden file %q: %s", path, err)
				}
			}

			golden, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read golden file %q: %s", path, err)
			}

			if have, want := diff, string(golden); have != want {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(have, want, false)
				t.Error(dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func newClient(t testing.TB, name string) (*phabricator.Client, func()) {
	t.Helper()

	cassete := filepath.Join("testdata/vcr/", strings.Replace(name, " ", "-", -1))
	rec, err := httptestutil.NewRecorder(cassete, *update, func(i *cassette.Interaction) error {
		// Remove all tokens
		i.Request.Body = ""
		i.Request.Form = map[string][]string{}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	hc, err := httpcli.NewFactory(nil, httptestutil.NewRecorderOpt(rec)).NewClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	cli, err := phabricator.NewClient(
		ctx,
		"https://secure.phabricator.com",
		os.Getenv("PHABRICATOR_TOKEN"),
		hc,
	)

	if err != nil {
		t.Fatal(err)
	}

	return cli, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	}
}
