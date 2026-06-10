package handlers

import (
	"bytes"
	"context"
	"io"
	"log"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func GetPlayersForServer(containerName string) int {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return 0
	}

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return 0
	}

	for _, c := range containers {
		for _, name := range c.Names {

			if strings.Contains(name, containerName) {

				reader, err := cli.ContainerLogs(context.Background(), c.ID, container.LogsOptions{
					ShowStdout: true,
					ShowStderr: true,
					Tail:       "all",
				})
				if err != nil {
					log.Printf("Logs err: %v", err)
					return 0
				}

				contentBytes, err := io.ReadAll(reader)
				if err != nil {
					return 0
				}

				var stdout, stderr bytes.Buffer
				_, err = stdcopy.StdCopy(&stdout, &stderr, bytes.NewReader(contentBytes))
				var logContent string
				if err != nil {
					logContent = string(contentBytes)
				} else {
					logContent = stdout.String()
				}

				players := 0
				for _, line := range strings.Split(logContent, "\n") {
					if strings.Contains(line, " joins.") {
						players++
					} else if strings.Contains(line, " disconnected.") {
						players--
						if players < 0 {
							players = 0
						}
					}
				}
				log.Printf("Container %s: computed players=%d", containerName, players)

				return players
			}
		}
	}

	return 0
}
