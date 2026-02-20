# What does the AI Agent need?
- Necessary context.
- Sufficiently powered tooling.

Which were not present when using LLMs. This is where tooling provided with MCP becomes handy.

# Introducing the Model Context Protocol (MCP)
MCP is an opensource protocol that standardizes how applications provide context to LLMs.

MCP framework is going to help improve the accuracy of LLMs, tailor them to the needs of your organization or your userbase and equip LLMs with the power to perform
useful actions within your system.

So in simple terms, MCP is a protocol for AI models to interact with external tools. You can think of it like a universal system agnostic language like TCP or HTTP
but for AI applications enabling them to interact with external tools and resources in a consistent way.

Let's look at a practical example:

<img width="876" height="506" alt="Screenshot 2026-02-20 at 12 14 25 PM" src="https://github.com/user-attachments/assets/8a08ff30-5597-4447-9125-00f32323b128" />

Let's imagine we are running Kubernetes commands and experience a sudden crash. In an AI environment with proper MCP server tools installed, this may be as simple as
asking. So for example, you are in your IDE, you ask your AI agent something like "Hey, can you create an issue in the Kubernetes repository that describes the bug that I just encountered?"

So the AI agent has to do a lot to solve this problem. It has to perform a number of actions and gather information from various sources to achieve it. So for example
- What bug just occurred? The AI agent needs to understand the context of the crash, gather logs, and analyze them to identify the issue.
- What even is Kubernetes?
- What repo is Kubernetes in?
- It has to format the bug, it has to actually create the issue on Github.
- Give the issue link to the user.

Some of this, it can do on it's own using local state or what's inside of the Agent. But things like the repository it doesn't know what that is. It's going to 
interact with Github. 

So those questions like ""which repo is Kubernetes in?" and "actually creating the issue itself" requires the interfacing with Github platfomr.
The AI agents can do this through tools provided by the Github MCP server installed in your environment and configured with your Github credentials.

This is actually already available and you can use it today and it's free. You can download Github MCP server, if you want to experiment with it.

If you notice closely these MCP tools look a lot like API functions that a typical server will expose.
It can do things like:
- Create Issue
- List Issues
- Get pull request details
- Search
- Create Branches
- etc.

Difference is rather than you having to handwrite code that calls every single relevant API, AI agents is going to figure our all that for us.

Now let's imagine we have tons of MCP servers that are going to help us.

<img width="934" height="466" alt="Screenshot 2026-02-20 at 1 02 17 PM" src="https://github.com/user-attachments/assets/785f32ab-79ee-44d0-a2b2-5690c0e49f0d" />

So for example, may be we have a MCP server A that's going to be spun up by our IDE and then it's going to automatically connect an MCP client to it.
So now when you ask it to update a database, it's going to do it for you using a MCP server A that can help.

Or imagine you ask your IDE's AI agent to build and deploy your code. Well maybe you have a MCP server B that has credentials configured to interface with Github
and Github actions and Gitlab in the way that's specific to your organization.

Or imagine you ask that company's specific question that's not available in the public internet.

Or you already have a http server that's running on your company's cluster and you just want to be able to interfacee with it in your IDE. May be MCP server C
is just a very thin wrapper around this http server which can help the agent by speaking the MCP protocol.

So in short:
- LLMs empower users to perform actions using human language.
- MCP servers empower the LLMs to be more accurate and targeted.

We will be using go-sdk to build our MCP server and we will be using the go-sdk client to interface with it. The go-sdk client is going to be used by the AI agent to interact with the MCP server

https://github.com/modelcontextprotocol/go-sdk

We will be building a MCP server that interfaces with the Gophercon agenda so that you can get very good information about Gophercon.

# MCP Server
Below is very basic simple MCP server that greets.

```go
package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Input struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
}

type Output struct {
	Greeting string `json:"greeting" jsonschema:"the greeting to tell to the user"`
}

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input Input) (
	*mcp.CallToolResult,
	Output,
	error,
) {
	return nil, Output{Greeting: "Hi " + input.Name}, nil
}

func main() {
	// Create a server with a single tool.
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "greeter",
		Version: "v1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "greet",
		Description: "say hi",
	}, SayHi)

	// Run the server over stdin/stdout, until the client disconnects.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

```
Now build the server to create an executable:
```
go build -o myserver main.go
```
Now make this binary executable and update the $PATH variable to use this executable.

# MCP Client
Let's create another directlory called client and create a file main.go in it. This is going to be our client that interfaces with the server.

```go
package main

import (
	"context"
	"log"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	// Create a new client, with no features.
	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "v1.0.0"}, nil)

	// Connect to a server over stdin/stdout.
	transport := &mcp.CommandTransport{Command: exec.Command("myserver")}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// Call a tool on the server.
	params := &mcp.CallToolParams{
		Name:      "greet",
		Arguments: map[string]any{"name": "you"},
	}
	res, err := session.CallTool(ctx, params)
	if err != nil {
		log.Fatalf("CallTool failed: %v", err)
	}
	if res.IsError {
		log.Fatal("tool failed")
	}
	for _, c := range res.Content {
		log.Print(c.(*mcp.TextContent).Text)
	}
}
```

Now run the client as below:

```go
go run main.go
```
