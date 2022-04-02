package artifact

import (
	"context"

	"github.com/google/go-github/v43/github"
)

// A MenmosReleaseFetcher fetches a given release from the menmos repository.
type MenmosReleaseFetcher interface {
	GetRelease(context context.Context, versionTag string) ([]*Asset, error)
}

type GithubReleaseFetcher struct {
	client *github.Client
}

func NewGithubFetcher(token string) *GithubReleaseFetcher {
	return &GithubReleaseFetcher{
		client: getGithubClient(token),
	}
}

func (g *GithubReleaseFetcher) GetRelease(ctx context.Context, versionTag string) ([]*Asset, error) {
	release, _, err := g.client.Repositories.GetReleaseByTag(ctx, "menmos", "menmos", versionTag)
	if err != nil {
		return nil, err
	}

	assets := make([]*Asset, len(release.Assets))
	for i := 0; i < len(release.Assets); i++ {
		assets[i] = &Asset{FullName: release.Assets[i].GetName(), DownloadURL: release.Assets[i].GetBrowserDownloadURL()}
	}

	return assets, nil
}
