https://youtu.be/8qVL8cef7Iw?si=4_NmnUFcKSGDK_6T

https://chat.demo.gke.ninja

https://github.com/mastersingh24/genai-with-go-gopherconuk24

# What are Large Language Models
- ML algorithms that can **recognize, predict and generate** human languages.
- Pre-trained on petabyte scale text-based datasets resulting in large models with **10s to 100s of billions of parameters.**
- LLMs are normally **pretrained on a large corpus of text** followed by fine-tuning on a specific task.
- LLMs can also be called **Large Models**(includes all types of data modality) and **Generative AI**(a model that produces content)

There is difference between Q&A Model & TextGenenration Model.
Q&A model will generate same answer for the same question. However Generative Model doesn't necessairly generate same text everytime.

- Generative AI is a subset of Deep Learning
- DeepLearning is a subset of ML
- Large Language Models are also a subset of Deep Learning.

# Why are Large Language Models Different?
- LLMs are characterized by **emergent abilities**, or the ability to perform tasks that were not present in smaller models.
- LLMs contextual understanding of human language **changes how we interact** with data and intelligent systems.
- LLMs can find patterns and connections in **massive, disparate data corpora**

# Generative AI is driving new opportunities
It is much easier to consume compared to prior models. You just need to know how to interact with model rather than knowing how model works. Lots of usecases:
- Analyst (Complex Data, intuitively accessible)
  - **Improve time-to-value** to search, navigate and extract insights and understanding from large amounts of complex data.
- Customer Service (Online transactions made conversational)
  - **Improve customer-experience**, reaching larger client bases by making online interactions more natural, conversational and rewarding.
- Creative (Content generation at the click of a button)
  - Generate code, text, image, video or music quickly and multi-modally, speeding up every business process and **maximizing employee productivity**.
- AI Practitioner (Customize Foundational Model)
  - Customize large models and incorporate state of the art generative capabilities natively into **your own internal ML operational platforms**.

You can say in google docs create presentation based on streaming content. 

Foundational models are produced by companies who have lots of money as it's very expensive to train these models. E.g.
- Gemini
- Chat GPT

# Growth in Go's Ecosystem leads new usecases
- Local AI
- Ollama
- LinGoose
- Firebase Genkit
- Milvus
- Zep
- Weaviate
- LangChain Go

## Example 1: Gemini with Google AI
```go
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
	// model.SystemInstruction = genai.NewUserContent(genai.Text("You are Yoda from Star Wars."))
	model.ResponseMIMEType = "application/json"

	// Here we are passing the prompt.
	// Essentially the input we are sending to the model
	iter := model.GenerateContentStream(ctx, genai.Text("Write a story about a magic backpack"))

	for {
		resp, err := iter.Next()
		if err == iterator.Done() {
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
```

In basic sense it's not different than calling an API.

## Prompting Models
Prompts are definitely the way you interact with models. There is whole field out there that is called **Prompt Engineering**, because one of the things that you can actually do, so there's a couple of way to customize or tweak a model or responses from a model.
- Fine Tune it.
  - Essentially add your own data to it and there are different algorithms.
- If you know on what a model is trained on, sort of corpus of text or whatever that these various models were trained on there'll be few hidden secrets as to how you actually prompt the model to get the responses that you want from it and most of these models can support **single-shot prompts** which is like one and done, send-in send-out. Others can do have these like sort of turns right where you can chitchat with the model and based on sort of context.

### Prompt Engineering
What is prompt engineering
- It's the process of creating prompts that are designed to elicit specific responses from an LLM.
- The goal is to improve the accuracy and fluency of the LLM's responses.
- It's more of an art than science.

Usecases of Prompt Engineering:
- We can use prompts to:
  - Achieve complex tasks that would have previously taken a huge amount of engineering time.
  - Achieve tasks that were previously impossible.
  - Be 10x productive.

### Designing a Prompt
Prompt:
- Preamble
  - Context
  - Instruction/Task
  - Example(s) (one -, few-shot)
- Input
  - Input (to predict to)

- Not all the components above are required for a prompt.
- The format depends on the task at hand.
- Order of the elements can also change.

