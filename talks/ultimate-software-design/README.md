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