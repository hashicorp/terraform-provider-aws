# Enhanced Region Support

Most AWS resources are Regional – they are created and exist in a single AWS Region, and to manage these resources the Terraform AWS Provider directs API calls to endpoints in the Region. The default AWS Region used to provision a resource using the provider is defined in the [provider configuration](https://developer.hashicorp.com/terraform/language/providers/configuration) used by the resource, either implicitly via [environment variables](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#environment-variables) or [shared configuration files](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#shared-configuration-and-credentials-files), or explicitly via the [`region` argument](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region). Version **6.0.0** of the Terraform AWS Provider introduces [Enhanced Region Support]((https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/enhanced-region-support#global-services)), an additional top-level `region` argument which allows that resource to be managed in a Region other than the one defined in the provider configuration.

Every Regional resource, data source and ephemeral resource supports this feature transparently – the new top-level `region` argument does not need to be explicitly defined in the resource’s schema and the resource implementation does need to be aware whether or not a resource-level Region override is in place.

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

When using Terraform Plugin Framework, a resource's [model structure](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values#get-the-entire-configuration-plan-or-state) must correspond to all the attributes in the resource's [schema](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas). Although the top-level `region` argument is transparently injected into a resource's schema it must be explictly added to the resource's model. This can be done by directly embedding the `framework.WithRegionModel` structure.

```go
type exampleResourceModel struct {
    framework.WithRegionModel
    // Fields corresponding to attributes declared in the Schema.
}
```

## Annotations

## Documentation
