package internal

const (
	WEBSOCKET_LEVEL_PATH    = "/socket/level"
	WEBSOCKET_DATAPOLL_PATH = "/socket/datapoll"
)

// Pointer converts any value to a pointer to the same value.
func Pointer[K any](val K) *K {
	return &val
}
