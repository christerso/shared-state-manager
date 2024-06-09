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
