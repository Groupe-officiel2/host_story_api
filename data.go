package main

import (
	"encoding/json"
	"os"
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
