package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"host_story_api/internal/auth"
	"host_story_api/internal/common"
	"host_story_api/internal/docker"

	"github.com/docker/docker/api/types/container"
)

func CreateTemplateContainer(w http.ResponseWriter, r *http.Request) {
	common.ServerMutex.Lock()
	defer common.ServerMutex.Unlock()

	ownerID := auth.UserIDFromContext(r.Context())
	if ownerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Container name is required", http.StatusBadRequest)
		return
	}
	fmt.Printf("CreateTemplateContainer request name=%s players=%s owner=%s\n", name, r.URL.Query().Get("players"), ownerID)
	availablePort, err := docker.FindAvailablePort()
	if err != nil {
		http.Error(w, fmt.Sprintf("no available ports: %v", err), http.StatusInternalServerError)
		return
	}
	hostPort := strconv.Itoa(availablePort)
	image := r.URL.Query().Get("image")
	if image == "" {
		image = "server-vintagestory:latest"
	}
	playerSlotsParam := r.URL.Query().Get("players")
	playerSlots := 2
	if playerSlotsParam != "" {
		if parsedSlots, err := strconv.Atoi(playerSlotsParam); err == nil && parsedSlots > 0 {
			playerSlots = parsedSlots
		}
	}
	const (
		defaultMemoryLimit = int64(1395864371)
		memoryPerPlayer    = int64(314572800)
	)
	totalMemory := defaultMemoryLimit + (int64(playerSlots-1) * memoryPerPlayer)

	var containerID string
	var createErr error
	for attempt := 0; attempt < 5; attempt++ {
		// CORRECTION : On passe playerSlots en dernier paramètre
		containerID, createErr = docker.CreateContainer(r.Context(), image, name, hostPort, totalMemory, ownerID, playerSlots)
		if createErr == nil {
			break
		}
		if strings.Contains(createErr.Error(), "port is already allocated") {
			availablePort, err := docker.FindAvailablePort()
			if err != nil {
				http.Error(w, fmt.Sprintf("no available ports: %v", err), http.StatusInternalServerError)
				return
			}
			hostPort = strconv.Itoa(availablePort)
			continue
		}
		http.Error(w, createErr.Error(), http.StatusInternalServerError)
		return
	}
	if createErr != nil {
		http.Error(w, createErr.Error(), http.StatusInternalServerError)
		return
	}
	err = docker.StartContainer(r.Context(), containerID)
	if err != nil {
		_ = docker.RemoveContainer(r.Context(), containerID)
		http.Error(w, fmt.Sprintf("failed to start container: %v", err), http.StatusInternalServerError)
		return
	}
	go func() {
		time.Sleep(500 * time.Millisecond)
		jsonData := fmt.Sprintf(`{"id": "%s", "name": "%s", "slots": %d}`, containerID, name, playerSlots)
		req, err := http.NewRequest("POST", "http://nginx/api/servers", strings.NewReader(jsonData))
		if err != nil {
			fmt.Println("Laravel request error:", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", "SECRET123")
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Laravel API error:", err)
			return
		}
		defer resp.Body.Close()
		fmt.Println("Server saved in Laravel:", resp.Status)
	}()
	SVRPort, err := strconv.Atoi(hostPort)
	if err == nil {
		docker.AddSRVRecord(SVRPort, name)
	}
	fmt.Fprintf(w, "Container launched: %s with name %s on host port %s and %d player slots\n", containerID, name, hostPort, playerSlots)
}

func ToggleHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "server name is required", http.StatusBadRequest)
		return
	}
	actorRole := auth.UserRoleFromContext(r.Context())
	if !strings.EqualFold(actorRole, "admin") {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}
	result, err := docker.ToggleContainer(context.Background(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Container %s %s\n", name, result)
}

func GetPlayers(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "server name required", http.StatusBadRequest)
		return
	}
	players := GetPlayersForServer(name)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"players": players})
}

func GetServers(w http.ResponseWriter, r *http.Request) {
	cli, err := docker.GetDockerClient()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	liveServers := make([]map[string]interface{}, 0)
	for _, c := range containers {
		if c.Labels["owner-id"] == "" && c.Labels["com.docker.compose.service"] != "vintagestory" && !strings.Contains(c.Image, "server-vintagestory") {
			continue
		}
		name := c.Labels["name"]
		if name == "" && len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		slots := 1
		if val, ok := c.Labels["slots"]; ok {
			if p, err := strconv.Atoi(val); err == nil {
				slots = p
			}
		} else if val, ok := c.Labels["players"]; ok {
			if p, err := strconv.Atoi(val); err == nil {
				slots = p
			}
		}
		var port int
		if len(c.Ports) > 0 {
			for _, p := range c.Ports {
				if p.PublicPort != 0 {
					port = int(p.PublicPort)
					break
				}
			}
		}
		liveServers = append(liveServers, map[string]interface{}{
			"ID":      c.ID[:12],
			"Name":    name,
			"Players": GetPlayersForServer(name),
			"Slots":   slots,
			"Port":    port,
			"State":   c.State,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(liveServers)
}
