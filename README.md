DI Container
============

Dependency injection is one form of the broader technique of inversion
of control. It is used to increase modularity of the program and make it
extensible.

Before using this repository, you should understand what does
mean **go way**, but you want to reduce the boilerplate and add some
brain to your code.

## Contents

- [Documentation](https://github.com/goava/di/blob/master/README.md#documentation)
- [Install](https://github.com/goava/di/blob/master/README.md#installing)
- [Tutorial](https://github.com/goava/di/blob/master/README.md#tutorial)
  - [Provide](https://github.com/goava/di/blob/master/README.md#provide)
  - [Compile](https://github.com/goava/di/blob/master/README.md#compile)
  - [Resolve](https://github.com/goava/di/blob/master/README.md#resolve)
  - [Invoke](https://github.com/goava/di/blob/master/README.md#invoke)
  - [Lazy-loading](https://github.com/goava/di/blob/master/README.md#lazy-loading)
  - [Interfaces](https://github.com/goava/di/blob/master/README.md#interfaces)
  - [Groups](https://github.com/goava/di/blob/master/README.md#groups)
- [Advanced features](https://github.com/goava/di/blob/master/README.md#advanced-features)
  - [Named definitions](https://github.com/goava/di/blob/master/README.md#named-definitions)
  - [Optional parameters](https://github.com/goava/di/blob/master/README.md#optional-parameters)
  - [Prototypes](https://github.com/goava/di/blob/master/README.md#prototypes)
  - [Cleanup](https://github.com/goava/di/blob/master/README.md#cleanup)


## Documentation

You can use standard [pkg.go.dev](https://pkg.go.dev/github.com/goava/di) and inline code
comments or if you does not have experience with auto-wiring libraries
as [google/wire](https://github.com/google/wire),
[uber-go/dig](https://github.com/uber-go/dig) or another start with
[tutorial](https://github.com/goava/di/blob/master/README.md#tutorial).

## Install

```shell
go get github.com/goava/di
```

## Tutorial

Let's learn to use di by example. We will code a simple application
that processes HTTP requests.

The full tutorial code is available [here](./_tutorial/main.go)

### Provide

To start, we will need to create two fundamental types: `http.Server`
and `http.ServeMux`. Let's create a simple constructors that initialize
them:

```go
// NewServer creates a http server with provided mux as handler.
func NewServer(mux *http.ServeMux) *http.Server {
	return &http.Server{
		Handler: mux,
	}
}

// NewServeMux creates a new http serve mux.
func NewServeMux() *http.ServeMux {
	return &http.ServeMux{}
}
```

> Supported constructor signature:
>
> ```go
> func([dep1, dep2, depN]) (result, [cleanup, error])
> ```

Now let's teach a container to build these types:

```go
// build container
c := container.New(
	di.Provide(NewServer),
	di.Provide(NewServeMux),
)
```

The function `di.New()` creates new container with provided options.

### Compile

`Compile` compiles the container: parse constructors, build
dependency graph, check that it is acyclic.

```go
// compile dependency graph
if err := c.Compile(); err != nil {
	// handle error
}
```

### Resolve

We can extract the built server from the container. For this, define the
variable of extracted type and pass variable pointer to `Extract`
function.

> If extracted type not found or the process of building instance cause
> error, `Extract` return error.

If no error occurred, we can use the variable as if we had built it
yourself.

```go
// declare type variable
var server *http.Server
// extracting
err := container.Resolve(&server)
if err != nil {
	// check extraction error
}

server.ListenAndServe()
```

> Note that by default, the container creates instances as a singleton.
> But you can change this behaviour. See [Prototypes](https://github.com/goava/di/blob/master/README.md#prototypes).

### Invoke

As an alternative to extraction we can use `Invoke()` function of `Container`. It
resolves function dependencies and call the function. Invoke function
may return optional error.

```go
// StartServer starts the server.
func StartServer(server *http.Server) error {
    return server.ListenAndServe()
}

if err := container.Invoke(StartServer); err != nil {
	// handle error
}
```

Also you can use `di.Invoke()` container options for call some initialization code.

```go
container := di.New(
	di.Provide(NewServer),
	di.Invoke(StartServer),
)

if err := container.Compile(); err != nil {
	// handle error
}
```

Container run all invoke functions on compile stage. If one of them
failed (return error), compile cause error.

### Lazy-loading

Result dependencies will be lazy-loaded. If no one requires a type from
the container it will not be constructed.

### Interfaces

Inject make possible to provide implementation as an interface.

```go
// NewServer creates a http server with provided mux as handler.
func NewServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler: handler,
	}
}
```

For a container to know that as an implementation of `http.Handler` is
necessary to use, we use the option `di.As()`. The arguments of this
option must be a pointer(s) to an interface like `new(Endpoint)`.

> This syntax may seem strange, but I have not found a better way to
> specify the interface.

Updated container initialization code:

```go
container := di.New(
	// provide http server
	di.Provide(NewServer),
	// provide http serve mux as http.Handler interface
	di.Provide(NewServeMux, di.As(new(http.Handler)))
)
```

Now container uses provide `*http.ServeMux` as `http.Handler` in server
constructor. Using interfaces contributes to writing more testable code.

### Groups

Container automatically groups all implementations of interface to
`[]<interface>` group. For example, provide with
`di.As(new(http.Handler)` automatically creates a group
`[]http.Handler`.

Let's add some http controllers using this feature. Controllers have
typical behavior. It is registering routes. At first, will create an
interface for it.

```go
// Controller is an interface that can register its routes.
type Controller interface {
	RegisterRoutes(mux *http.ServeMux)
}
```

Now we will write controllers and implement `Controller` interface.

##### OrderController

```go
// OrderController is a http controller for orders.
type OrderController struct {}

// NewOrderController creates a auth http controller.
func NewOrderController() *OrderController {
	return &OrderController{}
}

// RegisterRoutes is a Controller interface implementation.
func (a *OrderController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/orders", a.RetrieveOrders)
}

// Retrieve loads orders and writes it to the writer.
func (a *OrderController) RetrieveOrders(writer http.ResponseWriter, request *http.Request) {
	// implementation
}
```

##### UserController

```go
// UserController is a http endpoint for a user.
type UserController struct {}

// NewUserController creates a user http endpoint.
func NewUserController() *UserController {
	return &UserController{}
}

// RegisterRoutes is a Controller interface implementation.
func (e *UserController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/users", e.RetrieveUsers)
}

// Retrieve loads users and writes it using the writer.
func (e *UserController) RetrieveUsers(writer http.ResponseWriter, request *http.Request) {
    // implementation
}
```

Just like in the example with interfaces, we will use `di.As()`
provide option.

```go
container := di.New(
	di.Provide(NewServer),        // provide http server
	di.Provide(NewServeMux),       // provide http serve mux
	// endpoints
	di.Provide(NewOrderController, di.As(new(Controller))),  // provide order controller
	di.Provide(NewUserController, di.As(new(Controller))),  // provide user controller
)
```

Now, we can use `[]Controller` group in our mux. See updated code:

```go
// NewServeMux creates a new http serve mux.
func NewServeMux(controllers []Controller) *http.ServeMux {
	mux := &http.ServeMux{}

	for _, controller := range controllers {
		controller.RegisterRoutes(mux)
	}

	return mux
}
```

## Advanced features

### Named definitions

In some cases you have more than one instance of one type. For example
two instances of database: master - for writing, slave - for reading.

First way is a wrapping types:

```go
// MasterDatabase provide write database access.
type MasterDatabase struct {
	*Database
}

// SlaveDatabase provide read database access.
type SlaveDatabase struct {
	*Database
}
```

Second way is a using named definitions with `di.WithName()` provide
option:

```go
// provide master database
di.Provide(NewMasterDatabase, di.WithName("master"))
// provide slave database
di.Provide(NewSlaveDatabase, di.WithName("slave"))
```

If you need to extract it from container use `di.Name()` extract
option.

```go
var db *Database
container.Resolve(&db, di.Name("master"))
```

If you need to provide named definition in other constructor use
`di.Parameter` with embedding.

```go
// ServiceParameters
type ServiceParameters struct {
	di.Parameter
	
	// use `di` tag for the container to know that field need to be injected.
	MasterDatabase *Database `di:"master"`
	SlaveDatabase *Database  `di:"slave"`
}

// NewService creates new service with provided parameters.
func NewService(parameters ServiceParameters) *Service {
	return &Service{
		MasterDatabase:  parameters.MasterDatabase,
		SlaveDatabase: parameters.SlaveDatabase,
	}
}
```

### Optional parameters

Also `di.Parameter` provide ability to skip dependency if it not exists
in container.

```go
// ServiceParameter
type ServiceParameter struct {
	di.Parameter
	
	Logger *Logger `di:"optional"`
}
```

> Constructors that declare dependencies as optional must handle the
> case of those dependencies being absent.

You can use naming and optional together.

```go
// ServiceParameter
type ServiceParameter struct {
	di.Parameter
	
	StdOutLogger *Logger `di:"stdout"`
	FileLogger   *Logger `di:"file,optional"`
}
```

### Prototypes

If you want to create a new instance on each extraction use
`di.Prototype()` provide option.

```go
di.Provide(NewRequestContext, di.Prototype())
```

> todo: real use case

### Cleanup

If a provider creates a value that needs to be cleaned up, then it can
return a closure to clean up the resource.

```go
func NewFile(log Logger, path Path) (*os.File, func(), error) {
    f, err := os.Open(string(path))
    if err != nil {
        return nil, nil, err
    }
    cleanup := func() {
        if err := f.Close(); err != nil {
            log.Log(err)
        }
    }
    return f, cleanup, nil
}
```

After `container.Cleanup()` call, it iterate over instances and call
cleanup function if it exists.

```go
container := di.New(
	// ...
    di.Provide(NewFile),
)

// do something
container.Cleanup() // file was closed
```