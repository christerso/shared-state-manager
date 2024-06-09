
# SharedStateManager

`SharedStateManager` is a Go package that provides a thread-safe, in-memory key-value store with support for conditional notifications and timed expiration of keys. It is designed for efficient internal messaging within your application.

## Features

- Thread-safe key-value storage
- Conditional notifications for subscribers
- Timed value expiration with notifications
- Simple API for setting, getting, and deleting values

## Installation

```bash
go get github.com/christerso/shared-state-manager
```

## Usage

### Import the package

```go
import "github.com/christerso/shared-state-manager/managers"
```

### Example

```go
package main

import (
    "fmt"
    "time"

    "github.com/christerso/shared-state-manager/managers"
)

func main() {
    ssm := managers.NewSharedStateManager()

    // Define the event handler
    eventHandler := func(data interface{}) {
        switch v := data.(type) {
        case managers.ExpirationNotification:
            fmt.Println("Event expired for key:", v.Key)
        default:
            fmt.Println("Received:", data)
        }
    }

    // Start subscription with conditional notifications enabled
    managers.StartSubscription(ssm, "exampleEvent", eventHandler, true)

    // Set a value
    ssm.Set("exampleEvent", "Hello, EventBus!")

    // Try setting the same value again (should not notify)
    ssm.Set("exampleEvent", "Hello, EventBus!")

    // Change the value (should notify)
    ssm.Set("exampleEvent", "New Value")

    // Start another subscription without conditional notifications
    managers.StartSubscription(ssm, "exampleEvent", func(data interface{}) {
        switch v := data.(type) {
        case managers.ExpirationNotification:
            fmt.Println("Always received expiration for key:", v.Key)
        default:
            fmt.Println("Always received:", data)
        }
    }, false)

    // Set a value again (should notify both subscribers)
    ssm.Set("exampleEvent", "Another Value")

    // Set a timed value without a subscriber
    ssm.SetWithTimeout("timedKey", "This will expire", 5*time.Second)

    // Wait to observe the expiration
    time.Sleep(6 * time.Second)
}
```

## API

### `NewSharedStateManager() *SharedStateManager`

Creates a new instance of `SharedStateManager`.

### `Set(key string, value interface{})`

Sets a value for a given key.

### `SetWithTimeout(key string, value interface{}, duration time.Duration)`

Sets a value for a given key with an expiration time.

### `Get(key string) (interface{}, bool)`

Retrieves a value for a given key.

### `GetString(key string) (string, bool)`

Retrieves a string value for a given key.

### `GetStruct(key string) (interface{}, bool)`

Retrieves a struct value for a given key.

### `Delete(key string)`

Removes a key-value pair.

### `Subscribe(key string, ch chan interface{}, conditional bool)`

Adds a subscriber for a specific key with conditional notifications.

### `StartSubscription(ssm *SharedStateManager, key string, handler func(interface{}), conditional bool)`

Helper function to start a subscription goroutine with a handler.

## License

This project is licensed under the MIT License.
