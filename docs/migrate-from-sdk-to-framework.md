# Migrating from Terraform SDKv2 to Framework

With the introduction of [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) it will become necessary to migrate existing resources from SDKv2. The Provider currently implements both plugins so migration can be done at a resource level.

## Migration Tooling

Tooling has been created that will scaffold an existing resource into a Framework resource. This tool is meant to be used as a starting point so additional editing will be needed.

Build:

```console
make tfsdk2fw
```

Convert a resource:

The following pattern is used to generate a file:  `tfsdk2fw [-resource <resource-type>|-data-source <data-source-type>] <package-name> <name> <generated-file>`

Example:

```console
tfsdk2fw -resource aws_example_resource examplepackage ResourceName internal/service/examplepackage/resource_name_fw.go
```

## State Upgrade

Terraform Plugin Framework introduced `null` values, which differ from `zero` values. Since the Plugin SDKv2 marked both `null` and `zero` values as the same, it will be necessary to use the [State Upgrader](https://developer.hashicorp.com/terraform/plugin/framework/migrating/resources/state-upgrade).

### Custom Types

The Plugin Framework introduced [custom types](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/custom) that allow custom validation on basic types. The following attribute types will require a state upgrade to utilize these custom types.

- ARNs
- CIDR Blocks
- Duration
- Timestamps

## Tagging

Tagging in the Plugin Framework is done by implementing the `ModifyPlan()` method on a resource.

```go
func (r *resourceExampleResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}
```

## Testing

It is important to not cause any state diffs that result in breaking changes. Testing will check that the diff before, and after, the migration present no changes.

**Note**: `VersionConstraint` should be set to the most recently published version of the AWS Provider.

```go
func TestAccExampleResource_MigrateFromPluginSDK(t *testing.T) {
	ctx := acctest.Context(t)
	var example service.ExampleResourceOutput
	resourceName := "aws_example_resource.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, service.EndpointsID),
		CheckDestroy: testAccCheckExampleResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.21.0", // always use most recently published version of the Provider
					},
				},
				Config: testAccExampleResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleResourceExists(ctx, resourceName, &example),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccExampleResourceConfig_basic(rName),
				PlanOnly:                 true,
			},
		},
	})
}
```
