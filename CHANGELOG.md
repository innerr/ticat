# Change Log of ticat

All notable changes to this project are documented here.

## [1.4.0] - 2022-09-20
+ Highlights
  + Support `args.auto` in meta, to auto generate args and arg2env for cmd (#192 #199 #201 #203 and more)
  + Support `pack-subflow` to totally hide suflow (#221)
  + New command `hook.exit/error/done` for hooking a flow with system events, to do result report or other jobs easily (#183)
+ Other New Features
  + Support `hide-in-sessions-last=true` `quiet=true` flag in meta files (#186 #188)
  + Support sys var `RANDOM` in meta file template (#218)
  + New command `env.rm.prefix`: batchly delete key-values by prefix (#213)
  + New command `session.running.desc.monitor` to monitor running status
  + New command `bg.after-main.*`: enable/disable auto wait for bg task when main thread ends (#205)
  + New command `display.sensitive` to display sensitive env values or args, which are hide by default (#210)
  + Display optimizations: fix lots of display bugs and some enhancements (#184 #185 #190 and more)
+ Important bug fixs
  + Fix bugs when use `desc` on an executing/executed session
  + Fix bugs when use breakpoints on a command with both flow and script

## [1.3.1] - 2022-06-15
+ New Features
  + Support combined meta file (#178)
  + Support macro definition in meta file (#176)
  + Env snapshot manage toolbox `env.snapshot.*` (#171)
  + Add break-point command: `break.here`
+ Default env key-values
  + New: `sys.paths.cache` for cache path (eg: download files)
  + New: `sys.paths.data.shared` for shared data (eg: repos could be used by more than one commands)

## [1.2.1] - 2022-04-01
+ New Features
  + Run command selftests in parallel mode (#125)
  + Add command branch `api`; add session id and host ip to env (#128)
  + support quiet-error flag in meta (#129)
  + Command `repeat`: run a command many times (#139)
  + (Disabled, too many bugs) A command set for dynamic changing flow during executing (#127)
+ Usability
  + not display sensitive(eg: password) key-value or args (#115)
  + add blender.forest-mode: reset env on each command, but not reset on their subflows (#118)
  + move command to global: `dbg.break.*`, `dbg.step`, `dbg.delay` (#126)
  + add commands: list sessions by type (#120)
  + Display enhance and bug fixes
+ Compatibility Changes
  + Change init repo from innerr/marsh.ticat to ticat-mods/marsh (#132)
+ Bug Fixs
  + Fix some bugs with tail-mode

## [1.2.0] - 2021-12-29

+ New Features
  + Support breakpoints and more executing control
    + Command `dbg.break.at <cmd-list>` [#97](https://github.com/innerr/ticat/pull/97)
    + Command `dbg.break.at.begin`, step in/out/over [#106](https://github.com/innerr/ticat/pull/106)
  + Interactive mode with history and completion [#108](https://github.com/innerr/ticat/pull/108)
  + Executed sessions management
    + Write logs for each command, shows them in executed session [#95](https://github.com/innerr/ticat/pull/95)
    + Retry an failed session from error point [#93](https://github.com/innerr/ticat/pull/93)
    + Tracing env changes in an executed session [#91](https://github.com/innerr/ticat/pull/91)
    + Display executed session details [#86](https://github.com/innerr/ticat/pull/86)
    + List, add, remove sessions [#83](https://github.com/innerr/ticat/pull/83)
+ Usability
  + Redesign command sets: `desc` `cmds` and more, for locating commands easier [#98](https://github.com/innerr/ticat/pull/98)
  + Remove tail-edit-mode usage info from all help-strings and hints, treat it as a hidden hack mode [#98](https://github.com/innerr/ticat/pull/98)
+ Compatibility Changes
  + Remove all command alias `*.--`(alias of `*.clear`) `*.+`(alias of `*.increase`) `*.-`(alias of `*.decrease`)
  + Rename command `dbg.echo` to `echo`
  + Rename command set `verbose` `quiet` to `display.verbose` `display.quiet`
  + Relocate repo local storage dir: `.../<project> => .../<git-server>/<author>/<project>` [#112](https://github.com/innerr/ticat/pull/112)
  + All capical abbrs of commands and args are removed

## [1.0.1] - 2021-11-29

+ New Features
  + Support background task by sys-arg `%delay` [#82](https://github.com/innerr/ticat/pull/82)
  + Support auto-timer in meta files `*.ticat` `*.tiflow` [#81](https://github.com/innerr/ticat/pull/81)
  + Add builtin commands `timer.begin` `timer.elapsed` for time recording [#77](https://github.com/innerr/ticat/pull/77)

## [1.0.0] - 2021-08-25

+ Init version
