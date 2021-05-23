# [Spec] Git repos in ticat

## A git repo in ticat
Any repo could added to **ticat** by:
```bash
$> ticat hub.add <repo-address>
```

A repo could provide 3 types of things to **ticat**:
* modules, will be register to command tree by the relative path in the repo
* flows, could be in any dir in the repo, will be register to command tree by it's file name
* sub-repo list, **ticat** will add sub-repos to hub automatically

## Repo tree
A repo could provide sub-repos, sub-repo could provide sub-sub-repos, so they could form a repo tree.

This is useful for developer to organize features.
And also useful for end-users to choose what to add to hub (eg, only add a branch).

## Sub-repo list defining format
A repo could define it's sub-repos in a specific file,
the file name is defined by env value "strs.repos-file-name", default value is "README.md"

In this file, every line after a special mark `[ticat.hub]` will consider as one repo,
The line format will be: `* [git-address](could-be-anything): help string of this repo`

**ticat** will also try to use part(the part after ":") of the first line(first line has a ":") in "README.me" as help string,
if no other place provides help string.
