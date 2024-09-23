# Rego VA

This is an experiment to see what kind of things having the ability to create custom builtins for Rego could help with. 

It is quite a hassle dealing with `http.send`, so abstracting these calls away from the policies is already a win.

This experiment was the result of me watching some presentations and reading some other implementations of using abstract custom builtins for a better policy writing experience.

## Sources:

- [Chime: Getting OPA the Data it Craves with Custom Rego Batch Loading Functions](https://www.youtube.com/watch?v=qHvh7ilYGQk)
- [Customizing OPA for a "Perfect Fit" Authorization Sidecar - Patrick East, Styra](https://www.youtube.com/watch?v=uCra4Uq9bCM)
- [Topaz Custom Builtins](https://github.com/aserto-dev/topaz/tree/main/builtins/edge/ds)

## Running The Repl

```sh
go run main.go run
OPA 0.68.0 (commit , built at )

Run 'help' to see a list of commands and check for updates.
```

You'll be presented with the repl, you can use the three builtins.

Get an `identity` (user or organization), it currently returns an `id` and a `reason`

```sh
va.v1.identity({"id": "UUID-3"})

>> [
  "UUID-3",
  "Advertiser found with id 'UUID-3'"
]
```

Get an `resource` (of any type), it currently returns an `id` and a `reason`

```sh
va.v1.resource({"id": "UUID-2"})

>> [
  "UUID-2",
  "Pixel found with id 'UUID-2'"
]
```

Check to see if a `subject` has access to an `object` through either an `owner` or `member` relationship.

```sh
va.v1.check({"object_id": "UUID-1", "object_type": "AUDIENCE", "relation": "owner", "subject_id": "UUID-1", "subject_type": "USER"})

>> [
  true,
  "USER 'UUID-1' is the owner of resource 'UUID-1'"
]

va.v1.check({"object_id": "UUID-2", "object_type": "PIXEL", "relation": "member", "subject_id": "UUID-1", "subject_type": "USER"})

>> [
  true,
  "USER 'UUID-1' is a member of resource 'UUID-2' through owner 'UUID-2'"
]
```

> Yes, these builtins were indeed heavily inspired by Topaz
