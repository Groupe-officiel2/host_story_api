package main

import (
	"context"
	"net"
	"strconv"

	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// findAvailablePort checks for an available port starting from a given base port
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

func CreateContainerFromTemplate(ctx context.Context, image string, name string, hostPort string, containerPort string, memoryLimit int64) (string, error) {
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
	}, hostConfig, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("container create error: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("container start error: %w", err)
	}

	return resp.ID, nil
}
