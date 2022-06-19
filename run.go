package fabric

import (
	"context"
	"strings"

	run "cloud.google.com/go/run/apiv2"
	"google.golang.org/api/option"
	runpb "google.golang.org/genproto/googleapis/cloud/run/v2"
	"google.golang.org/genproto/googleapis/iam/v1"
)

func UpsertRunService(ctx context.Context, serviceName string, config *Config) error {
	service, err := config.GetService(serviceName)
	if err != nil {
		return err
	}

	client, err := run.NewServicesClient(ctx, option.WithCredentialsFile(config.Organization.Credsfile))
	if err != nil {
		return err
	}
	defer client.Close()

	req := &runpb.UpdateServiceRequest{
		Service: &runpb.Service{
			Name: config.ServicePath(service),
			Template: &runpb.RevisionTemplate{
				Revision: service.GenRevisionString(),
				Containers: []*runpb.Container{
					{
						Image: config.ImageFullPath(service),
					},
				},
			},
		},
		AllowMissing: true,
	}

	op, err := client.UpdateService(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "code = NotFound") {
			err = createRunService(ctx, service, config)
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

func createRunService(ctx context.Context, service *Service, config *Config) error {
	client, err := run.NewServicesClient(ctx, option.WithCredentialsFile(config.Organization.Credsfile))
	if err != nil {
		return err
	}
	defer client.Close()

	req := &runpb.CreateServiceRequest{
		Parent: config.GetLocation(service),
		Service: &runpb.Service{
			Template: &runpb.RevisionTemplate{
				Revision: service.GenRevisionString(),
				Containers: []*runpb.Container{
					{
						Image: config.ImageFullPath(service),
					},
				},
			},
		},
		ServiceId: service.Name,
	}
	op, err := client.CreateService(ctx, req)
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}

func GetRunService(ctx context.Context, serviceName string, config *Config) (*runpb.Service, error) {
	service, err := config.GetService(serviceName)
	if err != nil {
		return nil, err
	}

	client, err := run.NewServicesClient(ctx, option.WithCredentialsFile(config.Organization.Credsfile))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	req := &runpb.GetServiceRequest{
		Name: config.ServicePath(service),
	}
	return client.GetService(ctx, req)
}

func SetIamPolicy(ctx context.Context, serviceName string, config *Config) error {
	service, err := config.GetService(serviceName)
	if err != nil {
		return err
	}

	client, err := run.NewServicesClient(ctx, option.WithCredentialsFile(config.Organization.Credsfile))
	if err != nil {
		return err
	}
	defer client.Close()

	req := &iam.SetIamPolicyRequest{
		Resource: config.ServicePath(service),
		Policy: &iam.Policy{
			Bindings: []*iam.Binding{
				{
					Role:    "roles/run.invoker",
					Members: []string{"allUsers"},
				},
			},
		},
	}

	_, err = client.SetIamPolicy(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
