# Roadmap and progress

## Progess
```
****-  Cli framework
***--      Command line parsing. TODO: char-escaping
-----          Rewrite the shitty parser
****-      Full context search
****-      Full abbrs supporting. TODO: extra abbrs manage
****-      Env framework. TODO: save or load from a tag
***--      Log and search
-----      Command history and search
****-  Mod framework
****-      Env-ops framework
***--          Env-ops dependencies checking
-----          "one-of" statment: write-one-of(key1, key1)
*----      Os-command dependencies:
-----          Auto install?
***--      Args supporting. TODO: free args
****-      Mod-ticat interacting
*****      Support mod types:
*****          Builtin
****-              Flags: power, quiet, priority. TODO: priority-level
*****          File by ext: python, golang, ...
*****          Executable file
*****          Directory (include repo supporting)
-----  Flow framework
****-      Executor
*****          Base executor
-----          Middle re-enter
-----          Intellegent interactive
-----          Auto mocking
-----          Background running
-----          Concurrent running
****-      Save, edit/remove flow
****-      Help and abbrs
**---      Executing ad-hot help
-----      Combine mods' props
-----          Args
-----          Dependencies
-----      Flatten in executing and desc
*****  Hub framework
*****      Mod and flow sharing
*****      Authority control (by git for now)
*----      Command path or abbrs confliction
-----  Module version manage
```
