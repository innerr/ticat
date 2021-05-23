# [Spec] Ticat local store

## The local store dir
**Ticat**'s store dir is defined by env key "sys.paths.data",
the default value is the **ticat** binary path plus suffix ".data".

Users can pass new dir to **ticat** to modify the store dir:
(TODO: implement)
```bash
$> ticat {sys.paths.data=./mydir} : ...
```

## The "flows", "hub" and "sessions" dir
The flows/hub/sessions store dirs are all under store dir:
* "sys.paths.data"/flows
* "sys.paths.data"/hub
* "sys.paths.data"/sessions

There are env keys to change these dirs:
* "sys.paths.flows"
* "sys.paths.hub"
* "sys.paths.sessions"
(TODO: implement, now they are all only under store dir)
