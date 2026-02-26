https://www.youtube.com/live/vzoaBfxbrdo?si=sgp1V6Zv00jiXOGa

# High Level Design
We will be building a chat app.
- We will be building a go service called client access point (CAP).
- We will have clients such as browser, applications written in javascript etc. We want to support different types of clients.
- They are going to come in and establish socket connections with the CAP.

We could decide to use our own socket protocol. But that would have been important if we were building close sort of system. It would be more fun to build open system. To do that we are going to use old school tech and we will be using WebSockets. 

**WebSockets** are a protocol that allows us to have a persistent connection between the client and the server. It is a **full duplex communication channel** that operates over a single TCP connection. It is designed to be implemented in web browsers and web servers, but it can be used by any client or server application.

If we use WebSockets, we already have some good go support like with Gorilla packaging and JavaScript already have support for WebSockets. So we can use that to our advantage and we can have a lot of different clients connecting to our CAP.

We actullay need back and forth communication between the client and the server. 

If CAP is the only service that we are going to build it's going to be fairly simple.
- We need to keep some sort of map of socket to the user that is connected to that socket.
  - We could ask for name or userid.
  - We can go really crazy and use Web3 stuff where you Sign a message with your private key and we can verify that you are who you say you are.
- Then we need some sort of router, given a message we can lookup in and then send the message out.

This is fairly easy but we don't want to have just one CAP and this is where things get interesting.

If we want it to be a distributed system, we are not going to ask somebody in Australia to connect to CAP server running in US. This will be phase 2 of our ptoject.

How all these CAP servers know about each other. How does one know where to send the message. We need some sort of discovery mechanism. We need some sort of way to route messages between different CAP servers.

The way we will handle is we will not be coding any of this because it's already been coded before. There is a bunch of protocols out there that sort of do this.

It would be interesting to bring in NATS here. NATS is like persisted message bus.
CAP servers can connect to NATS and we could leverage Pub/Sub mechanism to send messages between different CAP servers. 

Nice thing about NATS is that they have build redundancy and they have build in clustering. So we can have multiple NATS servers running and they will be able to talk to each other and they will be able to route messages between each other.

They have already solved redundancy problems etc, so we don't need to solve those problems. We can just leverage NATS and we can focus on building the CAP servers and we can focus on building the clients. We will also setup and code it as part of Phase 2 of our project.

We are going to solve this problem without NATS first. One refactoring eventually will have to be where we will make the decision, do we want to check our map first to see if that user already exists to stay off the NATS or do we want to just send everything to NATS and let NATS figure out and send back to us. There is no intelligece in that, but then we can argue why to throw the code that we already build. Will come back to it later.

# Phase 1
- We will build a single CAP server that can handle multiple clients.

Start thinking about what are the walls infront of us, what are the things that we don't know. The big
thing that we don't know right now is what information we have once the websocket is connected to be able to create that map and how to do that.

So let's start building our service, start with APP layer which will be the handler for the websocket connection. 

