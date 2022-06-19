package fabric

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

type Creds struct {
	ProjectId string `json:"project_id"`
}

type Organization struct {
	Name      string
	Credsfile string
	Services  []struct {
		Service Service
	}
}

type Service struct {
	Name     string
	Context  string
	Region   string
	Public   bool
	Database Database
}

type Database struct {
	Name    string
	Dialect string
	Version int
	Vcpu    int
	Memory  int64
}

type Config struct {
	Organization Organization
	Creds        Creds
}

func NewConfig() *Config {
	data, err := ioutil.ReadFile("fabric.yaml")
	if err != nil {
		log.Fatal(err)
	}

	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}

	credsData, err := ioutil.ReadFile(config.Organization.Credsfile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(credsData, &config.Creds)
	if err != nil {
		log.Fatal(err)
	}

	return &config
}

func (s *Service) GenRevisionString() string {
	return fmt.Sprintf("%s-%s", s.Name, String(5))
}

func (c *Config) GetService(serviceName string) (*Service, error) {
	for _, v := range c.Organization.Services {
		if strings.EqualFold(v.Service.Name, serviceName) {
			return &v.Service, nil
		}
	}
	return nil, errors.New("Failed to find service")
}

func (c *Config) GetLocation(service *Service) string {
	return fmt.Sprintf("projects/%s/locations/%s", c.Creds.ProjectId, service.Region)
}

func (c *Config) GetLocationByRegion(region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", c.Creds.ProjectId, region)
}

func (c *Config) ServicePath(service *Service) string {
	return fmt.Sprintf("projects/%s/locations/%s/services/%s", c.Creds.ProjectId, service.Region, service.Name)
}

func (c *Config) RepoBasePath() string {
	return fmt.Sprintf("projects/%s/locations/us", c.Creds.ProjectId)
}

func (c *Config) RepoFullPath() string {
	return fmt.Sprintf("%s/repositories/%s", c.RepoBasePath(), c.Organization.Name)
}

func (c *Config) ImageFullPath(service *Service) string {
	return path.Join("us-docker.pkg.dev", c.Creds.ProjectId, c.Organization.Name, service.Name)
}

func (c *Config) GetSqlInstanceTier(service *Service) string {
	return fmt.Sprintf("db-custom-%d-%d", service.Database.Vcpu, service.Database.Memory)
}
