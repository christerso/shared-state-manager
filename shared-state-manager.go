package managers

import (
	"sync"
	"time"
)

// SharedStateManager is a Go package for managing shared state across different parts of an application.
// It provides a thread-safe, in-memory key-value store with support for conditional notifications and timed expiration of keys.
// This can be particularly useful for building event-driven systems, caching mechanisms, and coordinating state between goroutines.
//
// Features:
// - Thread-safe key-value storage
// - Conditional notifications for subscribers
// - Timed value expiration with notifications
// - Simple API for setting, getting, and deleting values
//
// Example use cases:
// - Real-time event systems where components need to react to state changes
// - Caching system where entries automatically expire after a certain time
// - Coordination of shared state between multiple goroutines
//
// Version: 1.0.0

// ExpirationNotification represents a notification that a value has expired.
type ExpirationNotification struct {
	Key string
}

// Subscription holds the channel and conditional flag for each subscriber.
type Subscription struct {
	Channel                  chan interface{}
	ConditionalNotifications bool
}

// SharedStateManager manages states with string keys and any type of value.
type SharedStateManager struct {
	stateMap    map[string]interface{}
	timers      map[string]*time.Timer
	subscribers map[string][]Subscription
	mu          sync.RWMutex
}

// NewSharedStateManager creates a new instance of SharedStateManager.
func NewSharedStateManager() *SharedStateManager {
	return &SharedStateManager{
		stateMap:    make(map[string]interface{}),
		timers:      make(map[string]*time.Timer),
		subscribers: make(map[string][]Subscription),
	}
}

// Set sets a value for a given key.
func (ssm *SharedStateManager) Set(key string, value interface{}) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()
	oldValue, exists := ssm.stateMap[key]
	ssm.stateMap[key] = value
	ssm.notifySubscribers(key, value, exists, oldValue)
}

// SetWithTimeout sets a value for a given key with an expiration time.
func (ssm *SharedStateManager) SetWithTimeout(key string, value interface{}, duration time.Duration) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()
	ssm.stateMap[key] = value
	ssm.notifySubscribers(key, value, false, nil)

	// Cancel any existing timer for the key
	if timer, exists := ssm.timers[key]; exists {
		timer.Stop()
	}
	// Set a new timer for the key
	timer := time.AfterFunc(duration, func() {
		ssm.expireKey(key)
	})
	ssm.timers[key] = timer
}

// notifySubscribers notifies all subscribers of a key's value change.
func (ssm *SharedStateManager) notifySubscribers(key string, value interface{}, exists bool, oldValue interface{}) {
	for _, sub := range ssm.subscribers[key] {
		if !sub.ConditionalNotifications || !exists || oldValue != value {
			sub.Channel <- value
		}
	}
}

// expireKey handles the expiration of a key.
func (ssm *SharedStateManager) expireKey(key string) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()
	delete(ssm.stateMap, key)
	if timer, exists := ssm.timers[key]; exists {
		timer.Stop()
		delete(ssm.timers, key)
	}
	expirationNotification := ExpirationNotification{Key: key}
	for _, sub := range ssm.subscribers[key] {
		sub.Channel <- expirationNotification
	}
}

// Get retrieves a value for a given key.
func (ssm *SharedStateManager) Get(key string) (interface{}, bool) {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()
	value, exists := ssm.stateMap[key]
	return value, exists
}

// GetString retrieves a string value for a given key.
func (ssm *SharedStateManager) GetString(key string) (string, bool) {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()
	value, exists := ssm.stateMap[key]
	if !exists {
		return "", false
	}
	strValue, ok := value.(string)
	return strValue, ok
}

// GetStruct retrieves a struct value for a given key.
func (ssm *SharedStateManager) GetStruct(key string) (interface{}, bool) {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()
	value, exists := ssm.stateMap[key]
	if !exists {
		return nil, false
	}
	return value, true
}

// Delete removes a key-value pair.
func (ssm *SharedStateManager) Delete(key string) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()
	delete(ssm.stateMap, key)
	if timer, exists := ssm.timers[key]; exists {
		timer.Stop()
		delete(ssm.timers, key)
	}
	expirationNotification := ExpirationNotification{Key: key}
	for _, sub := range ssm.subscribers[key] {
		sub.Channel <- expirationNotification
	}
}

// Subscribe adds a subscriber for a specific key.
func (ssm *SharedStateManager) Subscribe(key string, ch chan interface{}, conditional bool) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()
	ssm.subscribers[key] = append(ssm.subscribers[key], Subscription{
		Channel:                  ch,
		ConditionalNotifications: conditional,
	})
}

// StartSubscription starts a subscription goroutine with a handler.
func StartSubscription(ssm *SharedStateManager, key string, handler func(interface{}), conditional bool) {
	ch := make(chan interface{})
	ssm.Subscribe(key, ch, conditional)
	go func() {
		for data := range ch {
			handler(data)
		}
	}()
}
