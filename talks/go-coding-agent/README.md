https://youtu.be/ybBVOa572Tw?si=1xBqVdXaZFutH7ej

https://ampcode.com/notes/how-to-build-an-agent

https://github.com/ardanlabs/ai-training

We will be using ollama as model orchestrator. 
We will be using gpt-oss model.

The idea is we want to have a basic workflow. It's going to be CLI based.
- We are going to ask some question.
- We are going to get some type of response back.

This model does two things that not every model does:
- It has tooling support, which we need to make it work.
- It also has reasoning support, which is really cool.

Being able to see how model's reasoning about things , helps you a lot , when you are tyring to develop this stuff.

Now the model is going to send us some reasoning content first. When we deal with this open AI model, reasoning information comes into this special field called 
**reasoning**. But if we were to use **Qwen** model who also does reasoning, sometime referred to as thinking. It comes back in regular content that comes back
in a HTML tag.

So we have code to figure out, is this reasoning though the reasoning field or reasoning through the content.

<img width="419" height="143" alt="image" src="https://github.com/user-attachments/assets/d0feca32-492d-4b3a-a616-4c003502126e" />

Let's start with this first and make it work.

# Creating an AI Agent

