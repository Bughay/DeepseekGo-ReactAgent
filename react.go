package deepseek

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type AgentResponse struct {
	Reasoning   string
	Act         string
	Observation string
}

type Agent struct {
	SystemPrompt string
	UserPrompt   string
	Memory       []Message
	Tools        []Tool
	Registry     map[string]func(string) (string, error)
}

func (a *Agent) oneloop() (*AgentResponse, error) {
	toolsDesc, err := ToolsToLLMString()
	if err != nil {
		return nil, fmt.Errorf("load tools: %w", err)
	}

	fullSystemPrompt := fmt.Sprintf(`%s

Available tools:
%s

You must respond in this exact JSON format:
{
    "reasoning": "your step-by-step thinking about what to do",
    "act": "tool_name|arg1,arg2 OR finish|your_final_answer",
    "observation": ""
}

If you need a tool, use "act": "tool_name|arguments".
If you have the answer, use "act": "finish|your answer here".`,
		a.SystemPrompt, toolsDesc)

	messages := []Message{
		{Role: "system", Content: fullSystemPrompt},
	}
	messages = append(messages, a.Memory...)

	var rawResponse string
	var resp AgentResponse

	for retries := 0; retries < 3; retries++ {
		rawResponse, err = DeepseekOneshotJSON(messages, 0.2)
		if err != nil {
			fmt.Printf("DEBUG: API error on retry %d: %v\n", retries, err)
			time.Sleep(30 * time.Second)
			continue
		}

		fmt.Printf("DEBUG: Raw LLM response: %q\n", rawResponse)

		err = json.Unmarshal([]byte(rawResponse), &resp)
		if err == nil {
			return &resp, nil
		}
		fmt.Printf("DEBUG: JSON parse error on retry %d: %v\n", retries, err)
	}

	return nil, fmt.Errorf("failed after 3 retries: %w", err)
}

func (a *Agent) Run() (*AgentResponse, error) {
	// Step 1: Initialize memory with user prompt
	a.Memory = []Message{
		{Role: "user", Content: a.UserPrompt},
	}

	maxIterations := 10

	for i := 0; i < maxIterations; i++ {

		// Step 2: Call oneloop (it reads a.Memory internally)
		resp, err := a.oneloop()
		if err != nil {
			return nil, err
		}
		fmt.Printf("\n=== Step %d ===\n", i+1)
		fmt.Printf("Reasoning: %s\n", resp.Reasoning)
		fmt.Printf("Act: %s\n", resp.Act)

		// Step 3: Check if finished
		if strings.HasPrefix(resp.Act, "finish|") {
			resp.Observation = strings.TrimPrefix(resp.Act, "finish|")
			return resp, nil
		}

		// Step 4: Parse tool call
		parts := strings.SplitN(resp.Act, "|", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid act format: %s", resp.Act)
		}
		toolName, toolArgs := parts[0], parts[1]

		// Step 5: Find and execute tool
		observation := fmt.Sprintf("Tool not found: %s", toolName)
		if executeFunc, exists := a.Registry[toolName]; exists {
			result, err := executeFunc(toolArgs)
			if err != nil {
				observation = fmt.Sprintf("Error: %v", err)
			} else {
				observation = result
			}
		}

		assistantContent := fmt.Sprintf("Reasoning: %s\nAct: %s", resp.Reasoning, resp.Act)
		a.Memory = append(a.Memory,
			Message{Role: "assistant", Content: assistantContent},
			Message{Role: "user", Content: "Observation: " + observation},
		)
	}

	return nil, fmt.Errorf("max iterations reached")
}

func (a *Agent) PrintConversation() {
	fmt.Println("=== Conversation History ===")
	for _, msg := range a.Memory {
		fmt.Printf("[%s]: %s\n", strings.ToUpper(msg.Role), msg.Content)
	}
	fmt.Println("============================")
}
