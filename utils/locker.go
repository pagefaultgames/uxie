// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"sync"
	"time"
)

// A Locker represents a simple concurrent lock, used to ensure exclusive access to a given resource.
// The keys should represent the resources being locked (e.g. rows in a database), while the values can hold information required to identify the lock's owner.
//
// Unlike standard synchronization primitives, the locker is designed to be non-blocking.
type Locker[K, V comparable] struct {
	locks map[K]lockerInfo[V]
	mu    *sync.Mutex
}

// A lockerInfo represents the value stored for a given lock.
type lockerInfo[V comparable] struct {
	value   V
	timeout *time.Timer
}

// NewLocker creates a new Locker instance.
func NewLocker[K, V comparable]() *Locker[K, V] {
	return &Locker[K, V]{
		locks: make(map[K]lockerInfo[V]),
	}
}

// Lock attempts to lock the given key, associating it with the provided value.
// It returns whether the lock was successfully added.
func (l *Locker[K, V]) Lock(key K, value V) (locked bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.locks[key]; ok {
		return false
	}
	l.locks[key] = lockerInfo[V]{value: value}
	return true
}

// LockWithTimeout functions like [Lock], but will ensure the lock is released at or before the provided duration elapses.
//
// If the key is already locked, this will function identically to [Lock] and return false.
func (l *Locker[K, V]) LockWithTimeout(
	key K,
	value V,
	timeout time.Duration,
) (ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.locks[key]; ok {
		return false
	}

	l.locks[key] = lockerInfo[V]{
		value:   value,
		timeout: time.AfterFunc(timeout, func() { l.Unlock(key) }),
	}
	return true
}

// Unlock unlocks the given key, allowing others to lock it again.
// If the key is not currently locked, the call to Unlock is a no-op.
func (l *Locker[K, V]) Unlock(key K) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if info, ok := l.locks[key]; ok && info.timeout != nil {
		info.timeout.Stop()
	}
	delete(l.locks, key)
}

// IsLockedBy returns the value that is locking the given key, if present.
// If the key is not currently locked, it will return the zero value of V and false.
func (l *Locker[K, V]) GetLocker(key K) (existing V, locked bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.locks[key]
	return e.value, ok
}
