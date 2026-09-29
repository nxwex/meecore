package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/moby/client"
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
