package common

import "sync"

var (
	ServerCounter int
	ServerMutex   sync.Mutex
)
