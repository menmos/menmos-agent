package artifact

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/google/go-github/v43/github"
	"go.uber.org/zap"
)

type Repository struct {
	githubClient *github.Client
	log          *zap.SugaredLogger
	path         string
}

func NewRepository(path string, githubToken string, log *zap.Logger) *Repository {
	return &Repository{
		githubClient: getGithubClient(githubToken),
		log:          log.Sugar().Named("artifacts"),
		path:         path,
	}
}

func (r *Repository) doesVersionDirectoryExist(path string) (bool, error) {
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !dirInfo.IsDir() {
		return false, fmt.Errorf("path '%v' is not a directory", path)
	}

	return true, nil
}

func (r *Repository) doesArtifactExist(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, nil
	}
	if fileInfo.IsDir() {
		return false, fmt.Errorf("artifact path '%v' is a directory", path)
	}

	return true, nil
}

func (r *Repository) downloadAsset(downloadURL, fileName, versionDirectory string) error {

	if downloadURL == "" || fileName == "" {
		r.log.Debugf("skipped asset, missingfilename or url")
		return nil
	}

	r.log.Debugf("downloading asset '%s'", fileName)

	assetPath := path.Join(versionDirectory, (&asset{FullName: fileName}).Name())
	assetFile, err := os.OpenFile(assetPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(assetFile, resp.Body)
	if err != nil {
		return err
	}

	r.log.Infof("downloaded asset '%v'", fileName)
	return nil
}

func (r *Repository) getPlatformAssets(assets []*github.ReleaseAsset) []*github.ReleaseAsset {
	validArchitectures := []string{runtime.GOARCH}
	if runtime.GOOS == "darwin" {
		// Hack to support M1 macs, they _can_ use amd64 via rosetta
		// but we prefer arm64 if available.
		validArchitectures = append(validArchitectures, "amd64")
	}

	for _, arch := range validArchitectures {
		var archAssets []*github.ReleaseAsset

		for _, ghAsset := range assets {
			currentAsset := asset{FullName: ghAsset.GetName()}
			if architectureEqual(arch, currentAsset.Architecture()) && platformEqual(runtime.GOOS, currentAsset.Platform()) {
				archAssets = append(archAssets, ghAsset)
			}
		}

		if len(archAssets) > 0 {
			r.log.Debugf("got %d platform assets", len(archAssets))
			return archAssets
		}
	}

	return []*github.ReleaseAsset{}
}

func (r *Repository) downloadRelease(version string) error {
	release, _, err := r.githubClient.Repositories.GetReleaseByTag(context.Background(), "menmos", "menmos", version)
	if err != nil {
		return err

	}
	r.log.Debugf("got github release for '%s'", version)

	versionDirectory := path.Join(r.path, version)
	if err := os.MkdirAll(versionDirectory, 0755); err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, asset := range r.getPlatformAssets(release.Assets) {
		wg.Add(1)
		go func(asset *github.ReleaseAsset) {
			if err := r.downloadAsset(asset.GetBrowserDownloadURL(), asset.GetName(), versionDirectory); err != nil {
				r.log.Errorf("failed to download asset '%s': %v", asset.GetName(), err.Error())
			}
			wg.Done()
		}(asset)
	}

	wg.Wait()

	return nil
}

func (r *Repository) Get(version, name string) (string, error) {
	versionDir := path.Join(r.path, version)
	exists, err := r.doesVersionDirectoryExist(versionDir)
	if err != nil {
		return "", err
	}

	if !exists {
		if err := r.downloadRelease(version); err != nil {
			return "", err
		}
	}

	// Check if artifact is there.
	// If not -> error
	artifactPath := path.Join(versionDir, name)
	exists, err = r.doesArtifactExist(artifactPath)
	if err != nil {
		return "", err
	}

	if exists {
		return artifactPath, nil
	} else {
		return "", fmt.Errorf("artifact '%s' does not exist for version '%s'", name, version)
	}
}
