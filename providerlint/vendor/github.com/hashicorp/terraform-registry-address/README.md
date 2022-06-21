# terraform-registry-address

This package helps with representation, comparison and parsing of
Terraform Registry addresses, such as
`registry.terraform.io/grafana/grafana` or `hashicorp/aws`.

The most common source of these addresses outside of Terraform Core
is JSON representation of state, plan, or schemas as obtained
via [`hashicorp/terraform-exec`](https://github.com/hashicorp/terraform-exec).

## Example

```go
p, err := ParseRawProviderSourceString("hashicorp/aws")
if err != nil {
	// deal with error
}

// p == Provider{
//   Type:      "aws",
//   Namespace: "hashicorp",
//   Hostname:  svchost.Hostname("registry.terraform.io"),
// }
```

## Legacy address

A legacy address is by itself (without more context) ambiguous.
For example `aws` may represent either the official `hashicorp/aws`
or just any custom-built provider called `aws`.

Such ambiguous address can be produced by Terraform `<=0.12`. You can
just use `ImpliedProviderForUnqualifiedType` if you know for sure
the address was produced by an affected version.

If you do not have that context you should parse the string via
`ParseRawProviderSourceString` and then check `addr.IsLegacy()`.

### What to do with a legacy address?

Ask the Registry API whether and where the provider was moved to

(`-` represents the legacy, basically unknown namespace)

```sh
# grafana (redirected to its own namespace)
$ curl -s https://registry.terraform.io/v1/providers/-/grafana/versions | jq '(.id, .moved_to)'
"terraform-providers/grafana"
"grafana/grafana"

# aws (provider without redirection)
$ curl -s https://registry.terraform.io/v1/providers/-/aws/versions | jq '(.id, .moved_to)'
"hashicorp/aws"
null
```

Then:

 - Reparse the _new_ address (`moved_to`) of any _moved_ provider (e.g. `grafana/grafana`) via `ParseRawProviderSourceString`
 - Reparse the full address (`id`) of any other provider (e.g. `hashicorp/aws`)

Depending on context (legacy) `terraform` may need to be parsed separately.
Read more about this provider below.

If for some reason you cannot ask the Registry API you may also use
`ParseAndInferProviderSourceString` which assumes that any legacy address
(including `terraform`) belongs to the `hashicorp` namespace.

If you cache results (which you should), ensure you have invalidation
mechanism in place because target (migrated) namespace may change.
Hard-coding migrations anywhere in code is strongly discouraged.

### `terraform` provider

Like any other legacy address `terraform` is also ambiguous. Such address may
(most unlikely) represent a custom-built provider called `terraform`,
or the now archived [`hashicorp/terraform` provider in the registry](https://registry.terraform.io/providers/hashicorp/terraform/latest),
or (most likely) the `terraform` provider built into 0.12+, which is
represented via a dedicated FQN of `terraform.io/builtin/terraform` in 0.13+.

You may be able to differentiate between these different providers if you
know the version of Terraform.

Alternatively you may just treat the address as the builtin provider,
i.e. assume all of its logic including schema is contained within
Terraform Core.

In such case you should just use `NewBuiltInProvider("terraform")`.
