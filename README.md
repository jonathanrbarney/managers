# Managers

To include in your repository: `import "github.com/flywinged/managers"`

Managers is a server-Like Local State Management System for Golang. It is intended to be used in instances where many goroutines all need to coordinated access to some shared state. The goal is to make it easier to write these applications by allowing the developer to not need to worry about memory read and write race conditions.

Below are a couple standard usages for the package and how you would implement them. It is comprehensive and includes every binding you should be using while using `managers`.

## Error Logging

Managers by default will log their processing errors to the console. If you'd like to omit this, then include: `managers.LOG_PROCESSING_ERRORS = false`

## Public Methods

Public methods are methods which are accessible globally from anywhere. These kinds of bindings are most useful when building a system which is dynamic enough that you are unable to have a handle on all the managers.

### New Manager

```go
// func NewManager(
//     name string, bufferSize int
// ) (*Manager, error) { ... }

manager, err := managers.NewManager("Example Manager", 128)
```

This will create a new manager object with the specified buffer size.The buffer size is used to determine how many jobs can be queued before the requests will start to be blocked.

### New Request

```go
// func NewRequest(
//     route string, data any
// ) *Request { ... }

request := managers.NewRequest("multiply", 42)
```

This will create a new request object which is assigned to the route `"multiply"` and will send to the processing function the data `42`.

### Requests

```go
// func Send(
//     managerName string, route string, data any
// ) (*Request, error) { ... }

request, err := managers.Send("Example Manager", "multiply", 42)

// func Await(
//     managerName string, route string, data any
// ) (any, error) { ... }

response, err := managers.Await("Example Manager", "multiply", 42)

// func SendRequest(managerName string, request *Request) error { ... }
request := managers.NewRequest("multiply", 42)
err := managers.SendRequest("Example Manager", request)
response, err := managers.AwaitRequest("Example Manager", request)
```

`Send()` and `Await()` will create a request with the data `42` and send it to the manager named `"Example Manager"` if it exists at the route `"multiply"`. `Send()` **IS NOT** blocking and will immediately return with the generated request object when the job has been successfully sent to the manager. `Await()` **IS** blocking and will wait until the manager process the request or there is an error before returning with the relevant response.

`SendRequest()` and `AwaitRequest()` will send a premade request and will emulate the same functions as above.

### Attach/Detach

```go
// func Attach(
//      managerName string,
//      route string,
//      f func(any, any) any
// ) error { ... }

func exampleMultiplication(
    managerState any, request any
) any {
    // Simple example function which will return the
    //  multiplication of whatever value is in the state
    //  with whatever value is in the request.

    // Return the square of the state value
    stateValue, ok := managerState.(*int)
    if !ok {
        return errors.New("Invalid Manager State Type!")
    }

    requestValue, ok := request.(int)
    if !ok {
        return errors.New("Invalid Request Type!")
    }

    return *stateValue * requestValue

}

err := managers.Attach("Example Manager", "multiply", exampleMultiplication)

// func Detach(managerName string, route string) error { ... }
managers.Detach("Example Manager", "multiply")
```

The above will attach the `exampleMultiplication` function to an existing manager `"Example Manager"` at the route `"multiply"`. Whenever a request is sent to this manager with that route, it will be handled by this function. Errors are automatically handled and will be returned accordingly.

The detach function will simply remove the handler from the manager.

### Basic Manager Functions

See Manager Methods heading below for more in depth detail of each.

```go

// func Start(managerName string, managerState any) error { ... }
err := managers.Start("Example Manager", nil) // Start the "Example Manager" with a nil state

// func getManager(managerName string) (*Manager, error) { ... }
manager, err := managers.getManager("Example Manager") // Get a reference to the "Example Manager"

// func Kill(managerName string) error { ... }
err := managers.Kill("Example Manager") // Kill the "Example Manager"

// func Remove(managerName string) error { ... }
err := managers.Remove("Example Manager") // Remove all internal references to "Example Manager"

// func KillAndRemove(managerName string) error { ... }
err := managers.KillAndRemove("Example Manager") // Stop the "Example Manager" and then remove all internal references to it.
```

## Manager Methods

### Start

```go
// func (manager *Manager) Start(managerState any) { ... }

manager, err := managers.NewManager("Example Manager", 128)

state = 42
go manager.Start( &state )
```

The above will start a manager with the state `42`. In general, the state passed in should be of the pointer type so that data can be updated by the internal routes. You can leave this nil if the manager state is always accessible by the bound functions. See example above for some different use cases. This function is **BLOCKING**. If you want it to run in the background, detach it.

### Attach and Detach

```go
// func (manager *Manager) Attach(
//     route string,
//     function func(managerState any, request any) any
// ) { ... }

func exampleMultiplication(
    managerState any, request any
) any {
    // Simple example function which will return the
    //  multiplication of whatever value is in the state
    //  with whatever value is in the request.

    // Return the square of the state value
    stateValue, ok := managerState.(*int)
    if !ok {
        return errors.New("Invalid Manager State Type!")
    }

    requestValue, ok := request.(int)
    if !ok {
        return errors.New("Invalid Request Type!")
    }

    return *stateValue * requestValue

}

manager, err := managers.NewManager("Example Manager", 128)
go manager.Start(42)

manager.Attach("multiply", exampleMultiplication)

// func (manager *Manager) Detach(route string) { ... }
manager.Detach("multiply")
```

