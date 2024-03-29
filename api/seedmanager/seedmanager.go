//
// Copyright 2019 Insolar Technologies GmbH
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
//

package seedmanager

import (
	"sync"
	"time"

	"github.com/insolar/insolar/insolar"
)

// Expiration represents time of expiration
type Expiration = int64

// DefaultTTL is default time period for deleting expired seeds
const DefaultTTL = 5 * time.Second

// DefaultCleanPeriod default time period for launching cleaning goroutine
const DefaultCleanPeriod = 5 * time.Second

type storedSeed struct {
	expiration Expiration
	pulse      insolar.PulseNumber
}

// SeedManager manages working with seed pool
// It's thread safe
type SeedManager struct {
	mutex    sync.RWMutex
	seedPool map[Seed]storedSeed
	ttl      time.Duration
	stopped  bool
}

// New creates new seed manager with default params
func New() *SeedManager {
	return NewSpecified(DefaultTTL, DefaultCleanPeriod)
}

// NewSpecified creates new seed manager with custom params
func NewSpecified(ttl time.Duration, cleanPeriod time.Duration) *SeedManager {
	sm := SeedManager{seedPool: make(map[Seed]storedSeed), ttl: ttl, stopped: false}
	go func() {
		for range time.Tick(cleanPeriod) {
			sm.deleteExpired()
			sm.mutex.RLock()
			if sm.stopped {
				break
			}
			sm.mutex.RUnlock()
		}
	}()

	return &sm
}

func (sm *SeedManager) Stop() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.stopped = true
}

// Add adds seed to pool
func (sm *SeedManager) Add(seed Seed, pulse insolar.PulseNumber) {
	expTime := time.Now().Add(sm.ttl).UnixNano()

	sm.mutex.Lock()
	sm.seedPool[seed] = storedSeed{
		expiration: expTime,
		pulse:      pulse,
	}
	sm.mutex.Unlock()

}

func (sm *SeedManager) isExpired(expTime Expiration) bool {
	return expTime < time.Now().UnixNano()
}

// Exists checks whether seed in the pool
func (sm *SeedManager) Pop(seed Seed) (insolar.PulseNumber, bool) {
	sm.mutex.RLock()
	stored, ok := sm.seedPool[seed]
	sm.mutex.RUnlock()

	if ok && !sm.isExpired(stored.expiration) {
		sm.mutex.Lock()
		delete(sm.seedPool, seed)
		sm.mutex.Unlock()
		return stored.pulse, true
	}

	return 0, false
}

func (sm *SeedManager) deleteExpired() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for seed, stored := range sm.seedPool {
		if sm.isExpired(stored.expiration) {
			delete(sm.seedPool, seed)
		}
	}
}

// SeedFromBytes converts slice of bytes to Seed. Returns nil if slice's size is not equal to SeedSize
func SeedFromBytes(slice []byte) *Seed {
	if len(slice) != int(SeedSize) {
		return nil
	}
	var result Seed
	copy(result[:], slice[:SeedSize])
	return &result
}
