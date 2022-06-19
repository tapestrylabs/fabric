package fabric

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/mitchellh/go-homedir"
)

func GetContext(path string) io.Reader {
	filePath, err := homedir.Expand(path)
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := archive.TarWithOptions(filePath, &archive.TarOptions{})
	return ctx
}

func PushImage(ctx context.Context, serviceName string, config *Config) error {
	service, err := config.GetService(serviceName)
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	resp, err := cli.ImageBuild(ctx, GetContext(service.Context), types.ImageBuildOptions{
		Tags:        []string{config.ImageFullPath(service)},
		ForceRemove: true,
		NoCache:     true,
	})
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	// Need to drain the resp otherwise it won't block
	_, err = ioutil.ReadAll(resp.Body)

	jsonKey, err := ioutil.ReadFile(config.Organization.Credsfile)
	if err != nil {
		return err
	}

	authConfig := types.AuthConfig{
		ServerAddress: "us-docker.pkg.dev",
		Username:      "_json_key",
		Password:      string(jsonKey),
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	res, err := cli.ImagePush(ctx, config.ImageFullPath(service), types.ImagePushOptions{
		RegistryAuth: authStr,
	})
	defer res.Close()
	if err != nil {
		return err
	}
	// Need to drain the resp otherwise it won't block
	_, err = ioutil.ReadAll(res)

	_, err = cli.ImagesPrune(ctx, filters.Args{})
	if err != nil {
		return err
	}

	return nil
}
