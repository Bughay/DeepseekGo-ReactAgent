package deepseek

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}
type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

func LoadToolsFromFile(filename string) ([]Tool, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var tools []Tool
	if err := json.Unmarshal(data, &tools); err != nil {
		return nil, err
	}
	return tools, nil
}

func (t Tool) ToLLMString() string {
	var b strings.Builder

	// Tool name and description
	b.WriteString(fmt.Sprintf("Tool: %s\n", t.Function.Name))
	b.WriteString(fmt.Sprintf("Description: %s\n", t.Function.Description))

	// Parameters
	if len(t.Function.Parameters.Properties) > 0 {
		b.WriteString("Parameters:\n")
		for paramName, prop := range t.Function.Parameters.Properties {
			required := ""
			for _, r := range t.Function.Parameters.Required {
				if r == paramName {
					required = " (required)"
					break
				}
			}
			b.WriteString(fmt.Sprintf("  - %s: %s%s\n", paramName, prop.Type, required))
			if prop.Description != "" {
				b.WriteString(fmt.Sprintf("    %s\n", prop.Description))
			}
		}
	}

	return b.String()
}

func ToolsToLLMString() (string, error) {
	data, err := os.ReadFile("tools.json")
	if err != nil {
		return "", fmt.Errorf("read tools.json: %w", err)
	}

	var tools []Tool
	if err := json.Unmarshal(data, &tools); err != nil {
		return "", fmt.Errorf("unmarshal tools: %w", err)
	}

	var results []string
	for _, t := range tools {
		results = append(results, t.ToLLMString())
	}

	return strings.Join(results, "\n---\n"), nil
}
