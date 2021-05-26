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
****-      Connector framework
***--      Args supporting. TODO: free args
****-      Mod-ticat interacting
-----      Dependencies checking
-----      Mod definition: map env-val to arg
*****      Support mod types:
*****          Builtin
*****          File by ext: python, golang
*****          Executable file
*****          Directory (include repo supporting)
****-      Executor
*****          Base executor
-----          Middle re-enter
-----          Intellegent
-----          Mocking
-----          Background running
-----  Flow framework
***--      Save, edit and execute flow
***--      Help and abbrs
**---      Executing ad-hot help
-----      Flatten in executing and desc
-----      Combine mods' props
-----          Args
-----          Dependencies
****-          Connectors
*****  Hub framework
*****      Mod and flow sharing
*****      Authority control (by git now)
```

## Risks
```
* The connector protocal is not stable now, need a best practice
    - May need: write-one-on-of(key1, key1)
* Concurrent support?
* Command path or abbrs confliction
```
