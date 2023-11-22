package invoke

import (
	"crypto/sha1"
	"sync"
)

type PlacementCache struct {
	applicationNames    map[string]int
	applicationNamesRev map[int]string
	// Map from actor type + key hash to application-name index.
	// We use a byte hash and index to application name for performance
	// reasons (saves us string pointers) => https://go101.org/optimizations/6-map.html
	items       map[[8]byte]int
	mutex       *sync.RWMutex
	capacity    int
	capacityLow int
}

func NewPlacementCache() *PlacementCache {
	var mutex sync.RWMutex
	cache := PlacementCache{
		applicationNames:    make(map[string]int),
		applicationNamesRev: make(map[int]string),
		items:               make(map[[8]byte]int),
		mutex:               &mutex,
		capacity:            1000,
		capacityLow:         950,
	}
	return &cache
}

func (cache PlacementCache) Update(actorType string, actorId []string, appId string) {
	hash := hashActor(actorType, actorId)
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	appIdx, appExists := cache.applicationNames[appId]
	if !appExists {
		appIdx = len(cache.applicationNames)
		cache.applicationNames[appId] = appIdx
		cache.applicationNamesRev[appIdx] = appId
	}
	cache.items[hash] = appIdx
	if len(cache.items) > cache.capacity {
		items := cache.items
		n := len(cache.items) - cache.capacityLow
		for key := range cache.items {
			if n < 0 {
				break
			}
			delete(items, key)
		}
	}
}

func (cache PlacementCache) Prepare(actorType string, actorId []string) [8]byte {
	return hashActor(actorType, actorId)
}

func (cache PlacementCache) Delete(prepared [8]byte) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	delete(cache.items, prepared)
}

func (cache PlacementCache) Get(prepared [8]byte) *string {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	appIdx, exists := cache.items[prepared]

	if !exists {
		return nil
	}

	appId, appExists := cache.applicationNamesRev[appIdx]
	if !appExists {
		return nil
	}
	return &appId
}

func hashActor(actorType string, actorId []string) [8]byte {
	// Yes, sha1 is broken. But faster than sha256. We only need it to avoid collisions.
	// When we have collisions, that is no big issue. We just return the appId for
	// a different actor, but the retry mechanism of dynamicinvoker will get this right
	// in the retry.
	h := sha1.New()
	h.Write([]byte(actorType))
	for _, part := range actorId {
		h.Write([]byte(part))
	}
	return [8]byte(h.Sum(nil))
}
