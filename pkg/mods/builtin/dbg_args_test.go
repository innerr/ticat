package builtin

import (
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

func TestDbgArgs(t *testing.T) {
	tests := []struct {
		name     string
		argv     model.ArgVals
		expected map[string]string
	}{
		{
			name: "basic args",
			argv: model.ArgVals{
				"arg1": model.ArgVal{Raw: "value1", Provided: true},
				"arg2": model.ArgVal{Raw: "value2", Provided: true},
			},
			expected: map[string]string{
				"arg1": "value1",
				"arg2": "value2",
			},
		},
		{
			name: "default values",
			argv: model.ArgVals{
				"str-val": model.ArgVal{Raw: "default-str", Provided: false},
				"int-val": model.ArgVal{Raw: "0", Provided: false},
			},
			expected: map[string]string{
				"str-val": "default-str",
				"int-val": "0",
			},
		},
		{
			name: "mixed provided and default",
			argv: model.ArgVals{
				"arg1":    model.ArgVal{Raw: "provided", Provided: true},
				"str-val": model.ArgVal{Raw: "default-str", Provided: false},
			},
			expected: map[string]string{
				"arg1":    "provided",
				"str-val": "default-str",
			},
		},
		{
			name: "special characters in values",
			argv: model.ArgVals{
				"arg1": model.ArgVal{Raw: "db=test", Provided: true},
				"arg2": model.ArgVal{Raw: "host=127.0.0.1", Provided: true},
			},
			expected: map[string]string{
				"arg1": "db=test",
				"arg2": "host=127.0.0.1",
			},
		},
		{
			name: "empty values",
			argv: model.ArgVals{
				"arg1": model.ArgVal{Raw: "", Provided: true},
				"arg2": model.ArgVal{Raw: "non-empty", Provided: true},
			},
			expected: map[string]string{
				"arg2": "non-empty",
			},
		},
		{
			name: "dots in values",
			argv: model.ArgVals{
				"arg1": model.ArgVal{Raw: "1.2.3.4", Provided: true},
				"arg2": model.ArgVal{Raw: "file.conf", Provided: true},
			},
			expected: map[string]string{
				"arg1": "1.2.3.4",
				"arg2": "file.conf",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for name, expectedVal := range tc.expected {
				val, ok := tc.argv[name]
				if !ok {
					t.Errorf("expected arg %s not found", name)
					continue
				}
				if val.Raw != expectedVal {
					t.Errorf("arg %s: expected [%s], got [%s]", name, expectedVal, val.Raw)
				}
			}
		})
	}
}

func TestDbgArgsTail(t *testing.T) {
	tests := []struct {
		name     string
		argv     model.ArgVals
		expected map[string]string
	}{
		{
			name: "tail mode args",
			argv: model.ArgVals{
				"arg1": model.ArgVal{Raw: "tail-value1", Provided: true},
				"arg2": model.ArgVal{Raw: "tail-value2", Provided: true},
			},
			expected: map[string]string{
				"arg1": "tail-value1",
				"arg2": "tail-value2",
			},
		},
		{
			name: "env-style args in tail mode",
			argv: model.ArgVals{
				"arg1": model.ArgVal{Raw: "db=mysql", Provided: true},
				"arg2": model.ArgVal{Raw: "host=127.0.0.1", Provided: true},
			},
			expected: map[string]string{
				"arg1": "db=mysql",
				"arg2": "host=127.0.0.1",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for name, expectedVal := range tc.expected {
				val, ok := tc.argv[name]
				if !ok {
					t.Errorf("expected arg %s not found", name)
					continue
				}
				if val.Raw != expectedVal {
					t.Errorf("arg %s: expected [%s], got [%s]", name, expectedVal, val.Raw)
				}
			}
		})
	}
}

func TestDbgArgsEnv(t *testing.T) {
	tests := []struct {
		name     string
		argv     model.ArgVals
		expected map[string]string
	}{
		{
			name: "env style args",
			argv: model.ArgVals{
				"db":   model.ArgVal{Raw: "mysql", Provided: true},
				"host": model.ArgVal{Raw: "127.0.0.1", Provided: true},
				"port": model.ArgVal{Raw: "4000", Provided: true},
			},
			expected: map[string]string{
				"db":   "mysql",
				"host": "127.0.0.1",
				"port": "4000",
			},
		},
		{
			name: "partial args with defaults",
			argv: model.ArgVals{
				"db": model.ArgVal{Raw: "tidb", Provided: true},
			},
			expected: map[string]string{
				"db": "tidb",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for name, expectedVal := range tc.expected {
				val, ok := tc.argv[name]
				if !ok {
					t.Errorf("expected arg %s not found", name)
					continue
				}
				if val.Raw != expectedVal {
					t.Errorf("arg %s: expected [%s], got [%s]", name, expectedVal, val.Raw)
				}
			}
		})
	}
}

func TestArgValsGetRaw(t *testing.T) {
	argv := model.ArgVals{
		"existing": model.ArgVal{Raw: "value", Provided: true},
	}

	if argv.GetRaw("existing") != "value" {
		t.Errorf("expected 'value', got '%s'", argv.GetRaw("existing"))
	}

	if argv.GetRaw("nonexistent") != "" {
		t.Errorf("expected empty string for nonexistent arg, got '%s'", argv.GetRaw("nonexistent"))
	}
}

func TestArgValsGetRawEx(t *testing.T) {
	argv := model.ArgVals{
		"existing": model.ArgVal{Raw: "value", Provided: true},
	}

	if argv.GetRawEx("existing", "default") != "value" {
		t.Errorf("expected 'value', got '%s'", argv.GetRawEx("existing", "default"))
	}

	if argv.GetRawEx("nonexistent", "default") != "default" {
		t.Errorf("expected 'default', got '%s'", argv.GetRawEx("nonexistent", "default"))
	}
}

func TestArgValsProvided(t *testing.T) {
	argv := model.ArgVals{
		"provided":   model.ArgVal{Raw: "value", Provided: true},
		"default":    model.ArgVal{Raw: "default-val", Provided: false},
		"empty-prov": model.ArgVal{Raw: "", Provided: true},
	}

	if !argv["provided"].Provided {
		t.Error("expected 'provided' to be marked as provided")
	}
	if argv["default"].Provided {
		t.Error("expected 'default' to NOT be marked as provided")
	}
	if !argv["empty-prov"].Provided {
		t.Error("expected 'empty-prov' to be marked as provided even with empty value")
	}
}