Example:
- Preamble
  - VideoS is a video platform where users can add short comments. Comments can be positive, neutral, negative.
  - You need to classify the comments as positive, neutral, negative. Here are some examples.
  - Comment: This video is aweful...
  - The review is: Negative
  - Comment: No opinions about it...
  - The review is: Neutral
  - Comment: Loved it!
  - The reivew is: Positive
- Input
  - Comment: I don't know what to think about the video.
  - The reivew is:

Multimodal Models are models that take different types of inputs e.g. they can take, images, videos, text etc as prompts.

You can also give sample of response template that you want to come out.

#### CO-STAR Technique
It stands for:
- Context
- Objective
- Style
- Tone
- Audience
- Response

Let's revisit our previous example.

Models usually have two sets of instructions in turn based models:
- System Instructions 
  - Telling the model kind of how to act as.
- User Instructions
  - Sort of inputs that a user provides.

So in the previous go code let's just add a system instruction:
```go
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
```

All of a sudden just by giving it a different role we are able to create a different response because now it has this additional context.

Nice part of doing all this in code is, before you make something like a chat application, which takes different inputs and stream things back and forth you can basically simulate these conversations or these turn-based models right all in code right in your sort of workflow and figureout how things work and then workout to build how you're going to get those inputs or various outputs from the user in the system.

Beauty of Go is we can also enforce format of our input.

# The LLM Stack
In the early days, there were only models.

We called them **Prediction Model**, you send it some input, it gives you some outputs.

<img width="426" alt="Screenshot 2025-04-14 at 7 42 33 PM" src="https://github.com/user-attachments/assets/942ef1cf-a7a0-416d-86c4-4d0618080eb5" />

**Then, retrieval augumented generation (RAG) emerged.**
Where we provided additional context using RAG. So now I not only have my model but I have interface to Vector DBs and other things for more information.

<img width="551" alt="Screenshot 2025-04-14 at 7 44 37 PM" src="https://github.com/user-attachments/assets/460ff0cb-e06c-49a5-8298-e045140966e6" />

**...and evolved to AI Agents for reasoning and orchestration**
Agents are like fancy bots but they're much smarter because bots are mostly coded like hardcoded right, you are really putting different conditions in those bots, events come in and you prepare responses. 

<img width="691" alt="Screenshot 2025-04-14 at 7 46 55 PM" src="https://github.com/user-attachments/assets/335569e1-a6c0-496e-9a9d-afb0e45338c2" />

With AI theoretically we don't have to write all those conditions it should be able to adopt to those. Those become sort of AI agents, but in order to that it has to be given additional sort of context and information.

## Emrging LLM App Stack
With these many sorts of tools it might become nighmare to mange a monolithic app. So we need frameworks to sort of chaining these things or orchestrating these applications.

![Uploading Screenshot 2025-04-14 at 7.51.04 PM.png…]()

The most famous one out there is LangChain (for details look for Travis from Gophercon UK). Travis wrote Langchain in Go.

### Example 2: Gemini with LangChain Go
```go
func main() {
	genaiKey := os.Getenv("GOOGLE_GENAI_API_KEY")
	if genaiKey == "" {
		log.Fatal("please set GOOGLE_GENAI_API_KEY")
	}

	ctx := context.Background()

	llm, err := googleai.New(ctx, googleai.WithAPIKey(genaiKey))
	if err != nil {
		log.Fatal(err)
	}

	// Start by sending an initial question about the weather to the model, adding
	// "available tools" that include a getCurrentWeather function.
	// Thoroughout this sample, messageHistory collects the conversation history
	// with the model - this context is needed to ensure tool calling works
	// properly.
	messageHistory := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is the weather like in Chicago?"),
	}
	resp, err := llm.GenerateContent(ctx, messageHistory, llms.WithTools(availableTools))
	if err != nil {
		log.Fatal(err)
	}

	// Translate the model's response into a MessageContent element that can be
	// added to messageHistory.
	respchoice := resp.Choices[0]
	assistantResponse := llms.TextParts(llms.ChatMessageTypeAI, respchoice.Content)
	for _, tc := range respchoice.ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, tc)
	}
	messageHistory = append(messageHistory, assistantResponse)

	// "Execute" tool calls by calling requested function
	for _, tc := range respchoice.ToolCalls {
		switch tc.FunctionCall.Name {
		case "getCurrentWeather":
			var args struct {
				Location string `json:"location"`
			}
			if err := json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args); err != nil {
				log.Fatal(err)
			}
			if strings.Contains(args.Location, "Chicago") {
				toolResponse := llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							Name:    tc.FunctionCall.Name,
							Content: "64 and sunny",
						},
					},
				}
				messageHistory = append(messageHistory, toolResponse)
			}
		default:
			log.Fatalf("got unexpected function call: %v", tc.FunctionCall.Name)
		}
	}

	resp, err = llm.GenerateContent(ctx, messageHistory, llms.WithTools(availableTools))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response after tool call:")
	b, _ := json.MarshalIndent(resp.Choices[0], " ", "  ")
	fmt.Println(string(b))
}

// availableTools simulates the tools/functions we're making available for
// the model.
var availableTools = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getCurrentWeather",
			Description: "Get the current weather in a given location",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "The city and state, e.g. San Francisco, CA",
					},
				},
				"required": []string{"location"},
			},
		},
	},
}
```

