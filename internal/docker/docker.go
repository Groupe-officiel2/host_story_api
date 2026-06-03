package docker

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func FindAvailablePort() (int, error) {
	basePort := 42420
	cli, err := GetDockerClient()
	used := map[int]bool{}
	if err == nil {
		containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
		if err == nil {
			for _, c := range containers {
				for _, p := range c.Ports {
					if p.PublicPort != 0 {
						used[int(p.PublicPort)] = true
					}
				}
			}
		}
	}
	for port := basePort; port < basePort+1000; port++ {
		if used[port] {
			continue
		}
		tcpLn, tcpErr := net.Listen("tcp", ":"+strconv.Itoa(port))
		if tcpErr != nil {
			continue
		}
		udpLn, udpErr := net.ListenPacket("udp", ":"+strconv.Itoa(port))
		if udpErr != nil {
			tcpLn.Close()
			continue
		}
		tcpLn.Close()
		udpLn.Close()
		return port, nil
	}
	return 0, fmt.Errorf("no available ports found")
}

// CORRECTION : Ajout de "slots int" dans les paramètres
// 1. AJOUT de "slots int" à la fin des paramètres de la fonction
func CreateContainer(ctx context.Context, image string, name string, hostPort string, memoryLimit int64, ownerID string, slots int) (string, error) {
    containerPort := "42420"

    cli, err := GetDockerClient()
    if err != nil {
       return "", fmt.Errorf("docker client error: %w", err)
    }

    hostConfig := &container.HostConfig{
       PortBindings: nat.PortMap{
            nat.Port(containerPort + "/tcp"): []nat.PortBinding{{HostIP:   "0.0.0.0", HostPort: hostPort}},
            nat.Port(containerPort + "/udp"): []nat.PortBinding{{HostIP:   "0.0.0.0", HostPort: hostPort}},
        },
       Resources: container.Resources{
          Memory: memoryLimit,
       },
    }

    labels := map[string]string{
       "owner-id": ownerID,
       "slots":    strconv.Itoa(slots),
       "name":     name,
    }
    fmt.Printf("CreateContainer labels=%v name=%s image=%s hostPort=%s slots=%d\n", labels, name, image, hostPort, slots)

    resp, err := cli.ContainerCreate(ctx, &container.Config{
       Image: image,
       ExposedPorts: nat.PortSet{
                nat.Port(containerPort + "/tcp"): struct{}{},
                nat.Port(containerPort + "/udp"): struct{}{},
            },
       Labels: labels,
       Tty:       true,
       OpenStdin: true,
    }, hostConfig, nil, nil, name)
    if err != nil {
       return "", fmt.Errorf("container create error: %w", err)
    }

    return resp.ID, nil
}

func GetDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func StartContainer(ctx context.Context, id string) error {
	cli, err := GetDockerClient()
	if err != nil {
		return err
	}
	return cli.ContainerStart(ctx, id, container.StartOptions{})
}

func StopContainer(ctx context.Context, id string) error {
	cli, err := GetDockerClient()
	if err != nil {
		return err
	}
	if err := cli.ContainerKill(ctx, id, "SIGTERM"); err != nil {
		return fmt.Errorf("SIGTERM error: %w", err)
	}
	time.Sleep(2 * time.Second)
	if err := cli.ContainerStop(ctx, id, container.StopOptions{}); err != nil {
		return fmt.Errorf("stop error: %w", err)
	}
	return nil
}

func RemoveContainer(ctx context.Context, id string) error {
	cli, err := GetDockerClient()
	if err != nil {
		return err
	}
	return cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
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
	cli, err := GetDockerClient()
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
		if err := StartContainer(ctx, id); err != nil {
			return "", err
		}
		return "started", nil
	}
	if err := StopContainer(ctx, id); err != nil {
		return "", err
	}
	return "stopped", nil
}