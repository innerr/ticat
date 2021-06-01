# Zen: the choices we made

## Why use git repo to publish components?

Q: normally a component platform will build it's own center service,
why **ticat** just use git repos?

The reason is to build a better community.

### User-centered VS official-centered

Most platforms are official-centered:
* The authority publish componets.
* Users pull their needed componets.
In this model it's hard to publish user-owned componets.

Some platform provide mirror tools and allow users to establish their center service.
But notice, the tools and maybe some access setting still under authority's control.

With git repo as a publish service,
**ticat** try to build a user-centered model,
Users can decide what to add to local disk,
And an important thing: anyone could become a publisher without any cost:
* Fork and edit the repo.
* Use git/github's access setting.

With this loose-auth model, we hope to constructe an active and layered ecosystem.

The layered community means, a developer don't need to be pro to contribute (to a specific group),
he could write crappy code at the very beginning,
but still can be a publisher sharing his work and assemble with all the pro modules,
once he have some good pieces, adapting his work into a higher-level publisher is very easy.

With **ticat**, the group can provide a easy-to-use environment for developer to get to know the project
(by many runnable flows), and a smooth path from beginner to core coding.

By that, we recommend "better than now" rule,
his work no need to be good, just need to be better (than now).

In this way, the project could partly improve,
and the community can gradually evolve.
