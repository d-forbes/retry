# retry

A simple and configurable Go package that provides a mechanism to retry operations with exponential backoff and optional jitter. This package is especially useful when dealing with transient errors during network calls, database queries, or other operations where a retry strategy is beneficial.

## Features

- **Configurable Attempts:** Specify the maximum number of attempts.
- **Exponential Backoff:** Increase the delay between retries using a configurable factor.
- **Jitter Support:** Optionally add randomness to delays to prevent thundering herd issues.
- **Context Support:** Respect context cancellations or timeouts.

## Installation

To install the package, use `go get` with your module path:

```bash
go get github.com/d-forbes/retry
```

## Usage

Below is an example of how to use the `retry` package in a Go program:

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/d-forbes/retry"
)

func main() {
	// Counter to simulate transient failures.
	var attempt int

	// Operation to be retried.
	operation := func() error {
		attempt++
		fmt.Printf("Attempt %d: ", attempt)
		// Simulate failure for the first 3 attempts.
		if attempt < 4 {
			fmt.Println("failed")
			return errors.New("transient error")
		}
		fmt.Println("succeeded")
		return nil
	}

	// Configure the retry behavior.
	cfg := retry.Config{
		Attempts:     5,                     // Total number of attempts.
		InitialDelay: 500 * time.Millisecond, // Delay before the second attempt.
		MaxDelay:     5 * time.Second,        // Maximum allowed delay.
		Factor:       2,                      // Exponential backoff factor.
		Jitter:       true,                   // Enable jitter.
	}

	// Create a context with timeout to avoid indefinite retries.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Execute the operation with retry logic.
	if err := retry.Do(ctx, operation, cfg); err != nil {
		log.Fatalf("Operation failed after retries: %v", err)
	} else {
		log.Println("Operation succeeded!")
	}
}
```

## API Reference
### `Config`
The `Config` struct defines the retry parameters:

| Field          | Type            | Description                                                                                                                                         |
| -------------- | --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Attempts`     | `int`           | Maximum number of attempts (including the initial try).                                                                                           |
| `InitialDelay` | `time.Duration` | The initial delay before retrying the operation.                                                                                                  |
| `MaxDelay`     | `time.Duration` | The maximum delay between retries. If set to zero, there is no limit.                                                                               |
| `Factor`       | `float64`       | Multiplier used to increase the delay after each attempt (e.g., a factor of 2 doubles the delay each time).                                          |
| `Jitter`       | `bool`          | If set to `true`, adds a random element to the delay (up to the current delay value) to mitigate thundering herd problems.                            |

### `Do`
```go
func Do(ctx context.Context, op func() error, cfg Config) error
```

- **Parameters**:
    - `ctx`: The context to control cancellation or timeout.
    - `op`: The operation function that returns an error. If the function returns nil, it is considered a success.
    - `cfg`: The retry configuration.
- **Returns**: `nil` on success or the last error encountered after all retries have been exhausted.

## Contributing

Contributions are welcome! Please feel free to open issues or submit pull requests if you have improvements or additional features.