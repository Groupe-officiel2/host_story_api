package main

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// checks for an available port starting from a given base port
func findAvailablePort(basePort int) (int, error) {
	for port := basePort; port < basePort+1000; port++ {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found")
}

func CreateContainerFromTemplate(ctx context.Context, image string, name string, hostPort string, containerPort string, memoryLimit int64, ownerID string) (string, error) {
	// Check if the port is available
	availablePort, err := findAvailablePort(baseHostPort + serverCounter)
	if err != nil {
		return "", fmt.Errorf("no available ports: %w", err)
	}
	hostPort = strconv.Itoa(availablePort)

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithVersion("1.44"),
	)
	if err != nil {
		return "", fmt.Errorf("docker client error: %w", err)
	}

	hostConfig := &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			nat.Port(containerPort + "/tcp"): {{HostPort: hostPort}},
			nat.Port(containerPort + "/udp"): {{HostPort: hostPort}},
		},
		Resources: container.Resources{
			Memory: memoryLimit,
		},
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Labels: map[string]string{
			"owner-id": ownerID,
		},
		Tty: true,
		OpenStdin: true,
	}, hostConfig, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("container create error: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("container start error: %w", err)
	}

	return resp.ID, nil
}

func getDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
}

func StartContainer(ctx context.Context, id string) error {
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	return cli.ContainerStart(ctx, id, container.StartOptions{})
}

func StopContainer(ctx context.Context, id string) error {
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	// SIGTERM pour sauvegarde propre
	if err := cli.ContainerKill(ctx, id, "SIGTERM"); err != nil {
		return fmt.Errorf("SIGTERM error: %w", err)
	}

	// Laisser le serveur sauvegarder
	time.Sleep(2 * time.Second)

	// Stop forcé si nécessaire
	if err := cli.ContainerStop(ctx, id, container.StopOptions{}); err != nil {
		return fmt.Errorf("stop error: %w", err)
	}

	return nil
}

func findContainerIDByName(ctx context.Context, cli *client.Client, name string) (string, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return "", err
	}

	for _, c := range containers {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == name {
				return c.ID, nil
			}
		}
	}

	return "", fmt.Errorf("container %q not found", name)
}

func ToggleContainer(ctx context.Context, name string) (string, error) {
	cli, err := getDockerClient()
	if err != nil {
		return "", err
	}

	id, err := findContainerIDByName(ctx, cli, name)
	if err != nil {
		return "", err
	}

	inspect, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return "", fmt.Errorf("inspect error: %w", err)
	}

	if !inspect.State.Running {
		// Le conteneur est OFF → ON
		if err := StartContainer(ctx, id); err != nil {
			return "", err
		}
		return "started", nil
	}

	// Le conteneur est ON → OFF
	if err := StopContainer(ctx, id); err != nil {
		return "", err
	}

	return "stopped", nil
}
