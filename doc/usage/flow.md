# Flow: be a pro user

## Save command sequence to a flow
```
$> ticat dummy : dummy : dummy : flow.save x
```
If the flow "x" already exists, there will be an overwriting confirming.

## Run a saved flow
```
$> ticat x
(execute dummy * 3)
```

## Save command sequence to a flow with longer path:
```
$> ticat dummy : dummy : dummy : flow.save aa.bb.cc
```

```
$> ticat x : desc.simple
$> ticat x : desc.flow
$> ticat x : desc.flow.simple
```

```
$> ticat flow.ls
$> ticat flow.help
$> ticat flow.rm
$> ticat flow.clear
$> ticat hub.mv
```


## Best practice: no env definition in flow
