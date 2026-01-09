package mcp

import (
	"encoding/json"
	"fmt"
)

// GenerateToolSchema generates a JSON schema for a tool from skill metadata
func GenerateToolSchema(name, description string, params map[string]interface{}) map[string]interface{} {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "The action to perform",
			},
		},
		"required": []string{"action"},
	}

	// Add custom properties if provided
	if params != nil {
		props := schema["properties"].(map[string]interface{})
		for k, v := range params {
			props[k] = v
		}
	}

	return schema
}

// ValidateToolCall validates a tool call against its schema
func ValidateToolCall(tool Tool, call ToolCallParams) error {
	// Basic validation
	if call.Name != tool.Name {
		return fmt.Errorf("tool name mismatch: expected %s, got %s", tool.Name, call.Name)
	}

	// Validate against schema if available
	if schema, ok := tool.InputSchema["properties"].(map[string]interface{}); ok {
		// Check required fields
		if required, ok := tool.InputSchema["required"].([]interface{}); ok {
			for _, reqField := range required {
				fieldName := reqField.(string)
				if _, exists := call.Arguments[fieldName]; !exists {
					return fmt.Errorf("required field missing: %s", fieldName)
				}
			}
		}

		// Validate field types (simplified)
		for fieldName, fieldValue := range call.Arguments {
			if fieldDef, ok := schema[fieldName].(map[string]interface{}); ok {
				expectedType := fieldDef["type"].(string)
				actualType := getJSONType(fieldValue)
				if expectedType != actualType {
					return fmt.Errorf("type mismatch for field %s: expected %s, got %s", fieldName, expectedType, actualType)
				}
			}
		}
	}

	return nil
}

// getJSONType returns the JSON type of a value
func getJSONType(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// SerializeToolCall serializes a tool call to JSON
func SerializeToolCall(call ToolCallParams) ([]byte, error) {
	return json.Marshal(call)
}

// DeserializeToolCall deserializes a tool call from JSON
func DeserializeToolCall(data []byte) (*ToolCallParams, error) {
	var call ToolCallParams
	if err := json.Unmarshal(data, &call); err != nil {
		return nil, fmt.Errorf("failed to deserialize tool call: %w", err)
	}
	return &call, nil
}



