package fabric

import (
	"context"
	"log"
	"strings"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1beta2"
	"google.golang.org/api/option"
	artifactregistrypb "google.golang.org/genproto/googleapis/devtools/artifactregistry/v1beta2"
)

type ArtifactoryService struct {
	client *artifactregistry.Client
	config *Config
}

func NewArtifactoryService(config *Config) *ArtifactoryService {
	client, err := artifactregistry.NewClient(context.Background(), option.WithCredentialsFile(config.Organization.Credsfile))
	if err != nil {
		log.Fatal(err)
	}

	return &ArtifactoryService{
		client,
		config,
	}
}

func (as *ArtifactoryService) Close() error {
	return as.client.Close()
}

func (a *ArtifactoryService) CreateArtifactory(ctx context.Context) error {
	req := &artifactregistrypb.CreateRepositoryRequest{
		// See https://pkg.go.dev/google.golang.org/genproto/googleapis/devtools/artifactregistry/v1beta2#CreateRepositoryRequest.
		Parent:       a.config.RepoBasePath(),
		RepositoryId: a.config.Organization.Name,
		Repository: &artifactregistrypb.Repository{
			Name:        a.config.RepoFullPath(),
			Description: "The place we store things",
			Format:      artifactregistrypb.Repository_DOCKER,
		},
	}

	op, err := a.client.CreateRepository(ctx, req)
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
