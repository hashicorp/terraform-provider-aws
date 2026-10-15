<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Enhanced Region Support

Most AWS resources are Regional – they are created and exist in a single AWS Region, and to manage these resources the Terraform AWS Provider directs API calls to endpoints in the Region. The default AWS Region used to provision a resource using the provider is defined in the [provider configuration](https://developer.hashicorp.com/terraform/language/providers/configuration) used by the resource, either implicitly via [environment variables](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#environment-variables) or [shared configuration files](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#shared-configuration-and-credentials-files), or explicitly via the [`region` argument](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region). Version **6.0.0** of the Terraform AWS Provider introduces [Enhanced Region Support](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/enhanced-region-support), an additional top-level `region` argument which allows that resource to be managed in a Region other than the one defined in the provider configuration.

In the codebase, this feature is often referred to as "OverrideRegion" or "per-resource region override".

Every Regional resource, data source and ephemeral resource supports this feature transparently – the new top-level `region` argument does not need to be explicitly defined in the resource’s schema and the resource implementation does **not** need to be aware whether or not a resource-level Region override is in place.

## Effective Region

The effective Region is the value of the top-level `region` argument if configured or the Region defined in the provider configuration. The `Region` method on the provider’s global state object (the provider’s _meta_ object) can be called to obtain the effective Region.

=== "Terraform Plugin Framework (Preferred)"
    ```go
    region := r.Meta().Region(ctx)
    ```

=== "Terraform Plugin SDK V2"
    ```go
    region := meta.(*conns.AWSClient).Region(ctx)
    ```

## Model Structure

When using Terraform Plugin Framework, a resource's [model structure](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values#get-the-entire-configuration-plan-or-state) must correspond to all the attributes in the resource's [schema](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas). Although the top-level `region` argument is transparently injected into a resource's schema it must be explicitly added to the resource's model. This can be done by directly embedding the `framework.WithRegionModel` structure.

```go
type exampleResourceModel struct {
    framework.WithRegionModel
    // Fields corresponding to attributes declared in the Schema.
}
```

## Annotations

Overriding the default behavior of Enhanced Region Support is done via adding resource-level annotations.

Most resources and data sources do not need to override the default behavior; these annotations are for advanced or special cases.

[`make gen`](makefile-cheat-sheet.md) should be run after changing any annotations.

### Global resources in regional services

If a resource in a Regional service is global, i.e. the addition of a top-level `region` argument is redundant or confusing (e.g. a resource representing account-wide settings) then Enhanced Region Support can be disabled by adding the `@Region(global=true)` annotation to the resource.

```go
// @FrameworkResource("aws_something_example", name="Example")
// @Region(global=true)
func newExampleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
    return &resourceExample{}, nil
}

type exampleResourceModel struct {
    // Fields corresponding to attributes in the Schema.
    // No top-level region argument so don't embed framework.WithRegionModel.
}
```

### Suppress `region` argument validation

By default any configured value of the top-level `region` is validated as being in the configured [partition](https://docs.aws.amazon.com/whitepapers/latest/aws-fault-isolation-boundaries/partitions.html) as AWS IAM credentials are only valid for a single partition. If the argument's value need not be validated (e.g. a data source that looks up a well-known value per-Region), adding the `@Region(validateOverrideInPartition=false)` annotation suppresses validation. This is a rare requirement and should only be used for special cases.

```go
// @FrameworkDataSource("aws_something_example", name="Example")
// @Region(validateOverrideInPartition=false)
func newExampleDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
    return &exampleDataSource{}, nil
}
```

## Documentation

The top-level `region` argument should be added to a resource's argument reference documentation. The standard text is

```
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
```
