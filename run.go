package fabric

import (
	"context"
	"log"
	"strings"

	run "cloud.google.com/go/run/apiv2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	runpb "google.golang.org/genproto/googleapis/cloud/run/v2"
	"google.golang.org/genproto/googleapis/iam/v1"
)

type RunService struct {
	client *run.ServicesClient
	config *Config
}

func NewRunService(config *Config) *RunService {
	client, err := run.NewServicesClient(context.Background(), option.WithCredentialsFile(config.Organization.Credsfile))
	if err != nil {
		log.Fatal(err)
	}

	return &RunService{
		client,
		config,
	}
}

func (run *RunService) Close() error {
	return run.client.Close()
}

func (run *RunService) UpsertRunService(ctx context.Context, service *Service) error {
	req := &runpb.UpdateServiceRequest{
		Service: &runpb.Service{
			Name: run.config.ServicePath(service),
			Template: &runpb.RevisionTemplate{
				Revision: service.GenRevisionString(),
				Containers: []*runpb.Container{
					{
						Image: run.config.ImageFullPath(service),
					},
				},
			},
		},
		AllowMissing: true,
	}

	op, err := run.client.UpdateService(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "code = NotFound") {
			err = run.CreateRunService(ctx, service)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		_, err = op.Wait(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (run *RunService) CreateRunService(ctx context.Context, service *Service) error {
	req := &runpb.CreateServiceRequest{
		Parent: run.config.GetLocation(service),
		Service: &runpb.Service{
			Template: &runpb.RevisionTemplate{
				Revision: service.GenRevisionString(),
				Containers: []*runpb.Container{
					{
						Image: run.config.ImageFullPath(service),
					},
				},
			},
		},
		ServiceId: service.Name,
	}
	op, err := run.client.CreateService(ctx, req)
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (run *RunService) GetRunService(ctx context.Context, service *Service) (*runpb.Service, error) {
	req := &runpb.GetServiceRequest{
		Name: run.config.ServicePath(service),
	}
	return run.client.GetService(ctx, req)
}

func (run *RunService) SetIamPolicy(ctx context.Context, service *Service) error {
	req := &iam.SetIamPolicyRequest{
		Resource: run.config.ServicePath(service),
		Policy: &iam.Policy{
			Bindings: []*iam.Binding{
				{
					Role:    "roles/run.invoker",
					Members: []string{"allUsers"},
				},
			},
		},
	}

	_, err := run.client.SetIamPolicy(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (run *RunService) ListDeployedServices(ctx context.Context, region string) ([]*runpb.Service, error) {
	req := &runpb.ListServicesRequest{
		Parent: run.config.GetLocationByRegion(region),
	}

	it := run.client.ListServices(ctx, req)
	foundServices := []*runpb.Service{}
	for {
		resp, err := it.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		foundServices = append(foundServices, resp)
	}

	return foundServices, nil
}

func (run *RunService) DeployServiceSequence(ctx context.Context, service *Service) (*string, error) {
	var err error
	err = PushImage(ctx, service.Name, run.config)
	if err != nil {
		return nil, err
	}

	err = run.UpsertRunService(ctx, service)
	if err != nil {
		return nil, err
	}

	foundService, err := run.GetRunService(ctx, service)
	if err != nil {
		return nil, err
	}

	if service.Public {
		err = run.SetIamPolicy(ctx, service)
		if err != nil {
			return nil, err
		}
	}

	return &foundService.Uri, nil
}

func (run *RunService) DeleteRunService(ctx context.Context, service *Service) error {
	req := &runpb.DeleteServiceRequest{
		Name: run.config.ServicePath(service),
	}
	op, err := run.client.DeleteService(ctx, req)
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}
