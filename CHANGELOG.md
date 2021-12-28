# Change Log of ticat

All notable changes to this project are documented in this file.

## [1.1.1] - 2021-12-23

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

## [1.0.1] - 2021-11-29

+ New Features
  + Support background task by sys-arg `%delay` [#82](https://github.com/innerr/ticat/pull/82)
  + Support auto-timer in meta files `*.ticat` `*.tiflow` [#81](https://github.com/innerr/ticat/pull/81)
  + Add builtin commands `timer.begin` `timer.elapsed` for time recording [#77](https://github.com/innerr/ticat/pull/77)

## [1.0.0] - 2021-08-25

+ Init version
