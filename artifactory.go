package fabric

import (
	"context"
	"log"
	"strings"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1beta2"
	"google.golang.org/api/option"
	artifactregistrypb "google.golang.org/genproto/googleapis/devtools/artifactregistry/v1beta2"
)

func CreateArtifactory(ctx context.Context, config *Config) error {
	client, err := artifactregistry.NewClient(ctx, option.WithCredentialsFile(config.Organization.Credsfile))
	if err != nil {
		return err
	}
	defer client.Close()

	req := &artifactregistrypb.CreateRepositoryRequest{
		// See https://pkg.go.dev/google.golang.org/genproto/googleapis/devtools/artifactregistry/v1beta2#CreateRepositoryRequest.
		Parent:       config.RepoBasePath(),
		RepositoryId: config.Organization.Name,
		Repository: &artifactregistrypb.Repository{
			Name:        config.RepoFullPath(),
			Description: "The place we store things",
			Format:      artifactregistrypb.Repository_DOCKER,
		},
	}

	op, err := client.CreateRepository(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "code = AlreadyExists") {
			log.Println(err.Error())
			return nil
		} else {
			return err
		}
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}
