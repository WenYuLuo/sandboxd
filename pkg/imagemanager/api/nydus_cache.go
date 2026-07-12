// Copyright (c) 2026 Ant Group Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

// NydusImageCache caches whether an image URL is a Nydus image to avoid repeated registry fetches
type NydusImageCache struct {
	mu       sync.RWMutex
	positive *expirable.LRU[string, struct{}]
	negative *expirable.LRU[string, struct{}]
}

// CacheConfig holds configuration for the cache
type CacheConfig struct {
	PositiveTTL     time.Duration // TTL for positive results, default 1 hour
	NegativeTTL     time.Duration // TTL for negative results, default 5 minutes
	MaxCacheEntries int           // Maximum cache entries, default 1000
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		PositiveTTL:     1 * time.Hour,
		NegativeTTL:     5 * time.Minute,
		MaxCacheEntries: 1000,
	}
}

// NewNydusImageCache creates a new cache instance with default config
func NewNydusImageCache() *NydusImageCache {
	return NewNydusImageCacheWithConfig(DefaultCacheConfig())
}

// NewNydusImageCacheWithConfig creates a new cache instance with custom config
func NewNydusImageCacheWithConfig(config *CacheConfig) *NydusImageCache {
	if config == nil {
		config = DefaultCacheConfig()
	}
	if config.MaxCacheEntries <= 0 {
		config.MaxCacheEntries = DefaultCacheConfig().MaxCacheEntries
	}
	if config.PositiveTTL <= 0 {
		config.PositiveTTL = DefaultCacheConfig().PositiveTTL
	}
	if config.NegativeTTL <= 0 {
		config.NegativeTTL = DefaultCacheConfig().NegativeTTL
	}

	positiveCap := config.MaxCacheEntries / 2
	if positiveCap == 0 {
		positiveCap = 1
	}
	negativeCap := config.MaxCacheEntries - positiveCap
	if negativeCap == 0 {
		negativeCap = 1
	}

	return &NydusImageCache{
		positive: expirable.NewLRU[string, struct{}](positiveCap, nil, config.PositiveTTL),
		negative: expirable.NewLRU[string, struct{}](negativeCap, nil, config.NegativeTTL),
	}
}

// Get retrieves cached result for an image URL
// Returns (isNydus, found)
func (c *NydusImageCache) Get(imageURL string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, found := c.positive.Get(imageURL); found {
		return true, true
	}
	if _, found := c.negative.Get(imageURL); found {
		return false, true
	}
	return false, false
}

// Set stores the result for an image URL
func (c *NydusImageCache) Set(imageURL string, isNydus bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if isNydus {
		c.negative.Remove(imageURL)
		c.positive.Add(imageURL, struct{}{})
		return
	}
	c.positive.Remove(imageURL)
	c.negative.Add(imageURL, struct{}{})
}

// Clear removes all cached entries
func (c *NydusImageCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.positive.Purge()
	c.negative.Purge()
}

// Size returns the number of cached entries
func (c *NydusImageCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.positive.Len() + c.negative.Len()
}

// Invalidate removes a specific entry from the cache
func (c *NydusImageCache) Invalidate(imageURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.positive.Remove(imageURL)
	c.negative.Remove(imageURL)
}
