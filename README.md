# gobatch

`gobatch` is a Go library crafted for efficient and robust concurrent asynchronous batch processing. It's optimized for I/O-based operations like HTTP requests but versatile enough to handle a wide range of batch processing tasks.

## What is gobatch?

In scenarios where you need to process a large number of tasks - like making billions of HTTP requests - traditional sequential processing methods fall short, both in terms of time and efficiency. `gobatch` steps in to dramatically cut down processing time by leveraging Go's powerful concurrency model, allowing tasks to be processed in batches across multiple goroutines.

batch processing logic so that it respects the overall rate limit across all goroutines

Update the batch processing logic to acquire capacity from the rate limiter before executing each operation. This ensures that the overall rate limit is respected across all batches and goroutines.


## Features

- **Concurrent Processing**: Utilizes Go's goroutines for concurrent execution of batch tasks.
- **Customizable Batch Size**: Control the number of tasks processed concurrently.
- **Error Handling and Retries**: Robust error handling with configurable retry logic.
- **Timeouts and Delays**: Set timeouts for tasks and delays between retries.
- **Resource Management**: Control over resource utilization like CPU cores and number of goroutines.
- **Progress Reporting**: Callbacks for tracking the progress of batch processing.
- **Extensibility**: Easily extendable for various types of batch operations.

## Installation

Install `gobatch` by running:

```shell
go get -u github.com/yourusername/gobatch
```

## Usage

Here's a simple example of how to use `gobatch`:

```go
package main

import (
    "github.com/yourusername/gobatch"
    "context"
    // other imports
)

func main() {
    // Define your batch operation
    operation := func(ctx context.Context, data interface{}) (interface{}, error) {
        // Your batch operation logic here
    }

    // Create a batch of data to process
    dataBatch := []interface{}{/* your data items */}

    // Execute batch operation
    results, err := gobatch.ExecuteBatchAsync(context.Background(), operation, dataBatch, gobatch.WithMaxGoroutines(100))
    if err != nil {
        // Handle error
    }

    // Handle results
    for result := range results {
        // Process each result
    }
}
```

## Configuration Options

`gobatch` provides several options to fine-tune your batch processing:

- `WithMaxGoroutines(int)`: Sets the maximum number of goroutines.
- `WithTimeout(time.Duration)`: Specifies the timeout for each operation.
- `WithRetryDelay(time.Duration)`: Sets the delay between retries.
- `WithCores(int)`: Determines the number of CPU cores to use.
- ... and more.

## Contributing

Contributions to `gobatch` are welcome! Whether it's bug reports, feature requests, or code contributions, please feel free to contribute.

To contribute:

1. Fork the repository.
2. Create a new branch for your feature or fix.
3. Write and test your code.
4. Submit a pull request.

## License

`gobatch` is released under the [MIT License](LICENSE).

---

Feel free to adjust the README as needed to better fit the specifics of your library, its dependencies, and your GitHub repository structure.