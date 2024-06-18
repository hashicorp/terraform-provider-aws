# Terraform Plugin Migrations

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

This command creates a separate file that exists alongside the existing SDKv2 resource. Ultimately, the new file should replace the SDKv2 resource.

When done creating the resource using the Framework run `make gen` to remove the SDK resource and add the Framework resource to the list of generated service packages.

## State Upgrade

Terraform Plugin Framework introduced `null` values, which differ from `zero` values. Since the Plugin SDKv2 marked both `null` and `zero` values as the same, it will be necessary to use the [State Upgrader](https://developer.hashicorp.com/terraform/plugin/framework/migrating/resources/state-upgrade).

An example of a resource with an upgraded state, while migrating, can be found [here](https://github.com/hashicorp/terraform-provider-aws/blob/88447d09f85dc737597243b31c5d0c8e212d055b/internal/service/batch/job_queue.go#L330).

### Custom Types

The Plugin Framework introduced [custom types](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/custom) that allow custom validation on basic types. The following attribute types will require a state upgrade to utilize these custom types.

- ARNs
- CIDR Blocks
- Duration
- Timestamps

SDKv2 `ARN` attribute.

```go
func ResourceExampleResource {
    return &schema.Schema{
        "arn_attribute": {		
	        Type:         schema.TypeString,
	        Optional:     true,
	        ValidateFunc: verify.ValidARN,
        },
        
        // other schema attributes
    }
}
```

Framework `ARN` attribute.

```go
func (r *resourceExampleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
    return schema.Schema{
        "arn_attribute": schema.StringAttribute{
            CustomType: fwtypes.ARNType,
            Optional:   true,
            PlanModifiers: []planmodifier.String{
                stringplanmodifier.UseStateForUnknown(),
            },
        },

        // other schema attributes
    }
}
```

## Tagging

Tagging in the Plugin Framework is done by implementing the `ModifyPlan()` method on a resource.

```go
func (r *resourceExampleResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}
```

Transparent Tagging that is used in SDKv2 also applies to the Framework by using the `@Tags` decorator when defining the resource.

```go
// @FrameworkResource("aws_service_example", name="Example Resource")
// @Tags(identifierAttribute="arn")
func newResourceExampleResrouce(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := resourceExampleResource{}
	return &r, nil
}
```

## Testing

It is important to not cause any state diffs that result in breaking changes. Testing will check that the diff before and after the migration presents no changes.

!!! tip
    `VersionConstraint` should be set to the most recently published version of the AWS Provider.

```go
func TestAccExampleResource_MigrateFromPluginSDK(t *testing.T) {
	ctx := acctest.Context(t)
	var example service.ExampleResourceOutput
	resourceName := "aws_example_resource.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ExampleServiceID),
		CheckDestroy: testAccCheckExampleResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.23.0", // always use most recently published version of the Provider
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
