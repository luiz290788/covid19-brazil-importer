package covid19brazilimporter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func buildGithubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func getFileSha(ctx context.Context, client *github.Client, properties *properties) (*string, error) {

	treeSha := fmt.Sprintf("%v^{tree}:src/data", properties.Branch)

	tree, _, err := client.Git.GetTree(context.Background(), properties.RepoOwner, properties.Repo, treeSha, false)
	if err != nil {
		return nil, err
	}
	for _, entry := range tree.Entries {
		if strings.HasSuffix(properties.Filename, entry.GetPath()) {
			return entry.SHA, nil
		}
	}
	return nil, errors.New("file not found")
}

func updateFile(ctx context.Context, client *github.Client, properties *properties,
	content []byte) (*github.RepositoryContentResponse, error) {

	sha, err := getFileSha(ctx, client, properties)
	if err != nil {
		return nil, err
	}

	contentResponse, _, err := client.Repositories.UpdateFile(ctx, properties.RepoOwner,
		properties.Repo, properties.Filename, &github.RepositoryContentFileOptions{
			SHA:     sha,
			Branch:  &properties.Branch,
			Content: content,
			Message: &properties.CommitMessage,
			Committer: &github.CommitAuthor{
				Name:  &properties.CommiterName,
				Email: &properties.CommiterEmail,
			},
		})
	return contentResponse, err
}
