# Agents
## What is an agent?
An agent consists of three components: a large language model (LLM), a set of tools it can use, and a prompt that provides instructions.

The **LLM operates in a loop**. In each iteration, 
- It selects a tool to invoke, 
- Provides input, 
- Receives the result (an observation), 
- And uses that observation to inform the next action. 
- The loop continues until a stopping condition is met — typically when the agent has gathered enough information to respond to the user.

pic

## Community Agents
If you’re looking for other prebuilt libraries, explore the community-built options below. These libraries can extend LangGraph's functionality in various ways.

https://langchain-ai.github.io/langgraph/agents/prebuilt/#available-libraries

## MCP Integration
[Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction) is an open protocol that standardizes how applications provide tools and context to language models. LangGraph agents can use tools defined on MCP servers through the **langchain-mcp-adapters** library.

```sh
pip install langchain-mcp-adapters
```

pic

### Custom MCP servers
To create your own MCP servers, you can use the mcp library. This library provides a simple way to define tools and run them as servers.

```sh
pip install mcp
```

https://modelcontextprotocol.io/introduction
https://modelcontextprotocol.io/docs/concepts/transports

## Multi-agent
A single agent might struggle if it needs to specialize in multiple domains or manage many tools. To tackle this, you can break your agent into smaller, independent agents and composing them into a [multi-agent system](https://langchain-ai.github.io/langgraph/concepts/multi_agent/).

In multi-agent systems, agents need to communicate between each other. They do so via [handoffs](https://langchain-ai.github.io/langgraph/agents/multi-agent/#handoffs) — a primitive that describes which agent to hand control to and the payload to send to that agent.

Two of the most popular multi-agent architectures are:

[supervisor](https://langchain-ai.github.io/langgraph/agents/multi-agent/#supervisor) — individual agents are coordinated by a central supervisor agent. The supervisor controls all communication flow and task delegation, making decisions about which agent to invoke based on the current context and task requirements.
[swarm](https://langchain-ai.github.io/langgraph/agents/multi-agent/#swarm) — agents dynamically hand off control to one another based on their specializations. The system remembers which agent was last active, ensuring that on subsequent interactions, the conversation resumes with that agent.

[LangGraph](https://langchain-ai.github.io/langgraph/) provides both low-level primitives and high-level prebuilt components for building agent-based applications. This section focuses on the prebuilt, reusable components designed to help you construct agentic systems quickly and reliably—without the need to implement orchestration, memory, or human feedback handling from scratch.

# AgentExecutor
To make agents more powerful we need to make them iterative, ie. call the model multiple times until they arrive at the final answer. That's the job of the AgentExecutor.

An example that initialize a **MRKL** (Modular Reasoning, Knowledge and Language, pronounced "miracle") agent executor.

Agents use an LLM to determine which actions to take and in what order. An action can either be using a tool and observing its output, or returning to the user.

When used correctly agents can be extremely powerful. In this tutorial, we show you how to easily use agents through the simplest, highest level API.

In order to load agents, you should understand the following concepts:

- **Tool**: A function that performs a specific duty. This can be things like: **Google Search**, **Database lookup**, **code REPL**, other **chains**. The interface for a tool is currently a function that is expected to have a string as an input, with a string as an output.
- **LLM**: The language model powering the agent.
- **Agent**: The agent to use. This should be a string that references a support agent class. Because this notebook focuses on the simplest, highest level API, this only covers using the standard supported agents.

For this example, you'll need to set the SerpAPI environment variables in the .env file.

SERPAPI_API_KEY="..."

## Load the LLM
```go
llm, err := ollama.New(ollama.WithModel("llama2"))
if err != nil {
    return err
}
```

## Define Tools
```go
search, err := serpapi.New()
if err != nil {
    return err
}
agentTools := []tools.Tool{
    tools.Calculator{},
    search,
}
```

## Create Prompt
```go
prompt := "Who is Olivia Wilde's boyfriend? What is his current age raised to the 0.23 power?"
```

## Create the Agent Executor
```go
executor, err := agents.Initialize(
    llm,
    agentTools,
    agents.ZeroShotReactDescription,
    agents.WithMaxIterations(3),
)
if err != nil {
    return err
}

answer, err := chains.Run(context.Background(), executor, prompt)
fmt.Println(answer)
return err
```

You can compare this to the base LLM.
```go
baseline, _ := llm.Call(context.Background(), prompt)
fmt.Println(baseline)
```	

Below is the full agent code:
```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/serpapi"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	if err = run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	llm, err := openai.New()
	if err != nil {
		return err
	}
	search, err := serpapi.New()
	if err != nil {
		return err
	}
	agentTools := []tools.Tool{
		tools.Calculator{},
		search,
	}
	agent := agents.NewOneShotAgent(llm,
		agentTools,
		agents.WithMaxIterations(3))
	executor := agents.NewExecutor(agent)

	question := "Who is Olivia Wilde's boyfriend? What is his current age raised to the 0.23 power?"
	answer, err := chains.Run(context.Background(), executor, question)
	fmt.Println(answer)
	return err
}
```