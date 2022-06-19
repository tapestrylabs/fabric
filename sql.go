package fabric

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	sql "google.golang.org/api/sql/v1beta4"
)

type SqlService struct {
	client *sql.Service
	config *Config
}

func NewSqlService(config *Config) *SqlService {
	client, err := sql.NewService(context.Background(), option.WithCredentialsFile(config.Organization.Credsfile))
	if err != nil {
		log.Fatal(err)
	}

	return &SqlService{
		client,
		config,
	}
}

func (s *SqlService) CreateSqlInstance(ctx context.Context, service *Service) error {
	project := s.config.Creds.ProjectId
	databaseInstance := *&sql.DatabaseInstance{
		DatabaseVersion: "POSTGRES_12",
		Name:            service.Database.Name,
		Project:         project,
		Region:          service.Region,
		Settings: &sql.Settings{
			Tier: s.config.GetSqlInstanceTier(service),
		},
		RootPassword: "password",
	}

	op, err := s.client.Instances.Insert(project, &databaseInstance).Do()
	if err != nil {
		return err
	}

	if op.OperationType == "CREATE" {
		fmt.Println("CREATED")
	}

	return nil
}

func (s *SqlService) ListSqlInstances(ctx context.Context) ([]*sql.DatabaseInstance, error) {
	project := s.config.Creds.ProjectId
	instances, err := s.client.Instances.List(project).Do()
	if err != nil {
		return nil, err
	}

	return instances.Items, nil
}