The above will attach the `exampleMultiplication` function to an existing manager at the route `"multiply"`. Whenever a request is sent to this manager with that route, it will be handled by this function. Errors are automatically handled and will be returned accordingly.

Detach will simply remove the attached handler.

### Request Methods

```go

manager, err := managers.NewManager("Example Manager", 128)
go manager.Start(42)

// func (manager *Manager) Send(route string, data any) *Request { ... }
request := manager.Send("multiply", 3)

// func (manager *Manager) Await(route string, data any) (any, error) { ... }
response, err := manager.Await("multiply", 3)

// func (manager *Manager) SendRequest(request *Request) { ... }
request := managers.NewRequest("multiply", 42)
manager.SendRequest(request)
response := manager.AwaitRequest(request)
```

`Send()` and `Await()` will send a job to the `"multiply"` route with a value of `42`. The `Send()` route is not blocking and will not receive a response, the `Await()` function is blocking and will receive a response.

`SendRequest()` and `AwaitRequest()` will send a premade request to the manager and then process in the expected way.

### Control Methods

```go

manager, err := managers.NewManager("Example Manager", 128)
go manager.Start(42)

// func (manager *Manager) IsRunning() bool { ... }
// Returns true if the manager is running. You should
//  NOT access manager.Running, as you could introduce race
//  conditions. Check if a manager is running through this method.
running := manager.IsRunning()

// func (manager *Manager) Kill() error { ... }
err := manager.Kill() // Stops the manager from processing.

// func (manager *Manager) Remove() error { ... }
err := manager.Remove() // Removes the manager from memory

// func (manager *Manager) KillAndRemove() error { ... }
err := manager.KillAndRemove() // Does the above to function in sequence.
```

## Requests Methods

Collection of all the methods you can make on the request object.

### Send

```go
// func (request *Request) Send(managerName string) error { ... }

request := managers.NewRequest("multiply", 42)
err := request.Send("Example Manager")
```

This will send a request to the Example Manager with the data `42`. It is not blocking.

### SendManager

```go
// func (request *Request) SendManager(manager *Manager) { ... }

manager, err := managers.NewManager("Example Manager", 128)
go manager.Start(42)

request := managers.NewRequest("multiply", 42)

request.SendManager(manager)
```

The above will send a job to a predefined Manager. It is this same as `Send()` but doesn't lookup a manager by name.

### Await

```go
// func (request *Request) Await(managerName string) (any, error) { ... }

request := managers.NewRequest("multiply", 42)
response, err := request.Await("Example Manager)
```

This will send a request to the Example Manager with the data `42`. It is blocking and will wait for the process to complete.

### AwaitManager

```go
// func (request *Request) AwaitManager(manager *Manager) (any, error) { ... }

manager, err := managers.NewManager("Example Manager", 128)
go manager.Start(42)

request := managers.NewRequest("multiply", 42)
response err := request.AwaitManager(manager)
```

The above will await a job at a predefined manager. It i the same as `AwaitManager()` but doesn't lookup a manager by name.

### Wait

```go
// func (request *Request) Wait() (any, error) { ... }

request := managers.NewRequest("multiply", 42)
err := request.Send("Example Manager")

// Do some other processing
response, err := request.Wait()

```

The above will wait for a job that has been sent to finally be processed. You can use this to do parallel processing while awaiting a return.
Wait is called by `Await()` as well. The important thing to note, is `Wait()` is what every function uses to fetch responses. If you have nested request objects for whatever reason, this will automatically follow the nested pattern and return the final result.

### Has Data

```go
// func (request *Request) HasData() bool { ... }
request := managers.NewRequest("multiply", 42)
err := request.Send("Example Manager")

hasData := request.HasData()
```

The above will check to see if a request has data yet or not.

## Structs

There are two main structs provided in this package, `Request` and `Manager`. The `ManagerFunction` is just a specified function type which is handled by the managers.

### Manager

The manager handles basically everything. A manager is responsible for processing requests as they come in. In general, there is no need to ever access any of the values within the manager as there is a Method (Either Public or Private) which will return to you all the information you need in order to run the manager.

```go
type Manager struct {
    Name string
    requests chan *Request
    running bool
    functions map[string]func(managerState any, request any) any
    stateLock sync.Mutex
}
```

#### ManagerFunction

A Manager function is simply a function of the following type:
`func(managerState any, request any) any`
These functions can be attached to managers so that the managers can process a range of different tasks. Think of them as API Routes.

### Request

The request object is very simple. It has a specified route it's supposed to be sent to, it has data which will end up as the argument to the specified route, and it has a response channel which is responsible for awaiting a response. The `Response` object is an internal object (Although you're welcome to read from it if you would like to implement a different methodology.)

```go
type Request struct {
    Route string
    Data any
    Response chan Response
}
```

```go
type Response struct {
    Data  any
    Error error
}
```
