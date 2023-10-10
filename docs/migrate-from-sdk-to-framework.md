# Migrating from Terraform SDKv2 to Framework

With the introduction of [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) it will become necessary to migrate existing resources to from SDKv2. The Provider currently implements both plugins so migration can be done at a resource level.

## Migration Tooling

Tooling has been created that will scaffold an existing resource into a Framework resource. This tool is meant to be used as a starting point but additional editing will be needed.

Build:

```console
make tfsdk2f
```

Convert a resource:

The following pattern is used to generate a file:  `tfsdk2fw [-resource <resource-type>|-data-source <data-source-type>] <package-name> <name> <generated-file>`

Example

```console
tfsdk2fw -resource aws_example_resource examplepackage ResourceName internal/service/examplepackage/resource_name_fw.go
```

## State Migration

## Setting State

## Plan Modifiers