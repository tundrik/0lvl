//go:build appengine || windows
// +build appengine windows

package cache

func getChunk() []byte {
	return make([]byte, chunkSize)
}

func putChunk(chunk []byte) {
	// No-op.
}