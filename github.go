package covid19brazilimporter

import (
	"context"

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

func getFileContent(ctx context.Context, client *github.Client,
	properties *properties) (*github.RepositoryContent, error) {
	fileContent, _, _, err := client.Repositories.GetContents(ctx, properties.RepoOwner, properties.Repo, properties.Filename, &github.RepositoryContentGetOptions{
		Ref: properties.Branch,
	})
	return fileContent, err
}

func updateFile(ctx context.Context, client *github.Client, properties *properties, content []byte) (*github.RepositoryContentResponse, error) {

	fileConetent, err := getFileContent(ctx, client, properties)
	if err != nil {
		return nil, err
	}

	contentResponse, _, err := client.Repositories.UpdateFile(ctx, properties.RepoOwner,
		properties.Repo, properties.Filename, &github.RepositoryContentFileOptions{
			SHA:     fileConetent.SHA,
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
