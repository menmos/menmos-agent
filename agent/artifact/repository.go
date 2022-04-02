package artifact

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"go.uber.org/zap"
)

type RepositoryParams struct {
	ReleaseFetcher MenmosReleaseFetcher
	Log            *zap.Logger
	Path           string
}

type Repository struct {
	releaseFetcher MenmosReleaseFetcher
	log            *zap.SugaredLogger
	path           string
}

func NewRepository(params RepositoryParams) *Repository {
	return &Repository{
		releaseFetcher: params.ReleaseFetcher,
		log:            params.Log.Sugar().Named("artifacts"),
		path:           params.Path,
	}
}

func (r *Repository) doesDirectoryExist(path string) (bool, error) {
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

func (r *Repository) downloadAsset(tgtAsset *Asset, versionDirectory string) error {

	if tgtAsset.DownloadURL == "" || tgtAsset.FullName == "" {
		r.log.Debugf("skipped asset, missingfilename or url")
		return nil
	}

	r.log.Debugf("downloading asset '%s'", tgtAsset.FullName)

	assetPath := filepath.Join(versionDirectory, tgtAsset.Name())
	assetFile, err := os.OpenFile(assetPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	resp, err := http.Get(tgtAsset.DownloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(assetFile, resp.Body)
	if err != nil {
		return err
	}

	r.log.Infof("downloaded asset '%v'", tgtAsset.FullName)
	return nil
}

func (r *Repository) getPlatformAssets(assets []*Asset) []*Asset {
	validArchitectures := []string{runtime.GOARCH}
	if runtime.GOOS == "darwin" {
		// Hack to support M1 macs, they _can_ use amd64 via rosetta
		// but we prefer arm64 if available.
		validArchitectures = append(validArchitectures, "amd64")
	}

	for _, arch := range validArchitectures {
		var archAssets []*Asset

		for _, currentAsset := range assets {
			if architectureEqual(arch, currentAsset.Architecture()) && platformEqual(runtime.GOOS, currentAsset.Platform()) {
				archAssets = append(archAssets, currentAsset)
			}
		}

		if len(archAssets) > 0 {
			r.log.Debugf("got %d platform assets", len(archAssets))
			return archAssets
		}
	}

	return []*Asset{}
}

func (r *Repository) downloadRelease(version string) error {
	assets, err := r.releaseFetcher.GetRelease(context.Background(), version)
	if err != nil {
		return err

	}
	r.log.Debugf("got github release for '%s'", version)

	versionDirectory := filepath.Join(r.path, version)
	if err := os.MkdirAll(versionDirectory, 0755); err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, currentAsset := range r.getPlatformAssets(assets) {
		wg.Add(1)
		go func(currentAsset *Asset) {
			if err := r.downloadAsset(currentAsset, versionDirectory); err != nil {
				r.log.Errorf("failed to download asset '%s': %v", currentAsset.Name(), err.Error())
			}
			wg.Done()
		}(currentAsset)
	}

	wg.Wait()

	return nil
}

func (r *Repository) Get(version, name string) (string, error) {
	versionDir := filepath.Join(r.path, version)
	exists, err := r.doesDirectoryExist(versionDir)
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
	artifactPath := filepath.Join(versionDir, name)
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
