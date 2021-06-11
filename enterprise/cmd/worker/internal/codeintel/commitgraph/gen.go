package commitgraph

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/derision-test/go-mockgen/cmd/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/commitgraph -i DBStore -i Locker -i GitserverClient -o mock_iface_test.go
