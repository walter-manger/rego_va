package builtins

import (
	"errors"
	"fmt"
	"slices"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Identity struct {
	Id   string   `json:"id"`
	Kind string   `json:"kind"`
	Orgs []string `json:"orgs"` // Need to work on entitlements, this could likely be a trimmed (flat) access graph for now
}

type Resource struct {
	Id          string   `json:"id"`
	Kind        string   `json:"kind"`
	Owner       string   `json:"owner"`
	Permissions []string `json:"permissions"` // These are not set for resources (yet), but they are for shares
}

var identities = map[string]Identity{
	"USER-1":            {Id: "UUID-1", Kind: "USER", Orgs: []string{"UUID-1", "UUID-2"}},
	"UUID-1":            {Id: "UUID-1", Kind: "USER", Orgs: []string{"UUID-1", "UUID-2"}},
	"HOLDING_COMPANY-1": {Id: "UUID-2", Kind: "HOLDING_COMPANY", Orgs: []string{}},
	"UUID-2":            {Id: "UUID-2", Kind: "HOLDING_COMPANY", Orgs: []string{}},
	"ADVERTISER-1":      {Id: "UUID-3", Kind: "ADVERTISER", Orgs: []string{}},
	"UUID-3":            {Id: "UUID-3", Kind: "ADVERTISER", Orgs: []string{}},
}

var resources = map[string]Resource{
	"AUD-1":    {Id: "UUID-1", Kind: "AUDIENCE", Owner: "UUID-1", Permissions: []string{"DMP.AUDIENCE.READ"}},
	"UUID-1":   {Id: "UUID-1", Kind: "AUDIENCE", Owner: "UUID-1", Permissions: []string{"DMP.AUDIENCE.READ"}},
	"PIXEL-1":  {Id: "UUID-2", Kind: "PIXEL", Owner: "UUID-2", Permissions: []string{"DIGITAL.PIXEL.READ"}},
	"UUID-2":   {Id: "UUID-2", Kind: "PIXEL", Owner: "UUID-2", Permissions: []string{"DIGITAL.PIXEL.READ"}},
	"REPORT-1": {Id: "UUID-3", Kind: "REPORT", Owner: "UUID-3", Permissions: []string{"REPORTING.EMBEDS.READ"}},
	"UUID-3":   {Id: "UUID-3", Kind: "REPORT", Owner: "UUID-3", Permissions: []string{"REPORTING.EMBEDS.READ"}},
}

// RegisterIdentity looks for the identity of either a user or an org for a given ID
func RegisterIdentity() (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name: "va.v1.identity",
			Decl: types.NewFunction(types.Args(types.A), types.NewArray([]types.Type{
				types.S, // Could be something more substantial later
				types.S,
			}, nil)),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			var args struct {
				ID string `json:"id"`
			}

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if args.ID == "" {
				return nil, errors.New("You need an id parameter")
			}

			// Make a network (yes, gRPC!) call, but what will the context be?
			identity, found := identities[args.ID]

			if found {
				caser := cases.Title(language.English)
				kind := caser.String(identity.Kind)
				return ast.ArrayTerm(
					ast.StringTerm(identity.Id),
					ast.StringTerm(fmt.Sprintf("%s found with id '%s'", kind, args.ID)),
				), nil
			}

			return ast.ArrayTerm(
				ast.StringTerm(""),
				ast.StringTerm(fmt.Sprintf("Object not found with id '%s'", args.ID)),
			), nil

		}
}

// RegisterResource looks up a resource for a given ID
func RegisterResource() (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name: "va.v1.resource",
			Decl: types.NewFunction(types.Args(types.A), types.NewArray([]types.Type{
				types.S, // Could be something more substantial later
				types.S,
			}, nil)),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			var args struct {
				ID string `json:"id"`
			}

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if args.ID == "" {
				return nil, errors.New("You need an id parameter")
			}

			// Make a network (yes, gRPC!) call, but what will the context be?
			resource, found := resources[args.ID]

			if found {
				caser := cases.Title(language.English)
				kind := caser.String(resource.Kind)
				return ast.ArrayTerm(
					ast.StringTerm(resource.Id),
					ast.StringTerm(fmt.Sprintf("%s found with id '%s'", kind, args.ID)),
				), nil
			}

			return ast.ArrayTerm(
				ast.StringTerm(""),
				ast.StringTerm(fmt.Sprintf("Object not found with id '%s'", args.ID)),
			), nil
		}
}

// RegisterCheck is stolen from https://github.com/aserto-dev/topaz/blob/main/builtins/edge/ds/check.go#L25
func RegisterCheck() (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name: "va.v1.check",
			Decl: types.NewFunction(types.Args(types.A), types.NewArray([]types.Type{
				types.B,
				types.S,
			}, nil)),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			var args struct {
				ObjectType  string `json:"object_type"`
				ObjectId    string `json:"object_id"`
				Relation    string `json:"relation"`
				SubjectType string `json:"subject_type"`
				SubjectId   string `json:"subject_id"`
			}

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if args.ObjectType == "" {
				return nil, errors.New("object_type is required")
			}
			if args.ObjectId == "" {
				return nil, errors.New("object_id is required")
			}
			if args.Relation == "" {
				return nil, errors.New("relation is required")
			}
			if args.SubjectType == "" {
				return nil, errors.New("subject_type is required")
			}
			if args.SubjectId == "" {
				return nil, errors.New("subject_id is required")
			}

			// Make a network (yes, gRPC!) call, but what will the context be?
			resource, resourceFound := resources[args.ObjectId]

			if !resourceFound {
				return ast.ArrayTerm(
					ast.BooleanTerm(false),
					ast.StringTerm(fmt.Sprintf("%s not found with id '%s'", args.ObjectType, args.ObjectId)),
				), nil
			}

			// Make a network (yes, gRPC!) call, but what will the context be?
			// Do all of the weird access graph stuff, and maybe even the share stuff...
			// But if we stick to a sane interface for getting this information we can iterate on the underlying
			// implementation... in Go!
			identity, identityFound := identities[args.SubjectId]

			if !identityFound {
				return ast.ArrayTerm(
					ast.BooleanTerm(false),
					ast.StringTerm(fmt.Sprintf("%s not found with id '%s'", args.SubjectType, args.SubjectId)),
				), nil
			}

			if args.Relation == "owner" {
				if resource.Owner == identity.Id {
					return ast.ArrayTerm(
						ast.BooleanTerm(true),
						ast.StringTerm(fmt.Sprintf("%s '%s' is the owner of resource '%s'", identity.Kind, identity.Id, resource.Id)),
					), nil
				}
			}

			if args.Relation == "member" {
				if slices.Contains(identity.Orgs, resource.Owner) {
					return ast.ArrayTerm(
						ast.BooleanTerm(true),
						ast.StringTerm(fmt.Sprintf("%s '%s' is a member of resource '%s' through owner '%s'", identity.Kind, identity.Id, resource.Id, resource.Owner)),
					), nil
				}
			}

			return ast.ArrayTerm(
				ast.BooleanTerm(false),
				ast.StringTerm(fmt.Sprintf("default no access, check logs? (%v)", args)),
			), nil
		}
}
