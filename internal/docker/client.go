package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/nxwex/meecore/internal/service"
)

type Client struct {
	client *client.Client
}

func New(ctx context.Context) (*Client, error) {
	dockerClient, err := client.New(
		client.FromEnv,
	)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	if _, err := dockerClient.Ping(ctx, client.PingOptions{}); err != nil {
		dockerClient.Close()
		return nil, fmt.Errorf("ping docker: %w", err)
	}

	return &Client{client: dockerClient}, nil
}

func (c *Client) Close() {
	c.client.Close()
}

func (c *Client) GetConatiners(ctx context.Context) ([]Container, error) {
	result, err := c.client.ContainerList(ctx, client.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	containers := make([]Container, 0, len(result.Items))

	for _, item := range result.Items {
		name := ""
		if len(item.Names) > 0 {
			name = strings.TrimPrefix(item.Names[0], "/")
		}

		containers = append(containers, Container{
			ID:    item.ID,
			Name:  name,
			Image: item.Image,
			State: string(item.State),
		})
	}

	return containers, nil
}

func (c *Client) GetContainer(ctx context.Context, id string) (*Container, error) {
	cont, err := c.client.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("inspect container: %w", err)
	}

	return &Container{
		ID:    cont.Container.ID,
		Name:  strings.TrimPrefix(cont.Container.Name, "/"),
		Image: cont.Container.Config.Image,
		State: string(cont.Container.State.Status),
	}, nil
}

func (c *Client) StartContainer(ctx context.Context, id string) error {
	_, err := c.client.ContainerStart(ctx, id, client.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	return nil
}

func (c *Client) StopContainer(ctx context.Context, id string) error {
	_, err := c.client.ContainerStop(ctx, id, client.ContainerStopOptions{})
	if err != nil {
		return fmt.Errorf("stop container: %w", err)
	}

	return nil
}

func (c *Client) RestartContainer(ctx context.Context, id string) error {
	_, err := c.client.ContainerRestart(ctx, id, client.ContainerRestartOptions{})
	if err != nil {
		return fmt.Errorf("restart container: %w", err)
	}

	return nil
}

func (c *Client) RemoveContainer(ctx context.Context, id string) error {
	_, err := c.client.ContainerRemove(ctx, id, client.ContainerRemoveOptions{})
	if err != nil {
		return fmt.Errorf("remove container: %w", err)
	}

	return nil
}

func (c *Client) CreateContainer(ctx context.Context, name string, template service.Template) (string, error) {
	env := make([]string, 0, len(template.Env))

	for _, setting := range template.Env {
		value := setting.Default

		env = append(env, setting.Name+"="+value)
	}

	if err := c.ensureImage(ctx, template.Image); err != nil {
		return "", fmt.Errorf("ensure image: %w", err)
	}

	cont, err := c.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Name: name,
		Config: &container.Config{
			Image: template.Image,
			Env:   env,
		},
	})
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}

	return cont.ID, nil
}

func (c *Client) imageExists(ctx context.Context, imageName string) (bool, error) {
	_, err := c.client.ImageInspect(ctx, imageName)
	if err == nil {
		return true, nil
	}

	if errdefs.IsNotFound(err) {
		return false, nil
	}

	return false, fmt.Errorf("inspect image: %w", err)
}

func (c *Client) pullImage(ctx context.Context, imageName string) error {
	log.Printf("start pulling image %s", imageName)
	reader, err := c.client.ImagePull(ctx, imageName, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pull image: %w", err)
	}
	defer reader.Close()

	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("read pull response: %w", err)
	}
	log.Printf("image %s pulled successfully", imageName)

	return nil
}

func (c *Client) ensureImage(ctx context.Context, imageName string) error {
	exists, err := c.imageExists(ctx, imageName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return c.pullImage(ctx, imageName)
}

func (c *Client) AttachContainer(ctx context.Context, id string) (client.HijackedResponse, error) {
	response, err := c.client.ContainerAttach(ctx, id, client.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Logs:   true,
	})
	if err != nil {
		return client.HijackedResponse{}, fmt.Errorf("attach container: %w", err)
	}

	return response.HijackedResponse, nil
}

func (c *Client) ExecuteCommand(ctx context.Context, containerID string, command string) (string, error) {
	exec, err := c.client.ExecCreate(
		ctx,
		containerID,
		client.ExecCreateOptions{
			Cmd:          []string{"rcon-cli", command},
			AttachStdout: true,
			AttachStderr: true,
		},
	)
	if err != nil {
		return "", fmt.Errorf("create exec: %w", err)
	}

	response, err := c.client.ExecAttach(
		ctx,
		exec.ID,
		client.ExecAttachOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("attach exec: %w", err)
	}
	defer response.Close()

	if _, err := c.client.ExecStart(
		ctx,
		exec.ID,
		client.ExecStartOptions{},
	); err != nil {
		return "", fmt.Errorf("start exec: %w", err)
	}

	var output bytes.Buffer

	if _, err := stdcopy.StdCopy(
		&output,
		&output,
		response.Reader,
	); err != nil {
		return "", fmt.Errorf("read exec output: %w", err)
	}

	return output.String(), nil
}
