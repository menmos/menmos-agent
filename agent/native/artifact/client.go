package artifact

import (
	"context"

	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
)

func getGithubClient(token string) *github.Client {
	if token == "" {
		return github.NewClient(nil)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}
