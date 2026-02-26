# DeepseekGo-ReactAgent
ReAct agent module to be implemented in Golang

Go React Agent with DeepSeek
This project implements a ReAct (Reasoning + Acting) agent in Go, powered by the DeepSeek API. The agent follows an iterative loop of reasoning, acting (calling tools), and observing results until it reaches a final answer.

Features
ReAct loop: The agent thinks step by step, decides which tool to use, executes it, and incorporates observations.

Tool integration: Define custom tools in a tools.json file. The agent can invoke them dynamically.

DeepSeek API: Uses DeepSeek's chat completions with JSON mode for structured outputs.

Retry logic: Automatically retries API calls on failure.

Conversation memory: Maintains the full interaction history.

Folder Structure
text
.
├── deepseek.go      # DeepSeek API client (oneshot calls)
├── prompts.go       # (Placeholder) for system/user prompts
├── react.go         # Agent implementation (ReAct loop)
├── tools.go         # Tool loading and formatting
└── README.md        # This file
Prerequisites
Go 1.18+ (for generics, though not heavily used)

A DeepSeek API key (get one at DeepSeek Platform)

tools.json defining the available tools (see below)

Installation
Clone the repository or copy the files into your project.

Install dependencies (only github.com/joho/godotenv for environment loading):

bash
go get github.com/joho/godotenv
Create a .env file in the root directory with your API key:

text
DEEPSEEKAPIKEY=your-api-key-here
Tool Configuration
Tools are defined in a tools.json file. Each tool follows the OpenAI function calling schema. Example:

json
[
  {
    "type": "function",
    "function": {
      "name": "calculator",
      "description": "Perform basic arithmetic",
      "parameters": {
        "type": "object",
        "properties": {
          "expression": {
            "type": "string",
            "description": "Math expression to evaluate (e.g., '2+2')"
          }
        },
        "required": ["expression"]
      }
    }
  },
  {
    "type": "function",
    "function": {
      "name": "weather",
      "description": "Get weather for a city",
      "parameters": {
        "type": "object",
        "properties": {
          "city": {
            "type": "string",
            "description": "City name"
          }
        },
        "required": ["city"]
      }
    }
  }
]
Usage
Creating an Agent
Create a new agent by specifying a system prompt, user prompt, and registering tool implementations.

go
package main

import (
    "fmt"
    "log"
    "yourmodule/deepseek" // adjust import path
)

func main() {
    // Load tools from JSON
    tools, err := deepseek.LoadToolsFromFile("tools.json")
    if err != nil {
        log.Fatal(err)
    }

    // Create a registry mapping tool names to actual functions
    registry := map[string]func(string) (string, error){
        "calculator": func(args string) (string, error) {
            // args is the raw argument string (e.g., "expression=2+2")
            // You should parse it properly; here's a simplified example
            return fmt.Sprintf("Result: %s", args), nil
        },
        "weather": func(args string) (string, error) {
            return "Sunny, 25°C", nil
        },
    }

    agent := &deepseek.Agent{
        SystemPrompt: "You are a helpful assistant with access to tools.",
        UserPrompt:   "What is the weather in Paris and calculate 15+27?",
        Tools:        tools,
        Registry:     registry,
    }

    // Run the agent
    response, err := agent.Run()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Final answer: %s\n", response.Observation)
    agent.PrintConversation()
}
Agent Loop Explanation
The agent initializes memory with the user prompt.

In each iteration:

It calls oneloop() which sends the conversation (including tool descriptions) to DeepSeek, requesting a JSON response with reasoning and act fields.

If act starts with finish|, the loop ends and returns the final answer.

Otherwise, it parses the tool name and arguments, executes the corresponding function from the registry, and obtains an observation.

The observation is appended to memory, and the loop continues.

After a maximum of 10 iterations, if no finish is reached, an error is returned.

API Functions
DeepseekOneshot(systemMessage, userMessage string, temperature float64) (string, error)
Simple one-shot completion (non-JSON mode).

DeepseekOneshotJSON(messages []Message, temperature float64) (string, error)
Completion that forces JSON output (used internally by the agent).

Customization
Prompts: You can store reusable prompts in prompts.go.

Tools: Extend the registry with any Go function that takes a string argument (the raw arguments from the model) and returns a string result or error. The model will call tools with arguments in the format arg1,arg2 (comma-separated) as specified in the prompt. You may want to implement a parser for structured arguments.

Limitations
The tool calling format is simplistic (pipe-separated). For complex arguments, consider implementing a more robust parsing strategy.

The agent uses a fixed retry mechanism (3 retries, 30s wait) that may not be optimal for all use cases.

No streaming support yet.

