package display

import (
	"testing"

	"github.com/innerr/ticat/pkg/core/model"
)

func TestFilterQuietCmds(t *testing.T) {
	tree := model.NewCmdTree(model.CmdTreeStrsForTest())

	quietSub := tree.GetOrAddSub("quiet")
	quietSub.RegEmptyCmd("quiet command").SetQuiet()

	newParsedCmd := func(name string, isQuiet bool) model.ParsedCmd {
		sub := tree.GetOrAddSub(name)
		if sub.Cmd() == nil {
			cic := sub.RegEmptyCmd("test command")
			if isQuiet {
				cic.SetQuiet()
			}
		}
		return model.ParsedCmd{
			Segments: []model.ParsedCmdSeg{
				{
					Matched: model.MatchedCmd{
						Name: name,
						Cmd:  sub,
					},
				},
			},
		}
	}

	t.Run("filter_first_quiet_cmd", func(t *testing.T) {
		env := model.NewEnv()
		flow := []model.ParsedCmd{newParsedCmd("quiet", true), newParsedCmd("normal", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 1)

		if len(newFlow) != 1 {
			t.Errorf("expected 1 command, got %d", len(newFlow))
		}
		if newIdx != 0 {
			t.Errorf("expected newIdx 0, got %d", newIdx)
		}
	})

	t.Run("filter_multiple_quiet_cmds", func(t *testing.T) {
		env := model.NewEnv()
		flow := []model.ParsedCmd{newParsedCmd("q1", true), newParsedCmd("q2", true), newParsedCmd("normal", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 2)

		if len(newFlow) != 1 {
			t.Errorf("expected 1 command, got %d", len(newFlow))
		}
		if newIdx != 0 {
			t.Errorf("expected newIdx 0, got %d", newIdx)
		}
	})

	t.Run("no_filter_when_display_quiet_enabled", func(t *testing.T) {
		env := model.NewEnv()
		env.SetBool("display.mod.quiet", true)
		flow := []model.ParsedCmd{newParsedCmd("quiet", true), newParsedCmd("normal", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 1)

		if len(newFlow) != 2 {
			t.Errorf("expected 2 commands, got %d", len(newFlow))
		}
		if newIdx != 1 {
			t.Errorf("expected newIdx 1, got %d", newIdx)
		}
	})

	t.Run("handle_empty_cmd", func(t *testing.T) {
		env := model.NewEnv()
		emptyCmd := model.ParsedCmd{}
		flow := []model.ParsedCmd{emptyCmd, newParsedCmd("n1", false), newParsedCmd("n2", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 2)

		if len(newFlow) != 2 {
			t.Errorf("expected 2 commands, got %d", len(newFlow))
		}
		if newIdx != 1 {
			t.Errorf("expected newIdx 1, got %d", newIdx)
		}
	})

	t.Run("all_quiet_cmds", func(t *testing.T) {
		env := model.NewEnv()
		flow := []model.ParsedCmd{newParsedCmd("q1", true), newParsedCmd("q2", true)}
		newFlow, newIdx := filterQuietCmds(env, flow, 0)

		if len(newFlow) != 0 {
			t.Errorf("expected 0 commands, got %d", len(newFlow))
		}
		if newIdx != 0 {
			t.Errorf("expected newIdx 0, got %d", newIdx)
		}
	})

	t.Run("middle_cmd_index", func(t *testing.T) {
		env := model.NewEnv()
		flow := []model.ParsedCmd{newParsedCmd("n1", false), newParsedCmd("n2", false), newParsedCmd("n3", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 1)

		if len(newFlow) != 3 {
			t.Errorf("expected 3 commands, got %d", len(newFlow))
		}
		if newIdx != 1 {
			t.Errorf("expected newIdx 1, got %d", newIdx)
		}
	})

	t.Run("quiet_cmd_in_middle", func(t *testing.T) {
		env := model.NewEnv()
		flow := []model.ParsedCmd{newParsedCmd("n1", false), newParsedCmd("quiet", true), newParsedCmd("n2", false)}
		newFlow, newIdx := filterQuietCmds(env, flow, 2)

		if len(newFlow) != 2 {
			t.Errorf("expected 2 commands, got %d", len(newFlow))
		}
		if newIdx != 1 {
			t.Errorf("expected newIdx 1, got %d", newIdx)
		}
	})
}
