echo "ticat dm : dbg.echo last                         =flow"
ticat dm : dbg.echo last
echo "-------------------"
read
echo "ticat dm ::dbg.echo first                        =tail-mode-flow"
ticat dm ::dbg.echo first
echo "-------------------"
read
echo "ticat dm ::dbg.echo first :-                     =recursive tail-mode-flow"
ticat dm ::dbg.echo first :-
echo "-------------------"
read
echo "ticat display.utf : e.ls                         =flow"
ticat display.utf : e.ls
echo "-------------------"
read
echo "ticat display.utf ::e.ls                         =tail-mode call (e.ls support it, and len(flow) == 2)"
ticat display.utf ::e.ls
echo "-------------------"
read
echo "ticat display.utf ::e.ls tip                     =mixed tail-mode-call"
ticat display.utf ::e.ls tip
echo "-------------------"
read
echo "ticat dm : display.utf ::e.ls sys                =tail-mode-flow"
ticat dm : display.utf ::e.ls sys
ticat dm : display.utf ::e.ls sys:-
echo "-------------------"
read
echo "ticat dm ::display.utf ::e.ls sys                =recursive tail-mode-flow"
ticat dm ::display.utf ::e.ls sys
ticat dm ::display.utf ::e.ls sys :-
echo "-------------------"
read
echo "ticat display.utf ::e.ls sys ::dbg.echo first    =recursive tail-mode-flow"
ticat display.utf ::e.ls sys ::dbg.echo first
ticat display.utf ::e.ls sys ::dbg.echo first :-
