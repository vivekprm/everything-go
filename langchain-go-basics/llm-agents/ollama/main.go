package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/serpapi"
)

func main() {
	llm, err := ollama.New(ollama.WithModel("llama2"))
	if err != nil {
		log.Fatal(err)
	}

	prompt := "Who is Olivia Wilde's boyfriend? What is his current age raised to the 0.23 power?"
	baseline, _ := llm.Call(context.Background(), prompt)
	fmt.Println("Model response:")
	fmt.Println("=================")
	fmt.Println(baseline)

	// search := wikipedia.New("langchaingo test (https://github.com/tmc/langchaingo)")
	search, err := serpapi.New()
	if err != nil {
		log.Fatal(err)
	}
	agentTools := []tools.Tool{
		search,
	}
	agent := agents.NewOneShotAgent(llm,
		agentTools,
		agents.WithMaxIterations(1))
	executor := agents.NewExecutor(agent)
	ans, err := chains.Run(context.Background(), executor, prompt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Agent response:")
	fmt.Println("=================")
	fmt.Println(ans)
}
