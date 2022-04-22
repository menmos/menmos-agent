package artifact_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"github.com/menmos/menmos-agent/agent/native/artifact"
	"go.uber.org/zap"
)

type mockReleaseFetcher struct {
	Releases map[string][]*artifact.Asset
	ts       *httptest.Server
}

func newMockFetcher(versions ...string) *mockReleaseFetcher {
	fetcher := mockReleaseFetcher{
		Releases: make(map[string][]*artifact.Asset),
	}

	fetcher.ts = httptest.NewServer(http.HandlerFunc(fetcher.serveAsset))

	for _, version := range versions {
		var assets []*artifact.Asset
		for _, plat := range []string{"linux", "darwin", "windows"} {
			for _, arch := range []string{"amd64", "arm", "arm64"} {
				fullName := fmt.Sprintf("myapp-%s-%s", plat, arch)
				assets = append(assets, &artifact.Asset{FullName: fullName, DownloadURL: fetcher.ts.URL + fmt.Sprintf("/%s/%s", version, fullName)})
			}

		}

		fetcher.Releases[version] = assets
	}

	return &fetcher
}

func (f *mockReleaseFetcher) serveAsset(w http.ResponseWriter, r *http.Request) {
	// We write the URL path directly in the response. This allows the test
	// to see what artifact was requested.
	w.Write([]byte(r.URL.Path))
	w.WriteHeader(http.StatusOK)
}

func (f *mockReleaseFetcher) GetRelease(ctx context.Context, version string) ([]*artifact.Asset, error) {
	if assets, ok := f.Releases[version]; ok {
		return assets, nil
	} else {
		return nil, errors.New("release not found")
	}
}

func (f *mockReleaseFetcher) Close() {
	f.ts.Close()
}

func TestRepository_Get(t *testing.T) {

	fetcher := newMockFetcher("v1.0.0", "v2.0.0")
	defer fetcher.Close()

	type args struct {
		version string
		name    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"simpleGet", args{"v1.0.0", "myapp"}, false},
		{"nonExistentArtifact", args{"v1.0.0", "something"}, true},
		{"nonExistentVersion", args{"v18.0.0", "myapp"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dir, err := ioutil.TempDir("", "TestRepository_Get")
			if err != nil {
				t.Fatal(err)
			}

			defer os.RemoveAll(dir) // clean up

			params := artifact.RepositoryParams{
				ReleaseFetcher: fetcher,
				Log:            zap.NewNop(),
				Path:           dir,
			}

			r := artifact.NewRepository(params)
			got, err := r.Get(tt.args.version, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Go read the file and make sure that the correct path is written in it
			contents, err := ioutil.ReadFile(got)
			if err != nil {
				t.Errorf("failed to read artifact: %v", err)
				return
			}

			expectedArtifact := fmt.Sprintf("/%s/%s-%s-%s", tt.args.version, tt.args.name, runtime.GOOS, runtime.GOARCH)
			if string(contents) != expectedArtifact {
				t.Errorf("artifact contents expected = '%v', actual = '%v'", expectedArtifact, string(contents))
				return
			}
		})
	}
}
