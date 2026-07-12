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
	"testing"
	"time"
)

func TestNydusImageCache_SetAndGet(t *testing.T) {
	cache := NewNydusImageCache()

	// Test cache miss
	_, found := cache.Get("test-image:latest")
	if found {
		t.Error("Expected cache miss, got hit")
	}

	// Test set and get
	cache.Set("test-image:latest", true)
	isNydus, found := cache.Get("test-image:latest")
	if !found {
		t.Error("Expected cache hit, got miss")
	}
	if !isNydus {
		t.Error("Expected isNydus=true, got false")
	}

	// Test negative cache
	cache.Set("regular-image:latest", false)
	isNydus, found = cache.Get("regular-image:latest")
	if !found {
		t.Error("Expected cache hit, got miss")
	}
	if isNydus {
		t.Error("Expected isNydus=false, got true")
	}
}

func TestNydusImageCache_CustomTTL(t *testing.T) {
	config := &CacheConfig{
		PositiveTTL:     2 * time.Second,
		NegativeTTL:     1 * time.Second,
		MaxCacheEntries: 100,
	}
	cache := NewNydusImageCacheWithConfig(config)

	// Set positive entry
	cache.Set("nydus-image:latest", true)
	// Set negative entry
	cache.Set("regular-image:latest", false)

	// Should be in cache immediately
	_, found := cache.Get("nydus-image:latest")
	if !found {
		t.Error("Expected positive cache hit immediately")
	}

	_, found = cache.Get("regular-image:latest")
	if !found {
		t.Error("Expected negative cache hit immediately")
	}

	// Wait for negative cache to expire
	time.Sleep(1100 * time.Millisecond)

	_, found = cache.Get("regular-image:latest")
	if found {
		t.Error("Expected negative cache miss after TTL")
	}

	// Positive cache should still be valid
	_, found = cache.Get("nydus-image:latest")
	if !found {
		t.Error("Expected positive cache hit after 1s")
	}

	// Wait for positive cache to expire
	time.Sleep(1 * time.Second)

	_, found = cache.Get("nydus-image:latest")
	if found {
		t.Error("Expected positive cache miss after TTL")
	}
}

func TestNydusImageCache_Size(t *testing.T) {
	cache := NewNydusImageCache()

	if cache.Size() != 0 {
		t.Errorf("Expected size=0, got %d", cache.Size())
	}

	cache.Set("image1:latest", true)
	cache.Set("image2:latest", false)
	cache.Set("image3:latest", true)

	if cache.Size() != 3 {
		t.Errorf("Expected size=3, got %d", cache.Size())
	}
}

func TestNydusImageCache_Clear(t *testing.T) {
	cache := NewNydusImageCache()

	cache.Set("image1:latest", true)
	cache.Set("image2:latest", false)

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size=0 after clear, got %d", cache.Size())
	}

	_, found := cache.Get("image1:latest")
	if found {
		t.Error("Expected cache miss after clear")
	}
}

func TestNydusImageCache_Invalidate(t *testing.T) {
	cache := NewNydusImageCache()

	cache.Set("image1:latest", true)
	cache.Set("image2:latest", false)

	// Invalidate one entry
	cache.Invalidate("image1:latest")

	if cache.Size() != 1 {
		t.Errorf("Expected size=1 after invalidate, got %d", cache.Size())
	}

	_, found := cache.Get("image1:latest")
	if found {
		t.Error("Expected cache miss after invalidate")
	}

	// image2 should still exist
	_, found = cache.Get("image2:latest")
	if !found {
		t.Error("Expected cache hit for non-invalidated entry")
	}
}

func TestNydusImageCache_Expiration(t *testing.T) {
	cache := NewNydusImageCacheWithConfig(&CacheConfig{
		PositiveTTL:     200 * time.Millisecond,
		NegativeTTL:     100 * time.Millisecond,
		MaxCacheEntries: 10,
	})

	cache.Set("regular-image:latest", false)
	cache.Set("nydus-image:latest", true)

	time.Sleep(150 * time.Millisecond)

	if _, found := cache.Get("regular-image:latest"); found {
		t.Error("Expected negative entry to expire first")
	}
	if _, found := cache.Get("nydus-image:latest"); !found {
		t.Error("Expected positive entry to still be valid")
	}

	time.Sleep(100 * time.Millisecond)

	if _, found := cache.Get("nydus-image:latest"); found {
		t.Error("Expected positive entry to expire")
	}
}

func TestNydusImageCache_ConcurrentAccess(t *testing.T) {
	cache := NewNydusImageCache()
	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(id int) {
			cache.Set("image", id%2 == 0)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			cache.Get("image")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should not panic
	if cache.Size() == 0 {
		t.Error("Expected at least one entry in cache")
	}
}

func TestNydusImageCache_DefaultConfig(t *testing.T) {
	config := DefaultCacheConfig()

	if config.PositiveTTL != 1*time.Hour {
		t.Errorf("Expected default positive TTL=1h, got %v", config.PositiveTTL)
	}

	if config.NegativeTTL != 5*time.Minute {
		t.Errorf("Expected default negative TTL=5m, got %v", config.NegativeTTL)
	}

	if config.MaxCacheEntries != 1000 {
		t.Errorf("Expected default max entries=1000, got %d", config.MaxCacheEntries)
	}
}
