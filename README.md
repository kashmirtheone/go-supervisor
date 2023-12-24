# go-supervisor

## Introduction
This is not a replacement for proper handling control channels or contexts and cancelation.   
The goal of this package is to provide a controller for handling multiple long-running goroutines gracefully.

## How to install dependency
```bash
me@ubuntu:~$ go get github.com/kashmirtheone/go-supervisor/v2
```

## Concepts

### Task
Used when you want to have a clear control of your long-running goroutine start and stop.  
There are some requirements to keep this task healthy to live into supervisor:
- Start should block while running the process.
- Stop should signal the Start method to return.
- It is not required for stop to only return after start has returned.

You can add specific shutdown policies to this task, or it will inherit from supervisor configuration.

#### Example

```golang
package main

import (
  "context"
  "github.com/kashmirtheone/go-supervisor/v2"
  "github.com/labstack/echo/v4"
)

type Webserver struct {
  *echo.Echo
  Address string
}

// Start starts webservice.
func (e *Webserver) Start() error {
  return e.Echo.Start(e.Address)
}

// Stop stops webservice.
func (e *Webserver) Stop() error {
  return e.Echo.Server.Shutdown(context.Background())
}

func main() {
  ws := &Webserver{
    Echo:    echo.New(),
    Address: ":8081",
  }

  sv := supervisor.New()
  sv.AddTask("webserver", ws)
  sv.Start()
}

```

### Runner
Used when we want to run long-running functions.  
It will deal with context cancellation that will be cancelled when supervisor is ordered to shut down.  
You can add specific shutdown policies to this runner, or it will inherit from supervisor configuration.

#### Example

```golang
package main

import (
  "context"
  "fmt"
  "time"

  "github.com/kashmirtheone/go-supervisor/v2"
)

func RunHealthCheck(ctx context.Context) error {
  for {
    select {
    case <-ctx.Done():
      return nil
    case <-time.Tick(time.Second):
      fmt.Println("I'm alive!")
    }
  }
}

func main() {
  sv := supervisor.New()
  sv.AddRunner("health-check", RunHealthCheck)
  sv.Start()
}

```


### Policies
You can optionally add shutdown policies. This means that you can specify the behaviour long-running goroutines  when throw and error.  
There are 3 types of shutdown policies:
- Ignore  
  - When the process returns an error, supervisor will simply ignore it and log that the process   terminates with error.
- Retry  
  - When the process returns an error, supervisor tries to start it again.  
  - You can also set how many times it will try to start and the time between attempts.
- Shutdown  
  - When the process returns an error, supervisor shuts down.  
  - This means that all live/healthy processes will be terminated gracefully.

#### Example

```golang
func main() {
   ...
   
   sv := supervisor.New(supervisor.WithFailurePolicyShutdown())  
   sv.AddTask("nr-harvester", harvester, supervisor.WithFailurePolicyIgnore())  
   sv.AddTask("webserver", ws, supervisor.WithFailurePolicyShutdown())  
   sv.AddRunner("token-refresher", token.Refresh, supervisor.WithFailurePolicyRetry(10, time.Second))
   
   ...
   
   sv.Start() 
}
```

## TODO
- [ ] Add Ticker/Periodical helper for periodical job execution
  - Add configs like:
    - Idle Time
    - Run Once
- [ ] Add Shutdown Timeout (kill process if it timeouts after stop)
- [ ] Optionally add delay after killing all processes
