# Retrieval Augmented Generation
How do I use a model but give it my own data and oneway of doing that is fine-tuning. Fine-tuning is basically training the model but without like retraining it. So if you take a model, you can then essentially strip out these outer layers, replace outer layers with your own data. You have to continually finetune because it's not like dynamic process.

One of the other methods people hace comeup with is a notion called RAG.

## How do we best augment LLMs with our own private data?
pic

How do you get data into your LLM, how do you give this thing more context and get your data in without having to specifically fine-tune or retrain your model. It's more affordable rather than finetuning the model and most people do this.

## RAG
'Grounding' on user data.

**The Problem**:
- LLMs don't know your business's proprietary or domain specific data.
- LLMs don't have real-time information.
- LLMs find it challenging to provide accurate citations from their parametric knowledge.

**The Solution**:
Feed the LLM *relevant* context in real-time, by using an information retrieval system.

How do you make it such that it works well? What are some of the algorithms that you can do to make this work well with LLMs?
There are numbers of algorithms out there like Vector Search, Encodings & Embeddings which are a great way of being able to index data, kind of search on data with compares. So doing not quite the same thing as what LLMs do in their Neural Networks but it's a great way of instead of you don't want to just say like obviously you could pass in all text. If you want to say rewrite a document you can pass in whole blob of stuff.

You have whole bunch of information related to some query, you might have a corpus of millions and millions of records, you want to find basically something like nearest neighbour or the ones that closely match that. The fastest way to search that stuff turns out to be these vector embeddings for searching. Which can create these sort of multi-relationship deep sort of network. They turn out to be very useful as inputs into LLMs.

- Essentially you'll use a model, an embedding model to generate text embeddings, 
- You will store those in a **Vector Database**.
- And then you will retrieve those using stnadard query. 
- And you take that data and pass that into your model as context and then hopefully magic happens and you get something that's like more about your data than anything else.

### Modified Prompt
```
You are an intelligent assistant helping the users with their questions on 
{{Company | research papers | ...}}. Strictly use ONLY the following pieces of 
context to answer the question at the end. Think step-by-step and then answer.

Do not try to make up an answer:
- If the answer to the question cannot be determined from the context alone, say "I cannot determine answer to that."
- If the context is empty, just say "I do not know the answer to that."

CONTEXT:
{{retrieved_information}}

QUESTION:
{{question}}

Helpful Answer:
```

In this particular case, we are telling it bunch of stuff, the interesting thing is you tell it in the prompt that you wanted to use the context so interestingly enough these models all understand fundamentally the ability to tell it to actually use something that you've given it. In this case it says if there's no context, don't answer the question.

So in theory it shouldn't make up an answer or it should only be using the the inputs that you give it as the data.

## Classic Information Retrieval
pic
- General purpose of an IR system is to help user find relevant information.
- Allow users retrieve items that fully or partially satisfy their information need, specified through a query.
- Typically in 2 stages:
  - **Retrieval**: retrieves a set of initial documents that are likely to be relevant to the query.
  - **Ranking**: rank the document based on their relevance score.
- Goal is to achieve high effectiveness i.e. maximize the effort-satisfaction proportion of its users about queries.

## Embeddings
pic

An *embedding* is relatively low dimensional vector into which you can translate high-dimensional vectors. Ideally, an embedding captures some of the semantics of the input by placing *semantically similar inputs close together* in the embedding space. 

It's very similar to how Network Models works it's just different version of it with specific set of data indexed on more searching and finding close relationships and edges that are in there, where inputs are closer together.

## Vector Search is a key component in Gen AI applications
pic

## Vector Search
pic

### How to find similar embeddings in the embedding space?
- Calculate the distance or similarity between vectors.
- Not easy when you have millions or billions of embeddings. If you have 8 million embeddings with 768 dimensions, you would need to repeat the calculation in the order of 8 million * 768. This would take very long time to finish.
- Use **Approximate Nearest Neighbour** far faster search. ANN uses "vector quantization" for separating the space into multiple spaces with a tree structure. This is similar to the index in relational databases for improving the query performance, enabling very fast and scalable search with billions of embeddings.

## RAG Workflow for building a QA system
pic

You can pass in documents raw text, so you have to have something that can do pipeline processing, adds it to the Vector Database and then on the other side you have a retrieval side and then you have a model that takes that and spits out an answer.

### Data Ingestion / Parsing
- Split up document(s) into even chunks.
- Each chunk is a piece of raw text.
- Generate Embedding for each chunk.
- Store each chunk into a Vector Database.

### Querying
- Find top-k most similar chunks from vector database.
- Plug into LLM response synthesis.

In demo we are going to have both parts into one but typically it's two seperate processes.

Generative AI is not actually QA. Even though you gave your specific data in context, it might not spit back exactly what you stored in your Vector Database. This is Generative AI, hopefully it's using a smaller corpus of information to do that so it's going to find that relationship to be in there but don't think that just because you put something in the Vector Database, you would actually get back if you asked the exact samething but this atleast helps approximate and narrowdown the set of data the model is going to operate on.  

```go
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
```

Here we are using in-memory VectorDatabase **localvec** great for testing, it's easier to run this rather than setting up your own vector database. You can use pgvector with postgres, but that requires some setup. It's easy to use in-memory database.

We can use genkit-cli and use genkit UI to interact with our model:
```sh
npm i -g genkit-cli
genkit start -- go run main.go
```

Normally you would like to put pdfs, unstructured data and all kinds of information relevant to you. E.g. For one of our product called **cloud-assist** what we are doing is, because we can't possibly like retrain the model on all the github code that's out there and may be you want to train it on some of your local code, the way that's typically going to work is you are not going to finetune the model, what you are going to do is use Vector Embeddings for your actual code. 
You take the corpus of your code, generate embeddings for that, store that in a Vector Database and then pass that in as a context and then generate code from that. And that's how overtime you'll start to see that you can actually get the code to be more specific to maybe the style that you use for example. As opposed to like a style that was found on the internet for training.

If it's text embedding model that we are using to generate embeddings, we need to convert pdfs etc. into text and then store the embeddings.
There are multi-model embedding models they can accept multiple types of blobs. Genkit has different parsers for different format of data.

Q: How do we prevent Hellucinations?
A: In the case of GenAI the answers are not going to be deterministic. So you kind of have to use AI or other sort of models to sort of compare the answers. Think of it as in most generic term as fuzzy match. If I just trained a model like in above case with RAG which it should hellucinate less, I should be able to know that my answers are somewhere in that.

So you have to use another set of similar AI or tools to actually check that your answers were true. If you are trying to prevent some of this stuffs from happening, there are few things that you can do:
- You can try in your prompt to say, a number of models out there are trained with like wiered sort of keywords in there. So you can tell it to be an accurate assistant. You can tell it, if you don't know the answer say don't know. So depending on the model or sort of layers you can also just adjust your basic prompts to help it avoid like blatant hallucinations, like when there was nothing it just makes it up. 
- But more specifically if you are really going to build one out there you're probably going to do it on corpus of your own data and you're going to have to do something analogous to fuzzy matching to make sure that stuff is up there. And then you're probably going to also have to like periodically just run random tests on live running model just to see that it's behaving the way we wanted it to. 

Q: How big a prompt can be?
A: It depends on the model, you can obviously set it too server side. But from model perspective yes there is a limit which they generally tell based on the number of parameters.

Bigger the context more accurate RAG will be.