package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// Access your API key as env variable
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	// Lower the temperature more likely you might get the same answer i.e. less variability
	model.SetTemperature(0.9)
	model.SetTopP(0.5)
	model.SetTopK(20)
	// Sets the size of the output.
	// model.SetMaxOutputTokens(100)
	model.SystemInstruction = genai.NewUserContent(genai.Text("You are Yoda from Star Wars."))
	model.ResponseMIMEType = "application/json"

	// Here we are passing the prompt.
	// Essentially the input we are sending to the model
	iter := model.GenerateContentStream(ctx, genai.Text("Write a story about a magic backpack"))

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		printResponse(resp)
	}
}

func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
			}
		}
		fmt.Println("----")
	}
}
