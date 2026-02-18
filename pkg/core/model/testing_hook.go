package model

type TestingHook interface {
	OnBreakPoint(reason string, choices []string, actions map[string]string) string
	OnInteractPrompt(prompt string) (string, bool)
	ShouldSkipBash() bool
}

type TestingHookFuncs struct {
	BreakPointAction func(reason string, choices []string, actions map[string]string) string
	InteractPrompt   func(prompt string) (string, bool)
	SkipBash         bool
}

func (h *TestingHookFuncs) OnBreakPoint(reason string, choices []string, actions map[string]string) string {
	if h.BreakPointAction != nil {
		return h.BreakPointAction(reason, choices, actions)
	}
	return "c"
}

func (h *TestingHookFuncs) OnInteractPrompt(prompt string) (string, bool) {
	if h.InteractPrompt != nil {
		return h.InteractPrompt(prompt)
	}
	return "", false
}

func (h *TestingHookFuncs) ShouldSkipBash() bool {
	return h.SkipBash
}

type DefaultTestingHook struct{}

func (h *DefaultTestingHook) OnBreakPoint(reason string, choices []string, actions map[string]string) string {
	return "c"
}

func (h *DefaultTestingHook) OnInteractPrompt(prompt string) (string, bool) {
	return "", false
}

func (h *DefaultTestingHook) ShouldSkipBash() bool {
	return false
}