# Serving Open Models
What if you wanted to serve your own model instead of using Gemini, OpenAI etc.

There are various open models:
- Llama
- Mistral
- Gemma
- Cohere
- Falcon
- etc.

These are not truly opesource models, they are more like Open Models because they don't actually give you the source to the training. If it was truly opensource you could recreate llama yourself, which you can't do. But what they do give us is the ability to use, they gives you the model and bunch of parameters for that model.

But how do we use these models:
## Model Serving
Model serving frameworks are able to take in the actual blob the big model artifact itself, some configuration information about that and then they can make that available over an endpoint.

<img width="364" alt="Screenshot 2025-04-14 at 11 09 40 PM" src="https://github.com/user-attachments/assets/4a79ae53-89ac-4712-9b37-131248ac4aff" />

What a number of these Serving Frameworks do is they basically standardize on the type of model. So we said that there is two main types in this world:
- Text Generation
- Stable Diffusion

So all the models listed above are all text generation, some of them are multimodal but they all have text generation interface.

So if we take this model and have a Model Serving Platform, we can just swap out models and always communicate through the same API. There are number of frameworks that are out there.
- Hugging Face's Text Generation inference
- Nvidia's Triton

But what if I want to run model on my own infrastructure?
Ollama is an open-source project that serves as a powerful user friendly platform for running LLMs on your local machine. It's written in Golang.

### Example 3 : Gemma With Ollama
Install ollama.

to check available models
```sh
ollama ls
```

pulling a model
```sh
ollama pull gemma2:2b
```

It also has chat interface which we can use using:
```sh
ollama run gemma2:2b
```

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ollama/ollama/api"
)

func main() {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	req := &api.GenerateRequest{
		Model:  "gemma2:2b",
		Prompt: "How many planets are there?",

		// set streaming to false
		Stream: new(bool),
	}

	ctx := context.Background()
	respFunc := func(resp api.GenerateResponse) error {
		// Only print the response here; GenerateResponse has number of other
		// interesting fields you want to examine
		fmt.Println(resp.Response)
		return nil
	}

	err = client.Generate(ctx, req, respFunc)
	if err != nil {
		log.Fatal(err)
	}
}
```

Nice thing about the interface here is that it's actually a callback streaming interface too. It'll continue to store the stream tokens right as they comeback rather than waiting for everything to get buffered.

# Firebase Genkit
Firebase got into the game with generative ai with a tool called **genkit**. 

<img width="984" alt="Screenshot 2025-04-14 at 7 51 04 PM" src="https://github.com/user-attachments/assets/b6bc81d0-a4ec-4ee3-a193-117845e7243e" />


Genkit's philosophy is that you shouldn't need to have, so in lots of cases the way these things work with LangChain and LangChain go while you can do it all in Go case, a lot of time you end up with this sort of middle tier server that handles your orchestration and that's kind of the way that like LangChain in the python world was sort of written.

You might endup with your web server something else that's in here your other logic and then you call out to these flows.

The idea that GenKit tried to go with is that you shouldn't have to have a separate orchestration layer, you should be able to embed kind of flows as functions as close to your business logic or your business application as possible right for you know simplicity, quick development etc. 

So they decided we are not going to do it that way:

<img width="691" alt="Screenshot 2025-04-15 at 10 02 27 AM" src="https://github.com/user-attachments/assets/2d186b2a-7ee9-4a92-9a33-ebac2111c2fa" />

We're going to try to make a more of a library that can be used more like inline or in your actual programs and so they focus on basically three particular areas:

- The ability for some of the apps itself. Basically everybody takes a plug-in architecture, this is how you kind of standardize on interfaces for common things and make plugins to call out to them different types of models, different types of tools.
- Then be able to have various sort of data pipelines for either you know injest or either post porcessing because you'll see that in some of these applications you will have to do a little bit of data processing, if you want to be able to use your own data within your apps, it won't be like retraining the model or fine-tuning the model but you want to be able to do that. So how can I build those in sort of native languages. 

## Genkit's Design Principles
- AI logic should live next to business logic.
- Prompts are code.
- Transparency all the way to the core.
- Simple, lean and lightweight library instead of a kitchen-sink of features
- Iterative AI development instead of waterfall.

It's opensource
Open ecosystem through plugins
Available in TypeScript and Go.

## Models
Models in Firebase Genkit are libraries and abstractions that provide access to various Google and non-Google LLMs.

```go
import (
    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/plugins/vertexai"
)

