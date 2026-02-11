package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const port = ":8080"

var (
	serverCounter     int
	serverMutex       sync.Mutex
	baseHostPort      = 42720
	baseContainerPort = 42420
)

func CreateTemplateContainer(w http.ResponseWriter, r *http.Request) {
	serverMutex.Lock()
	serverCounter++
	name := fmt.Sprintf("server%d", serverCounter)
	hostPort := fmt.Sprintf("%d", baseHostPort+serverCounter)
	containerPort := fmt.Sprintf("%d", baseContainerPort)
	serverMutex.Unlock()

	image := r.URL.Query().Get("image")
	if image == "" {
		image = "server-vintagestory:latest"
	}

	containerID, err := createContainerFromTemplate(r.Context(), image, name, hostPort, containerPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Container launched: %s with name %s on host port %s", containerID, name, hostPort)
}

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY is required")
	}

	http.HandleFunc("/template", withAPIKey(apiKey, CreateTemplateContainer))

	fmt.Printf("(http://localhost:8080) - Server is running on port %s\n", port)
	http.ListenAndServe(port, nil)
}

func createContainerFromTemplate(ctx context.Context, image string, name string, hostPort string, containerPort string) (string, error) {
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

func withAPIKey(apiKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provided := r.Header.Get("X-API-Key")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
