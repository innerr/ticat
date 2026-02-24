package parser

import (
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

func TestCmdParserParseEnvWithDotsInKey(t *testing.T) {
	root := model.NewCmdTree(model.CmdTreeStrsForTest())
	fetch := root.AddSub("fetch")
	fetch.RegPowerCmd(func(argv model.ArgVals, cc *model.Cli, env *model.Env, flow *model.ParsedCmds, currCmdIdx int) (int, error) {
		return currCmdIdx, nil
	}, "fetch metrics").AddArg("cluster-id", "").AddArg("metric", "").AddArg("duration", "7d")

	parser := &CmdParser{
		&EnvParser{Brackets{"{", "}"}, "\t ", "=", ".", "%"},
		".", "./", "\t ", "<root>", "^", map[byte]bool{'/': true, '\\': true},
	}

	t.Run("env key with dots should not be mapped to positional arg", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{
			"fetch",
			"{cluster-id=123}",
			"{metric=test_metric}",
			"{duration=6h}",
			"{tidbcloud-insight.cache-dir=/tmp/test}",
		})

		if parsed.ParseResult.Error != nil {
			t.Fatalf("unexpected parse error: %v", parsed.ParseResult.Error)
		}

		if len(parsed.Segments) == 0 {
			t.Fatal("expected segments to be set")
		}
		lastSeg := parsed.Segments[len(parsed.Segments)-1]
		if lastSeg.Env == nil {
			t.Fatal("expected env to be set")
		}

		t.Logf("All env keys: %v", getEnvKeys(lastSeg.Env))

		if val, ok := lastSeg.Env["fetch.cluster-id"]; !ok || val.Val != "123" {
			t.Errorf("expected cluster-id=123, got %v", lastSeg.Env["fetch.cluster-id"])
		}

		found := false
		for k, v := range lastSeg.Env {
			if k == "tidbcloud-insight.cache-dir" || k == "fetch.tidbcloud-insight.cache-dir" {
				found = true
				if v.Val != "/tmp/test" {
					t.Errorf("expected tidbcloud-insight.cache-dir=/tmp/test, got %s", v.Val)
				}
			}
		}
		if !found {
			t.Errorf("expected tidbcloud-insight.cache-dir to be set, got keys: %v", getEnvKeys(lastSeg.Env))
		}
	})

	t.Run("multiple env keys with dots", func(t *testing.T) {
		parsed := parser.Parse(root, nil, []string{
			"fetch",
			"{cluster-id=456}",
			"{tidbcloud-insight.cache-dir=/tmp/a}",
			"{tidbcloud-insight.fetch.target-chunk-size-mb=8}",
		})

		if parsed.ParseResult.Error != nil {
			t.Fatalf("unexpected parse error: %v", parsed.ParseResult.Error)
		}

		lastSeg := parsed.Segments[len(parsed.Segments)-1]
		if lastSeg.Env == nil {
			t.Fatal("expected env to be set")
		}

		t.Logf("All env keys: %v", getEnvKeys(lastSeg.Env))

		hasCacheDir := false
		hasChunkSize := false
		for k, v := range lastSeg.Env {
			if k == "tidbcloud-insight.cache-dir" || stringsHasSuffix(k, "tidbcloud-insight.cache-dir") {
				hasCacheDir = true
				if v.Val != "/tmp/a" {
					t.Errorf("cache-dir: expected /tmp/a, got %s", v.Val)
				}
			}
			if k == "tidbcloud-insight.fetch.target-chunk-size-mb" || stringsHasSuffix(k, "tidbcloud-insight.fetch.target-chunk-size-mb") {
				hasChunkSize = true
				if v.Val != "8" {
					t.Errorf("chunk-size: expected 8, got %s", v.Val)
				}
			}
		}

		if !hasCacheDir {
			t.Error("missing tidbcloud-insight.cache-dir in env")
		}
		if !hasChunkSize {
			t.Error("missing tidbcloud-insight.fetch.target-chunk-size-mb in env")
		}
	})
}

func stringsHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