// Default to the value of GCLOUD_PROJECT for the project.
// and "us-central-1" for the location.
// To specify these values directly pass a vertexai.Config value to Init
if err := vertexai.Init(ctx, nil); err != nil {
    return err
}
model := vertexai.Model("gemini-1.5-flash")
```

## Flows
Flows are wrapped functions with some additional characteristics over direct calls: they are strongly typed, streamable, locally and remotely callable and fully observable.

```go
type MenuSuggestion struct {
    ItemName    string  `json:"item_name"`
    Description string  `json:"description"`
    Calories    int     `json:"calories"`
}

menuSuggestionFlow := genkit.DefineFlow(
    "menuSuggestionFlow",
    func(ctx context.Context, resturantTheme string) (MenuSuggestion, error) {
        suggestion := makeStructuredMenuItemSuggestion(resturantTheme)
        return suggestion, nil
    }
)
```

## Prompts
```go
func helloPrompt(name string) *ai.Part {
    prompt := fmt.Sprintf("You are a helpful AI assistant named Walt. Say hello to %s", name)
    return ai.NewTextPart(prompt)
}

response, err := ai.GenerateText(context.Background(), model, ai.WithMessages(ai.NewUserMessage(helloPrompt("Fred"))))
```

These are simple prompts like oneshot prompts, mentioned earlier, Simple Text String not a lot of stuff that's in here, may be a simple variable substitution.

You can also do these more complicated ones, where the prompts actually include context. The way firebase does it is pretty clever, instead of having to like write it all in code, you can actually provide these prompt files.

Dotprompt
```
---
model: vertexai/gemini-1.5-flash
config:
    temperature: 0.9
input:
    schema:
        location: string
        style?: string
        name?: string
    default:
        location: a resturant
---

You're the world's most welcoming AI assistant and are currently working at {{location}}.

Greet a guest {{#if name}} named {{name}}{{/if}}{{#if style}} in the style of {{style}}{{/if}}
```

You could actually write it all in code too becasue Go obviously support you know static strings. But essentially you kind of give it Prompt format, you give it a schema, you tell it model you want it to go to a couple of parameters that are in there. 

Then you can invoke it within GenKit. So pretty poweful:
- Take our **Models** 
- Take our **Flows**
  - Be able to call out functions
- Be able to make **Prompts**, dotprompts.

https://firebase.google.com/docs/genkit-go/get-started-go

```go
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
```

Why Models require hardware acceleration?
Most of it is like basic math, tensor functions, algebraic equations, floating point calculations which these GPUs are great at. Fundamentally that's what Nvidia, AMD accelerators do. Even Google's Tensor Processors. Fundamentally they all do the same thing hardware acceleration, you can offload those functions, you can run them faster, in parallel. Other key things is models themselves can take a ton of memory. So if you start to look at number of parameters in the models, in general unless you have some mega rigs you won't be able to run these big models on like a normal laptop.

The reason why Nvidia is winning in the game is because they developed the Cuda ecosystem. So Cuda is default way of interacting with gpus that every Library out there uses. So pytorch, ollama etc all use Cuda in the backend to communicate with the GPUs.
