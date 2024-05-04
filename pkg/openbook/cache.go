package openbook

import (
	"sync"
)

// Map where key is mint address string and value is *OpenbookInfo
var openbookCache = make(map[string]*OpenbookInfo)
var openbookCacheMutex = &sync.Mutex{}

// TODO: Add a TTL for the cache (and flush unused entries)

// SetOpenbookInfo sets the OpenbookInfo for the given mint address.
func SetOpenbookInfo(mintAddress string, info *OpenbookInfo) {
	openbookCacheMutex.Lock()
	defer openbookCacheMutex.Unlock()

	openbookCache[mintAddress] = info
}

// GetOpenbookInfo returns the OpenbookInfo for the given mint address.
func GetOpenbookInfo(mintAddress string) *OpenbookInfo {
	openbookCacheMutex.Lock()
	defer openbookCacheMutex.Unlock()

	if info, ok := openbookCache[mintAddress]; ok {
		delete(openbookCache, mintAddress)
		return info
	}

	return nil
}
