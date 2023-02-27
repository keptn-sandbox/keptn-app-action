package repoaccess

import (
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"net/url"
	"strings"
)

type Client struct {
	githubInstance githubInstance
}

type githubInstance struct {
	owner      string
	repository string
	context    context.Context
	client     *github.Client
}

func NewClient(accessToken string, repository string) (client Client, err error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client.githubInstance.client = github.NewClient(tc)
	client.githubInstance.context = ctx
	if owner, repo, err := getGithubOwnerRepository(repository); err != nil {
		return client, err
	} else {
		client.githubInstance.owner = owner
		client.githubInstance.repository = repo
		return client, nil
	}
}

func getGithubOwnerRepository(raw string) (owner, repository string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return owner, repository, err
	}
	repo := strings.Split(u.Path, "/")
	return repo[0], repo[1], nil
}
