package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// CreateTemplateContainer handles the creation of a new container from a template
func CreateTemplateContainer(w http.ResponseWriter, r *http.Request) {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	ownerID := UserIDFromContext(r.Context())
	if ownerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get container name from query parameter
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Container name is required", http.StatusBadRequest)
		return
	}

	availablePort, err := findAvailablePort()
	if err != nil {
		http.Error(w, fmt.Sprintf("no available ports: %v", err), http.StatusInternalServerError)
		return
	}
	hostPort := strconv.Itoa(availablePort)

	image := r.URL.Query().Get("image")
	if image == "" {
		image = "server-vintagestory:latest"
	}

	// Get player slots from query parameter
	playerSlotsParam := r.URL.Query().Get("players")
	playerSlots := 2
	if playerSlotsParam != "" {
		if parsedSlots, err := strconv.Atoi(playerSlotsParam); err == nil && parsedSlots > 0 {
			playerSlots = parsedSlots
		}
	}

	const (
		defaultMemoryLimit = int64(1395864371) // 1.30 GB in bytes (base 1024)
		memoryPerPlayer    = int64(314572800)  // 300 MB in bytes (base 1024)
	)

	// Calculate memory limit based on player slots
	totalMemory := defaultMemoryLimit + (int64(playerSlots-1) * memoryPerPlayer)

	containerID, err := CreateContainer(r.Context(), image, name, hostPort, totalMemory, ownerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = StartContainer(r.Context(), containerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to start container: %v", err), http.StatusInternalServerError)
		return
	}

	SVRPort, err := strconv.Atoi(hostPort)
	AddSRVRecord(SVRPort, name)
	fmt.Fprintf(w, "Container launched: %s with name %s on host port %s and %d player slots\n", containerID, name, hostPort, playerSlots)
}

func ToggleHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	if name == "" {
		http.Error(w, "server name is required", http.StatusBadRequest)
		return
	}

	actorRole := UserRoleFromContext(r.Context())
	if !strings.EqualFold(actorRole, "admin") {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	result, err := ToggleContainer(context.Background(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Container %s %s\n", name, result)
}
