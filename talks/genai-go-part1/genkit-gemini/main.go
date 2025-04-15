package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/server"
)

// curl -X POST "http://localhost:3400/menuSuggestionFlow" -H "Content-Type: application/json" -d '{"data": "Indian"}'
func main() {
	ctx := context.Background()

	// Initialize the Google AI plugin. When you pass nil for the
	// Config parameter, the Google AI plugin will get the API key from the
	// GOOGLE_GENAI_API_KEY environment variable, which is the recommended
	// practice.
	g, err := genkit.Init(ctx,
		// Install the Google AI plugin which provides Gemini models.
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		// Set the default model to use for generate calls
		genkit.WithDefaultModel("googleai/gemini-2.0-flash"),
	)
	if err != nil {
		log.Fatal(err)
	}
	menuSuggestionFlow := genkit.DefineFlow(g, "menuSuggestionFlow", func(ctx context.Context, theme string) (string, error) {
		resp, err := genkit.Generate(ctx, g, ai.WithConfig(&googlegenai.GeminiConfig{Temperature: 1}), ai.WithPrompt(fmt.Sprintf(`Suggest an item for the menu of a %s themed resturant`, theme)))
		if err != nil {
			return "", err
		}

		// Handle the response from the model API. In this sample we just convert it to
		// a string, but more complicated flows might coerce the response into structured
		// output or chain the response into another LLM call.
		text := resp.Text()
		return text, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /menuSuggestionFlow", genkit.Handler(menuSuggestionFlow))
	log.Fatal(server.Start(ctx, "127.0.0.1:3400", mux))
}
