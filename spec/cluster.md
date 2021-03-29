# Component: cluster
```
$> ticat new {cluster-name} [{host-list}](seperated by ',')
$> ticat stop [{cluster-name}]
$> ticat fstop [{cluster-name}] (fstop: fast stop, use kill to stop process)
$> ticat up [{cluster-name}]
$> ticat remove [{cluster-name}]
$> ticat bin [pd=path] [tikv=path] [tidb=path] ...

$> ticat list|ls
{cluster-name}[current][running] {cluster-name}[stopped] ...

$> ticat current|curr
{cluster-name}[current][running] {cluster-name}[stopped] ...

$> ticat current|curr {cluster-name}
{cluster-name}[current][running] {cluster-name}[stopped] ...

$> ticat burn [{current-name}]
```
