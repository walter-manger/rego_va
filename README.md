# Rego VA

This is an experiment to see what kind of things having the ability to create custom builtins for Rego could help with. 

This experiment was the result of me watching some presentations and reading some other implementations of using abstract custom builtins for a better policy writing experience.

OPA out of the box is great for assertion, but without an easy way to give it data, we've had to resort to using `http.send`. `http.send` is a good starting point, but it quickly becomes a hastle when having to consider network responses outside of the highly desired `200 OK`. Abstracting these calls away from the policies and into custom builtins is already a win because we can control what is returned by the builtin, and let clients handle a small set of "reasons" why the data was or was not returned. Further more, taking some time to define a small set of these generic builtins would remove a lot of complexity from our policies.

Another reason to consider the custom builtin approach is that it gives us the ability to iterate on the underlying infrastructure while maintaining a standard interface -- no more churn on OPA functions!. In the beginning, we'd be hitting our own services which query Postgres tables and views. Later, we could come up with more efficient data access solutions, such as things like Redis or [BoltDB](https://github.com/boltdb/bolt) (as [Topaz](https://www.topaz.sh/docs/intro) uses)


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

## Now these are available in the policies

```rego
package authz

import rego.v1

default allow := {"allowed": false, "reason": "just no."}

allow := {"allowed": allowed, "reason": reason} if {
  [allowed, reason] := va.v1.check({
    "object_id": "UUID-2", 
    "object_type": "PIXEL", 
    "relation": "member", 
    "subject_id": "UUID-1", 
    "subject_type": "USER"
  })
}
```

## Next up -> Solidify VA Builtins

``` rego
# Return the identity of a user or an org
[identity, reason] := va.v1.identity({})

# Return a resource
[resource, reason] := va.v1.resource({})

# Check if a user has access to a specific resource through a relation
[is_allowed, reason] := va.v1.check({
    "object_id": "UUID-2", 
    "object_type": "PIXEL", 
    "relation": "member", 
    "subject_id": "UUID-1", 
    "subject_type": "USER" 
})

# Possible other
# Check if a user has access to a specific resource through a relation with specific permissions
[is_allowed, reason] := va.v1.check_permissions({
    "object_id": "UUID-2", 
    "object_type": "PIXEL", 
    "relation": "member", 
    "subject_id": "UUID-1", 
    "subject_type": "USER",
    "with_permissions": ["DIGITAL.PIXEL.READ"]
})

# In the case of evaluating if a user can create a new "object_type"
[is_allowed, reason] := va.v1.check_permissions({
    # "object_id": "UUID-2",  # new object won't have an id
    "object_type": "PIXEL", 
    "relation": "owner", 
    "subject_id": "UUID-1", 
    "subject_type": "USER",
    "with_permissions": ["DIGITAL.PIXEL.CREATE"]
})

# For filtering results (like Audiences does)

# Get a list of business entities that the subject has access to.
[entities, reason] := va.v1.graph({
    "object_types": ["HOLDING_COMPANY", "AGENCY", "ADVERTISER"], 
    "subject_id": "UUID-1", 
    "subject_type": "USER",
    "with_permissions": []
})

# Get a list of business entities that the subject has access to which have a set of permissions.
[entities, reason] := va.v1.graph({
    "object_types": ["HOLDING_COMPANY", "AGENCY", "ADVERTISER"], 
    "subject_id": "UUID-1", 
    "subject_type": "USER",
    "with_permissions": ["DIGITAL.PIXEL.READ", "DIGITAL.PIXEL.WRITE"]
})

# entities would later become an authz-filter
```

## Then, Writing a plugin to get a user (and maybe a resource) available

```rego
package authz

import rego.v1

default allow := {"allowed": false, "reason": "just no."}

allow := {"allowed": allowed, "reason": reason} if {
  [allowed, reason] := va.v1.check({
    "object_id": "UUID-2", 
    "object_type": "PIXEL", 
    "relation": "member", 
    "subject_id": input.user, # so we can do this instead of having to decode the jwt 
    "subject_type": "USER"
  })
}
```

It /can/ be done, as it is [over here](https://github.com/open-policy-agent/opa-envoy-plugin/blob/main/envoyauth/request.go#L30).
