package failoverassistant

import (
	"sync"
	"time"
)

// VolumeFailoverRestRequest struct
type VolumeFailoverRestRequest struct {
	VolumeName        string `json:"volumeName,omitempty"`
	StorageClassName  string `json:"storageClassName,omitempty"`
	VolmeAttachmentID string `json:"volumeAttachmentId,omitempty"`
	NodeName          string `json:"nodeName,omitempty"`
}

// Cache interface
type Cache interface {
	Set(key string, obj interface{}, expiration time.Duration)
	Get(key string) (interface{}, bool)
}

type cacheValue struct {
	Expire time.Time
	Object interface{}
}

type expiredCache struct {
	m sync.Map
}

// NewCache creates a cache with expired function
func NewCache() Cache {
	return &expiredCache{}
}

func (ec *expiredCache) Set(key string, obj interface{}, expiration time.Duration) {
	value := &cacheValue{
		Expire: time.Now().Add(expiration),
		Object: obj,
	}
	ec.m.Store(key, value)
}

// return: value, expired
func (ec *expiredCache) Get(key string) (interface{}, bool) {
	item, ok := ec.m.Load(key)
	if !ok || item == nil {
		return nil, true
	}
	value, ok := item.(*cacheValue)
	if !ok {
		return nil, true
	}
	if value.Expire.Before(time.Now()) {
		ec.m.Delete(key)
	}
	return value.Object, value.Expire.Before(time.Now())
}
