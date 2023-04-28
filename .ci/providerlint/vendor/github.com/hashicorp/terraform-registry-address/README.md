# terraform-registry-address

This module enables parsing, comparison and canonical representation of
[Terraform Registry](https://registry.terraform.io/) **provider** addresses
(such as `registry.terraform.io/grafana/grafana` or `hashicorp/aws`)
and **module** addresses (such as `hashicorp/subnets/cidr`).

**Provider** addresses can be found in

 - [`terraform show -json <FILE>`](https://www.terraform.io/internals/json-format#configuration-representation) (`full_name`)
 - [`terraform version -json`](https://www.terraform.io/cli/commands/version#example) (`provider_selections`)
 - [`terraform providers schema -json`](https://www.terraform.io/cli/commands/providers/schema#providers-schema-representation) (keys of `provider_schemas`)
 - within `required_providers` block in Terraform configuration (`*.tf`)
 - Terraform [CLI configuration file](https://www.terraform.io/cli/config/config-file#provider-installation)
 - Plugin [reattach configurations](https://www.terraform.io/plugin/debugging#running-terraform-with-a-provider-in-debug-mode)

**Module** addresses can be found within `source` argument
of `module` block in Terraform configuration (`*.tf`)
and parts of the address (namespace and name) in the Registry API.

## Compatibility

The module assumes compatibility with Terraform v0.12 and later,
which have the mentioned JSON output produced by corresponding CLI flags.

We recommend carefully reading the [ambigouous provider addresses](#Ambiguous-Provider-Addresses)
section below which may impact versions `0.12` and `0.13`.

## Related Libraries

Other libraries which may help with consuming most of the above Terraform
outputs in automation:

 - [`hashicorp/terraform-exec`](https://github.com/hashicorp/terraform-exec)
 - [`hashicorp/terraform-json`](https://github.com/hashicorp/terraform-json)

## Usage

### Provider

```go
pAddr, err := ParseProviderSource("hashicorp/aws")
if err != nil {
	// deal with error
}

// pAddr == Provider{
//   Type:      "aws",
//   Namespace: "hashicorp",
//   Hostname:  DefaultProviderRegistryHost,
// }
```

### Module

```go
mAddr, err := ParseModuleSource("hashicorp/consul/aws//modules/consul-cluster")
if err != nil {
	// deal with error
}

// mAddr == Module{
//   Package: ModulePackage{
//     Host:         DefaultProviderRegistryHost,
//     Namespace:    "hashicorp",
//     Name:         "consul",
//     TargetSystem: "aws",
//   },
//   Subdir: "modules/consul-cluster",
// },
```

## Other Module Address Formats

Modules can also be sourced from [other sources](https://www.terraform.io/language/modules/sources)
and these other sources (outside of Terraform Registry)
have different address formats, such as `./local` or
`github.com/hashicorp/example`.

This library does _not_ recognize such other address formats
and it will return error upon parsing these.

## Ambiguous Provider Addresses

Qualified addresses with namespace (such as `hashicorp/aws`)
are used exclusively in all recent versions (`0.14+`) of Terraform.
If you only work with Terraform `v0.14.0+` configuration/output, you may
safely ignore the rest of this section and related part of the API.

There are a few types of ambiguous addresses you may comes accross:

 - Terraform `v0.12` uses "namespace-less address", such as `aws`.
 - Terraform `v0.13` may use `-` as a placeholder for the unknown namespace,
   resulting in address such as `-/aws`.
 - Terraform `v0.14+` _configuration_ still allows ambiguous providers
   through `provider "<NAME>" {}` block _without_ corresponding
   entry inside `required_providers`, but these providers are always
   resolved as `hashicorp/<NAME>` and all JSON outputs only use that
   resolved address.

Both ambiguous address formats are accepted by `ParseProviderSource()`

```go
pAddr, err := ParseProviderSource("aws")
if err != nil {
	// deal with error
}

// pAddr == Provider{
//   Type:      "aws",
//   Namespace: UnknownProviderNamespace,    // "?"
//   Hostname:  DefaultProviderRegistryHost, // "registry.terraform.io"
// }
pAddr.HasKnownNamespace() // == false
pAddr.IsLegacy() // == false
```
```go
pAddr, err := ParseProviderSource("-/aws")
if err != nil {
	// deal with error
}

// pAddr == Provider{
//   Type:      "aws",
//   Namespace: LegacyProviderNamespace,     // "-"
//   Hostname:  DefaultProviderRegistryHost, // "registry.terraform.io"
// }
pAddr.HasKnownNamespace() // == true
pAddr.IsLegacy() // == true
```

However `NewProvider()` will panic if you pass an empty namespace
or any placeholder indicating unknown namespace.

```go
NewProvider(DefaultProviderRegistryHost, "", "aws")  // panic
NewProvider(DefaultProviderRegistryHost, "-", "aws") // panic
NewProvider(DefaultProviderRegistryHost, "?", "aws") // panic
```

If you come across an ambiguous address, you should resolve
it to a fully qualified one and use that one instead.

### Resolving Ambiguous Address

The Registry API provides the safest way of resolving an ambiguous address.

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

When you cache results, ensure you have invalidation
mechanism in place as target (migrated) namespace may change.

#### `terraform` provider

Like any other legacy address `terraform` is also ambiguous. Such address may
(most unlikely) represent a custom-built provider called `terraform`,
or the now archived [`hashicorp/terraform` provider in the registry](https://registry.terraform.io/providers/hashicorp/terraform/latest),
or (most likely) the `terraform` provider built into 0.11+, which is
represented via a dedicated FQN of `terraform.io/builtin/terraform` in 0.13+.

You may be able to differentiate between these different providers if you
know the version of Terraform.

Alternatively you may just treat the address as the builtin provider,
i.e. assume all of its logic including schema is contained within
Terraform Core.

In such case you should construct the address in the following way
```go
pAddr := NewProvider(BuiltInProviderHost, BuiltInProviderNamespace, "terraform")
```
