package main

// https://github.com/firebase/genkit/blob/main/go/samples/rag/main.go
import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/localvec"
)

// curl -d '{"question": "What is the capital of UK?"}' http://localhost:3400/simpleQaFlow
const simpleQaPromptTemplate = `
You're a helpful agent that answers the user's common questions based on the context provided.

Here is the user's query: {{query}}

Here is the context you should use: {{context}}

Please provide the best answer you can.
`

type simpleQaInput struct {
	Question string `json:"question"`
}

type simpleQaPromptInput struct {
	Query   string `json:"query"`
	Context string `json:"context"`
}

func main() {
	ctx := context.Background()
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel("googleai/gemini-1.0-pro"),
	)
	if err != nil {
		log.Fatal(err)
	}
	const embedderName = "embedding-001"
	embedder := googlegenai.GoogleAIEmbedder(g, embedderName)
	if embedder == nil {
		log.Fatalf("embedder %s is not known to the googlegenai plugin", embedderName)
	}
	if err := localvec.Init(); err != nil {
		log.Fatal(err)
	}
	indexer, retriever, err := localvec.DefineIndexerAndRetriever(g, "simpleQa", localvec.Config{Embedder: embedder})
	if err != nil {
		log.Fatal(err)
	}

	m := googlegenai.GoogleAIModel(g, "gemini-2.0-flash")
	if m == nil {
		log.Fatal("jokesFlow: failed to find model")
	}

	simpleQaPrompt, err := genkit.DefinePrompt(g, "simpleQaPrompt",
		ai.WithModel(m),
		ai.WithPrompt(simpleQaPromptTemplate),
		ai.WithInputType(simpleQaPromptInput{}),
		ai.WithOutputFormat(ai.OutputFormatText),
	)
	if err != nil {
		log.Fatal(err)
	}

	genkit.DefineFlow(g, "simpleQaFlow", func(ctx context.Context, input *simpleQaInput) (string, error) {
		d1 := ai.DocumentFromText("Paris is the capital of France", nil)
		d2 := ai.DocumentFromText("USA is the largest importer of coffee", nil)
		d3 := ai.DocumentFromText("Water exists in 3 states - solid, liquid and gas", nil)

		err := ai.Index(ctx, indexer, ai.WithDocs(d1, d2, d3))
		if err != nil {
			return "", err
		}

		dRequest := ai.DocumentFromText(input.Question, nil)
		response, err := ai.Retrieve(ctx, retriever, ai.WithDocs(dRequest))
		if err != nil {
			return "", err
		}

		var sb strings.Builder
		for _, d := range response.Documents {
			sb.WriteString(d.Content[0].Text)
			sb.WriteByte('\n')
		}

		promptInput := &simpleQaPromptInput{
			Query:   input.Question,
			Context: sb.String(),
		}

		resp, err := simpleQaPrompt.Execute(ctx, ai.WithInput(promptInput))
		if err != nil {
			return "", err
		}
		return resp.Text(), nil
	})
	// Later, run the flow:
	// ask(ctx, myFlow, "What is the capital of UK?")
	// ask(ctx, myFlow, "What is the capital of Paris?")
	<-ctx.Done()
}

func ask(ctx context.Context, myFlow *core.Flow[*simpleQaInput, string, struct{}], question string) {
	result, err := myFlow.Run(ctx, &simpleQaInput{
		Question: question,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
