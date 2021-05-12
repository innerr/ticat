# ticat
Cli components platform

## Why and how
TODO: here I have something to say, but not have the time yet

## Progess
```
****-  Cli framework
***--      Command line parsing. TODO: char-escaping
****-      Full context search
****-      Full abbrs supporting. TODO: extra abbrs manage
****-      Env framework. TODO: save or load from a tag
-----      Log and search
-----      Command history and search
****-  Mod framework
****-      Connector framework
***--      Args supporting. TODO: free args
****-      Mod-ticat interacting
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
-----      Help and abbrs
-----      Executing ad-hot help
-----      Flatten in executing and desc
*****  Hub framework
*****      Mod and flow sharing
*****      Authority control (by git now)
```

Risks
```
* Mods-ticat interacting, now is stdin/stderr
    - A mod can't easily read from tty
    - Stderr is occupied
    - How about ssh login?
* The connector protocal is not stable now, need a best practice
* Concurrent support?
```
