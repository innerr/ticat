package model

import (
	"encoding/json"
	"fmt"
)

// TextFormatter is an optional interface that types can implement
// to provide custom text output when sys.output.format=text.
type TextFormatter interface {
	FormatText() string
}

// IsJsonOutputMode returns true if the env flag sys.output.format is set to "json".
func IsJsonOutputMode(env *Env) bool {
	return env.GetRaw("sys.output.format") == "json"
}

// Output writes structured data to the screen.
// When sys.output.format=json, it marshals data as JSON.
// When sys.output.format=text, it uses TextFormatter if available, otherwise fmt.Sprintf.
func Output(cc *Cli, env *Env, data any) error {
	if IsJsonOutputMode(env) {
		b, err := json.Marshal(data)
		if err != nil {
			errorObj := map[string]any{
				"error": fmt.Sprintf("marshal failed: %v", err),
			}
			errorJSON, marshalErr := json.Marshal(errorObj)
			if marshalErr != nil {
				return cc.Screen.Error("{\"error\":\"marshal failed\"}\n")
			}
			return cc.Screen.Error(string(errorJSON) + "\n")
		}
		return cc.Screen.Print(string(b) + "\n")
	}
	if tf, ok := data.(TextFormatter); ok {
		return cc.Screen.Print(tf.FormatText())
	}
	return cc.Screen.Print(fmt.Sprintf("%v\n", data))
}

// OutputError writes a structured error to the screen's error stream.
// When sys.output.format=json, it produces {"error": ..., "type": ..., "detail": ...}.
// When sys.output.format=text, it returns false so the caller can use normal error display.
func OutputError(cc *Cli, env *Env, errType string, err error, detail map[string]string) bool {
	if !IsJsonOutputMode(env) {
		return false
	}
	obj := map[string]any{
		"error": err.Error(),
		"type":  errType,
	}
	if detail != nil {
		obj["detail"] = detail
	}
	b, marshalErr := json.Marshal(obj)
	if marshalErr != nil {
		fallback := map[string]any{
			"error": err.Error(),
		}
		fallbackJSON, _ := json.Marshal(fallback)
		_ = cc.Screen.Error(string(fallbackJSON) + "\n")
		return true
	}
	_ = cc.Screen.Error(string(b) + "\n")
	return true
}
