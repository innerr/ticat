# import operating symbols

[-+]
	flow desc
[=]
	tail cmd usage
	[==] cmd detail
[/] find
	global find cmds
	[//] with detail
[~] sub
	tail cmd branch find cmds
	[~~] with detail

# candidate symbols

[!]
	exe tail cmd
	[!!] exe tail cmd
[^]
	head (same as tail)
[@] tag
	tail cmd branch find tags
[=] env
	tail cmd branch find env key

[$]
	tail
	[$!] tail cmd usage
	[$!+] tail cmd detail

[%]