We know that we need API layer and APP layer and foundation layer (logger) that we will be stealing from our [service repo](https://github.com/ardanlabs/service).

This is boiler plate of our service.

```go
func main() {
	var log *logger.Logger

	traceIDFn := func(ctx context.Context) string {
		return "" // TODO: Need trace IDs
	}

	log = logger.New(os.Stdout, logger.LevelInfo, "SALES", traceIDFn)

	// -------------------------------------------------------------------------

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	return nil
}
```

In the run we will write startup/shutdown related code. run can return an error and if it does we will log it and exit with non zero code. If it doesn't return an error, we will just keep running until we get a signal to shutdown.

log.Fatal also logs and exits with status code 1 but we are being more explicit here.

We also like to know how many operating system threads we are sort of running with especially when we are running in Kubernetes environment. Now with this only we can setup our makefile.

```
chat-run:
	go run chat/api/services/cap/main.go | go run chat/api/tooling/logfmt/main.go
```

There is problem with using pipes in makefile. To understand that let's add shutdown code in our service.

```go
// -------------------------------------------------------------------------
shutdown := make(chan os.Signal, 1)
signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
<-shutdown
```

This is going to keep the service running until we hit CTRL+C.

Problem with using pipes in makefile is that when we hit CTRL+C, this signal will not propagate correctly through the pipe and we won't be able to shutdown our service gracefully. 

So we are going to add below code in our logfmt tool to handle this.

```go
signal.Notify(shutdown, syscall.SIGINT)
syscall.Kill(os.Getppid(), syscall.SIGINT)
```

So this code once get's the signal it will send the signal to the parent process which is the makefile and then makefile will be able to shutdown our service gracefully.

We are now able to start and stop our service gracefully. That's the first thing.

If we were going to put this in production, almost at this point we should start setting up staging env, start creating images, start setting up deployment now. Because if you wait till the end to deploy these things, you might have done something that's not deployable or used a tech that you shouldn't have been using.

Next we are trying to get oursleves in a position where we have a Handler that can accept a HTTP request and convert it to a websocket. We are going to use our web frametwork in the service project.

Good thing about the context API is, if something isn't there in the context we don't have to fail and shutdown, you can return a default value.

```go
var log *logger.Logger

traceIDFn := func(ctx context.Context) string {
    return web.GetTraceID(ctx)
}

log = logger.New(os.Stdout, logger.LevelInfo, "CAP", traceIDFn)
```

Everytime it goes to write a log, it will call the traceIDFn and it will get the trace ID from the context and it will include that in the log. So we will have trace IDs in our logs which is going to be very helpful when we are trying to debug issues in production. If it's not there, it will just  logs without trace ID with all zeros.

Now we have web framework in place, we are going to add App layer, routes and wire them up. In app layer we are going to have two things.
- domain: Will contain our routes 
- sdk: packages that will be used by app layer domain, providing support.
  - If it was needed by both API and APP, we would have put it in foundation layer but since it's only needed by APP layer, we will put it in SDK.
  - We should keep the packages as close as possible to the code that is using it.
  - We will copy err, mux, mid from service repo.

This err package is going to be used specifically by App layer. Business layer and foundation layers would return errors in a very traditional way, whatever they want to do. But we have to send error responses back to the client, we have to implement HandlerFunc interface in web.go if we want to send
anything back to the client.

There are few things that we should look at in error package. 
- We got this error type called **ErrCode**. It represents a code of a particular error. It implements the MarshalText and UnmarshalText methods, which means we can use it as a field and unmarshal it's data back and forth to a string from an int.

```go
// ErrCode represents an error code in the system.
type ErrCode struct {
	value int
}

// String returns the string representation of the error code.
func (ec ErrCode) String() string {
	return codeNames[ec]
}

// UnmarshalText implement the unmarshal interface for JSON conversions.
func (ec *ErrCode) UnmarshalText(data []byte) error {
	errName := string(data)

	v, exists := codeNumbers[errName]
	if !exists {
		return fmt.Errorf("err code %q does not exist", errName)
	}

	*ec = v

	return nil
}

// MarshalText implement the marshal interface for JSON conversions.
func (ec ErrCode) MarshalText() ([]byte, error) {
	return []byte(ec.String()), nil
}

// Equal provides support for the go-cmp package and testing.
func (ec ErrCode) Equal(ec2 ErrCode) bool {
	return ec.value == ec2.value
}
```

- We have defined our own set of error codes to sort of detach ourselves from HTTP. Ideally it might be overkill if you are only using HTTP but if your are going to have a system that's may be using different protocols gRPC, WebSockets etc, then it would be nice to have an abstraction.
- Below is our error type and it implements Error interface. It has a code and a message. We can use this to send back error responses to the client.

```go
type Error struct {
	Code     ErrCode `json:"code"`
	Message  string  `json:"message"`
	FuncName string  `json:"-"`
	FileName string  `json:"-"`
}

// New constructs an error based on an app error.
func New(code ErrCode, err error) *Error {
	pc, filename, line, _ := runtime.Caller(1)

	return &Error{
		Code:     code,
		Message:  err.Error(),
		FuncName: runtime.FuncForPC(pc).Name(),
		FileName: fmt.Sprintf("%s:%d", filename, line),
	}
}

// Errorf constructs an error based on a error message.
func Errorf(code ErrCode, format string, v ...any) *Error {
	pc, filename, line, _ := runtime.Caller(1)

	return &Error{
		Code:     code,
		Message:  fmt.Sprintf(format, v...),
		FuncName: runtime.FuncForPC(pc).Name(),
		FileName: fmt.Sprintf("%s:%d", filename, line),
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Encode implements the encoder interface.
func (e *Error) Encode() ([]byte, string, error) {
	data, err := json.Marshal(e)
	return data, "application/json", err
}

// HTTPStatus implements the web package httpStatus interface so the
// web framework can use the correct http status.
func (e *Error) HTTPStatus() int {
	return httpStatus[e.Code]
}

// Equal provides support for the go-cmp package and testing.
func (e *Error) Equal(e2 *Error) bool {
	return e.Code == e2.Code && e.Message == e2.Message
}
``` 

Any time we are implementing the error interface using a struct, we need to use pointer semantics. It's because of the internals of how concrete values are stored inside of interfaces.

In Error type we have FuncName and FileName fields. These are not going to be sent back to the client, and that's why we are having dash in the JSON tags. What we are doing there is when we construct an error value we're going to capture what the **FunctionName** and file function name line of code is we are at. 
That's because we are probably going to have some middleware that' going to log the error but that's always going to be in the same place and when you look at the log it always tells that error occurred in the middleware function. It's not where we logged the error, it's where we constructed the error.

We also have implemented the Encoder interface so we can use this error type to send back error responses to the client. We also have implemented HTTPStatus method so we can use this error type to send back correct HTTP status code to the client.

HTTPStatus is something else that the web framework uses.
```go
func (e *Error) HTTPStatus() int {
	return httpStatus[e.Code]
}
```

On the response if you notice, it's going to check if the response implements httpStatus method is implemented on the value that we are passing in, which is essentially going to be the value implementing the Encoder. So what this code says is if the concrete value stored inside the resp also implements HTTPStatus() method then use it.

```go
func Respond(ctx context.Context, w http.ResponseWriter, resp Encoder) error {
	if _, ok := resp.(NoResponse); ok {
		return nil
	}

	// If the context has been canceled, it means the client is no longer
	// waiting for a response.
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return errors.New("client disconnected, do not send response")
		}
	}

	statusCode := http.StatusOK

	switch v := resp.(type) {
	case httpStatus:
		statusCode = v.HTTPStatus()

	case error:
		statusCode = http.StatusInternalServerError

	default:
		if resp == nil {
			statusCode = http.StatusNoContent
		}
	}

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	data, contentType, err := resp.Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("respond: encode: %w", err)
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("respond: write: %w", err)
	}

	return nil
}
```

We're implementing the interface also as well and we are using that so we can convert our code back to
the HTTP status code, but we are doing it in a way that we are not coupling our error type to HTTP. 
We don't have a import, it's all about **checking to see that if this behaviour i.e. HTTPStatus() exists is one of the powerful things about Go**.

We are just saying if you want to be able to send back correct HTTP status code, then implement this interface and we will use it. If you don't want to implement this interface, we will just default to 500.

```go
func (e *Error) HTTPStatus() int {
	return httpStatus[e.Code]
}
```

We should not have met the second switch which means we should never have an error get all the way 
through. It doesn't mean that the value that's stored inside the encoder also doesn't implement error. It will but that would mean that something sort of leaked and we lost control and we better do a 500.

We are going to use handleFunc and once our handler returns we call respond and do all the things
auto-magically.
```go
func (a *App) HandlerFunc(method string, group string, path string, handlerFunc HandlerFunc, mw ...MidFunc) {
	handlerFunc = wrapMiddleware(mw, handlerFunc)
	handlerFunc = wrapMiddleware(a.mw, handlerFunc)

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := setWriter(r.Context(), w)

		resp := handlerFunc(ctx, r)

		if err := Respond(ctx, w, resp); err != nil {
			a.log(ctx, "web-respond", "ERROR", err)
			return
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	finalPath = fmt.Sprintf("%s %s", method, finalPath)

	a.mux.HandleFunc(finalPath, h)
}
```

Next we are going to copy middleware such as errors.go, panics.go, logging.go. 

errors.go middleware handles all our errors, it will call the next handler, checks to see if there is error in the response, it constructs one of our errors if it doesn't already exist, it will log 
everything with the function name and the file name and sometimes you have an internal error that you only want to log and you want to send internal server error the other way, so we go error code for that.

```go
func Errors(log *logger.Logger) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			resp := next(ctx, r)

			err := checkIsError(resp)
			if err == nil {
				return resp
			}

			var appErr *errs.Error
			if !errors.As(err, &appErr) {
				appErr = errs.Errorf(errs.Internal, "Internal Server Error")
			}

			log.Error(ctx, "handled error during request",
				"err", err,
				"source_err_file", path.Base(appErr.FileName),
				"source_err_func", path.Base(appErr.FuncName))

			if appErr.Code == errs.InternalOnlyLog {
				appErr = errs.Errorf(errs.Internal, "Internal Server Error")
			}

			// Send the error to the web package so the error can be
			// used as the response.

			return appErr
		}

		return h
	}

	return m
}
```

Logging is going to be for every request, it's good to have started and completed logs in Goroutines,
if we see started but don't see completed, then we know that something is wrong.

```go
func Logger(log *logger.Logger) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			now := time.Now()

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			log.Info(ctx, "request started", "method", r.Method, "path", path, "remoteaddr", r.RemoteAddr)

			resp := next(ctx, r)

			var statusCode = errs.None
			if err := checkIsError(resp); err != nil {
				statusCode = errs.Internal

				var appErr *errs.Error
				if errors.As(err, &appErr) {
					statusCode = appErr.Code
				}
			}

			log.Info(ctx, "request completed", "method", r.Method, "path", path, "remoteaddr", r.RemoteAddr,
				"statuscode", statusCode, "since", time.Since(now).String())

			return resp
		}

		return h
	}

	return m
}
```

Here we check if there is error, we are propagating it all the way back to the web framework. 
Middleware framework is going to check to see if the response is an error, if it is an error, it will log it and then send it back to the web framework. 

We could also send the response direclty from the middleware in case of errors but there are chances that thing could fail after we did that and then we couldn't report back on it.

Then the next middleware is for handling panics. The key about handling the panics is calling recover
inside of a defer.

```go
func Panics() web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) (resp web.Encoder) {

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					trace := debug.Stack()
					resp = errs.Errorf(errs.InternalOnlyLog, "PANIC [%v] TRACE[%s]", rec, string(trace))
				}
			}()

			return next(ctx, r)
		}

		return h
	}

	return m
}
```

Remember the defer is executed after the function h returns. Now the problem is that these defer dunctions don't have return type, so how do we return an error back to the calling function? In this 
case we use named return values, which we are setting from inside defer, so we are now using closures to set it.

We still want to be in control when our handler panics.

# Package Design
Ideally you don't want packages that contain things and good smell for that is when you don't have a file named after the package. If you look at mid.go it has just CheckIsError function, that's like cheating. This is the package that contains middleware functions. You can get away with containment
if the import tree to this is really really tiny which it is.

This is very specific containment package, it got middleware functions. It's only going to be used by one other package which is our mux package, which we haven't brought in yet. So it's not going to cause a problem. 

But when you have these containment packages like utils, helpers, commons etc. where like half the project if not more is importing them, you've done very bad things for yourself.

So we got to identify when these containment packages come up where it doesn't make sense to have a 
file named after it. And we have to be very careful about what is the import tree to that package.
If it's tiny then we could be fine like for sure at the app layer it probably is tiny because it's very specific to a domain or two but a package like that in the foundation you can pretty much bet
it's probably going to be abused in terms of its import.

Last thing is that mux function but we don't have a domain yet to bring it in.

We are going to build the domain package now, then we can bring in the mux package and then we can wire it into main and we could test that. We are going to call this domain chatapp for now.
These domain packages at the app layer they're responsible for:
- Receiving that external input.
- Validating that external input.
- Calling the business logic.
- Formulating the response or returning errors back to the middleware.

We don't know yet that we have a business layer yet. If we look at the diagram inside cap service
the two boxes that we have may represnt the business layer but they may be app layer sdk packages providing support. Rightnow there is no business logic, these are just sort of routing stuff. 

We can start them at the app/sdk layer now. We could make these even more generic at some point and
move them into a foundation layer but we would start at the app/sdk layer, see how we are using it, see if it's reasonable to make it little bit more generalized before we move it. But the question is,
is it something we can reuse in another project? If the answer is yes, then we foundation it, if the answer is no we keep it at the app/sdk level and then if we feel like there's some business logic in
there that goes beyond just app layer protocol stuff, then we can move it to the business layer.

Below is our chatapp.go
```go
type app struct {
}

func newApp() *app {
	return &app{}
}

func (a *app) test(ctx context.Context, r *http.Request) web.Encoder {
	return status{
		Status: "ok",
	}
}
```

Now we have model.go implementing this Encoder interface.
```go
type status struct {
	Status string `json:"status"`
}

// Encode implements the encoder interface.
func (app status) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}
```

And finally we have route.go containing the routes
```go
func Routes(app *web.App) {
	api := newApp()

	app.HandlerFuncNoMid(http.MethodGet, "", "/test", api.test)
}
```

Here we are using golbal mux.

Now let's add mux package inside sdk.

```go
// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log         *logger.Logger
}

// WebAPI constructs a http.Handler with all application routes bound.
func WebAPI(cfg Config) http.Handler {
	logger := func(ctx context.Context, msg string, args ...any) {
		cfg.Log.Info(ctx, msg, args...)
	}

	app := web.NewApp(
		logger,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Panics(),
	)

	chatapp.Routes(app)
	
	return app
}
```

chatapp.Routes binds all the routes. 

In the service project at build time you can say build me an instance of the service with [all the routes](https://github.com/ardanlabs/service/blob/master/api/services/sales/build/all.go) or you can say build me an instance which has [crud routes](https://github.com/ardanlabs/service/blob/master/api/services/sales/build/crud.go) or build me an instance which has [reporting routes](https://github.com/ardanlabs/service/blob/master/api/services/sales/build/reporting.go). It's done very quickly using a variable called route.

When you do go build, the go build tool takes ldflag X let's us change that string
```sh
go build -ldflags="-X main.routes=crud"
```

So at build time you could specify the routes and the mux would be using whatever is defined in that route.

# Websocket Implementation
We are going to take use Gorilla Websocket package. And take code from [blockchain project](https://github.com/ardanlabs/blockchain/blob/main/app/services/node/handlers/public/public.go).

In order for both side to know they are still there, there is a ping mechanic.

Let's say we have a package that's listening on the channel, when some data comes in, let's say
GRead (go routine for reading) it's job is to pull it and send it to the Router package.

Router package figures out the routing and sends it to the client GWrite (Go routine for writing) and then GWrite is responsible for writing it back to the client.

So GRead goroutine just reads and sends data, GWrite goroutine receives something and just writes data. So both have single purpose.

GRead can send data even on API we don't need channels for that, we can just call the function directly. But for GWrite we have two options:
- Router package can directly write message to the socket.
- Router knows the GoRoutine that's managing the writes and send data to that GoRoutine and that GoRoutine is responsible for writing it back to the client.

In this case we are going to second option because then we can have both read and write at a single place and it will be easy to fix any issues in future.

```go
```


Below is the code for handshake as we need to do a handshake before we can start sending messages back and forth between the client and the server.

```go
func (a *app) connect(ctx context.Context, r *http.Request) web.Encoder {
	// Web socket implemented here.
	// Just perform basic echo for now.
	// Make sure we are handling connection drops/issues (context)
	// How we will map connection to a user.
	c, err := a.WS.Upgrade(web.GetWriter(ctx), r, nil)
	if err != nil {
		return errs.Errorf(errs.FailedPrecondition, "unable to upgrade to websocket: %v", err)
	}
	defer c.Close()

	usr, err := a.handshake(ctx, c)

	if err != nil {
		return errs.Errorf(errs.FailedPrecondition, "handshake failed: %v", err)
	}

	a.log.Info(ctx, "handshake complete", "user", usr.Name)

	return nil
}

func (a *app) handshake(ctx context.Context, c *websocket.Conn) (user, error) {
	if err := c.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		return user{}, fmt.Errorf("write message: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()

	msg, err := a.readMessage(ctx, c)
	if err != nil {
		return user{}, fmt.Errorf("read message: %w", err)
	}

	var usr user

	if err := json.Unmarshal(msg, &usr); err != nil {
		return user{}, fmt.Errorf("unmarshal message: %w", err)
	}

	fmt.Printf("handshake successful, got message: %s\n", msg)

	v := fmt.Sprintf("WELCOME %s", usr.Name)
	if err := c.WriteMessage(websocket.TextMessage, []byte(v)); err != nil {
		return user{}, fmt.Errorf("write message: %w", err)
	}

	return usr, nil
}

func (a *app) readMessage(ctx context.Context, c *websocket.Conn) ([]byte, error) {
	type response struct {
		msg []byte
		err error
	}
	ch := make(chan response, 1)

	go func() {
		a.log.Info(ctx, "starting handshake read")
		defer a.log.Info(ctx, "completed handshake read")
		_, msg, err := c.ReadMessage()

		if err != nil {
			ch <- response{msg: nil, err: err}
			return
		}

		ch <- response{msg: msg, err: nil}
	}()

	var resp response

	select {
	case <-ctx.Done():
		c.Close()
		return nil, fmt.Errorf("handshake timeout: %w", ctx.Err())
	case resp = <-ch:
		if resp.err != nil {
			return nil, fmt.Errorf("handshake failed: %w", resp.err)
		}
	}

	return resp.msg, nil
}
```

Problem with readMessage above is that it has goroutine bug.

If we had a unbufferred channel and a goroutine writing the message to the channel, we should have a goroutine in receive state otherwise write gets blocked. And if we timedout and left and when ReadMessage() returns it's going to be blocked as we are not listening to the channel anymore.

We are leaking a Goroutine after Goroutine, after Goroutine and so on.

Quickest way to fix this is to make the channel buffered, so that when we timeout and leave, the goroutine can still write to the channel and exit gracefully. So we are not going to have any leaks.

Now to test this code we need client. So let's create client package inside tooling. 

```go
func main() {
	if err := hack1("ws://localhost:3000/connect"); err != nil {
		log.Fatal(err)
	}
}

func hack1(url string) error {
	req := make(http.Header)
	socket, _, err := websocket.DefaultDialer.Dial(url, req)

	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	defer socket.Close()

	// -------------------------------------------

	_, msg, err := socket.ReadMessage()
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	log.Printf("Received message: %s", msg)

	if string(msg) != "HELLO" {
		return fmt.Errorf("unexpected message: %s", msg)
	}

	// -------------------------------------------

	usr := struct{
		Name string `json:"name"`
		ID uuid.UUID `json:"id"`
	}{
		Name: "Vivek",
		ID: uuid.New(),
	}

	data, err := json.Marshal(usr)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := socket.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	// -------------------------------------------

	_, msg, err = socket.ReadMessage()
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	
	log.Printf("Received message: %s", msg)

	return nil
}
```

