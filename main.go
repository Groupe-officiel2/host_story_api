package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const port = ":8080"
const stateFile = "server_state.json"

var (
	serverCounter     int
	serverMutex       sync.Mutex
	baseHostPort      = 42720
	baseContainerPort = 42420
)

const (
	defaultPlayerSlots = 1
	defaultMemoryLimit = int64(1395864371) // 1.30 GB in bytes (base 1024)
	memoryPerPlayer    = int64(314572800)  // 300 MB in bytes (base 1024)
)

type ServerState struct {
	ServerCounter int `json:"server_counter"`
	BaseHostPort  int `json:"base_host_port"`
}

func loadServerState() (*ServerState, error) {
	file, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &ServerState{ServerCounter: 0, BaseHostPort: baseHostPort}, nil
		}
		return nil, err
	}

	var state ServerState
	if err := json.Unmarshal(file, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func saveServerState(state *ServerState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile, data, 0644)
}

func CreateTemplateContainer(w http.ResponseWriter, r *http.Request) {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	state, err := loadServerState()
	if err != nil {
		http.Error(w, "Failed to load server state", http.StatusInternalServerError)
		return
	}

	state.ServerCounter++

	// Get container name from query parameter
	name := r.URL.Query().Get("name")
	if name == "" {
		name = fmt.Sprintf("server%d", state.ServerCounter)
	}

	hostPort := fmt.Sprintf("%d", state.BaseHostPort+state.ServerCounter)
	containerPort := fmt.Sprintf("%d", baseContainerPort)

	image := r.URL.Query().Get("image")
	if image == "" {
		image = "server-vintagestory:latest"
	}

	// Get player slots from query parameter
	playerSlotsParam := r.URL.Query().Get("players")
	playerSlots := defaultPlayerSlots
	if playerSlotsParam != "" {
		if parsedSlots, err := strconv.Atoi(playerSlotsParam); err == nil && parsedSlots > 0 {
			playerSlots = parsedSlots
		}
	}

	// Calculate memory limit based on player slots
	totalMemory := defaultMemoryLimit + (int64(playerSlots-1) * memoryPerPlayer)

	containerID, err := createContainerFromTemplate(r.Context(), image, name, hostPort, containerPort, totalMemory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := saveServerState(state); err != nil {
		http.Error(w, "Failed to save server state", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Container launched: %s with name %s on host port %s and %d player slots", containerID, name, hostPort, playerSlots)
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

func createContainerFromTemplate(ctx context.Context, image string, name string, hostPort string, containerPort string, memoryLimit int64) (string, error) {
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
