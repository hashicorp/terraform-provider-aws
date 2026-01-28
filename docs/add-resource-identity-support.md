<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

<!-- markdownlint-configure-file { "code-block-style": false } -->
# Adding Resource Identity Support

!!! note

    This guide covers adding Resource Identity support to an existing resource type.
    To enable Resource Identity support when adding a new resource type, see the [Add a New Resource Type Guide](add-a-new-resource.md).

Adding Resource Identity support to an existing resource type is mostly managed using annotations and generated code.
It typically requires minimal changes, if any, to the existing code.

The [Resource Identity Reference](resource-identity.md) provides more detail on enabling Resource Identity on a resource type.

Follow these steps:

1. Identify which form of Resource Identity the resource type will use and add the appropriate annotation.
  The most common types of Resource Identity are ARN Identity and Parameterized Identity.

    * If the resource type has a Parameterized Identity with multiple attributes,
      it will need an Import ID Handler as described in the [Resource Identity Reference](resource-identity.md#multiple-identity-attributes).

1. Add testing annotations as appropriate for the resource type.
  For a full reference to annotations, see
  the [Resource Identity Reference](resource-identity.md#acceptance-testing)
  and [acceptance test generation documentation](acc-test-generation.md).

    * When adding Resource Identity to an existing resource type, you must always add the annotation `@Testing(preIdentityVersion="<version>")`,
      where version is the last version of the provider **before** Resource Identity is added to the resource type.
    * Typically, the `Exists` function used in tests returns a value from the AWS API.
      This references a Go type and package path with optional package alias, using the format `<package path>;[<package alias>;]<type>`.
    * If attribute differences need to be ignored during import tests,
      add the annotation `@Testing(importIgnore="...")` with a list of the attribute names separated by semi-colons (`;`).

1. Update import support:

    * For Plugin-Framework-based resource types:
        1. Add `framework.WithImportByIdentity` to the resource struct.
          If `framework.WithImportByID` is present, remove it.
        1. If the resource type has a custom `ImportState` function,
          if it simply calls `resource.ImportStatePassthroughID`, it can be removed.
          Otherwise, the custom `ImportState` function must be updated for Resource Identity as documented below.

    * For Plugin-SDK-based resource types:
        1. If the resource type uses `schema.ImportStatePassthroughContext` as the importer function, it can be removed.
          Otherwise, add the annotation `@CustomImport`
          and update the importer function for Resource Identity as documented below.

1. Create or update the template file for generated acceptance tests, at `testdata/tmpl/<resource file name>.gtpl`.
  If the file already exists, add the line `{{- template "region" }}` at the top of each resource and data source declaration in the file.
  If the file does not exist, create a minimal configuration that will create a resource of this type.
  Add the Go template directive `{{- template "region" }}` at the top of each resource and data source declaration in the file.
  If this resource type can be tagged, add the annotation Go template directive `{{- template "tags" . }}` as the last line in the resource declaration.
  Only the resource being tested should have tags.
  For more details on Terraform configurations for generated acceptance tests, see the [acceptance test generation documentation](acc-test-generation.md#terraform-configuration-templates-for-tests).

1. Run `go generate internal/service/<service>/generate.go` to update the resource type registration and generate acceptance tests for Resource Identity.

1. Run the acceptance tests for the resource type to ensure everything is functioning as expected.
  If needed, make adjustments to the annotations or test configuration and re-run `go generate internal/service/<service>/generate.go`.

1. Update resource import documentation following the templates below.

    * ARN Identity

        ``````markdown
        In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

        ```terraform
        import {
        to = <resource-name>.example
        identity = {
            "arn" = <example-arn-value>
        }
        }

        resource "<resource-name>" "example" {
        ### Configuration omitted for brevity ###
        }
        ```

        ### Identity Schema

        #### Required

        * `arn` (String) <description here>.
        ``````

    * Parameterized Identity

        ``````markdown
        In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

        ```terraform
        import {
        to = <resource-name>.example
        identity = {
            <required key/value pairs here>
        }
        }

        resource "<resource-name>" "example" {
        ### Configuration omitted for brevity ###
        }
        ```

        ### Identity Schema

        #### Required

        <required attributes here>

        #### Optional

        * `account_id` (String) AWS Account where this resource is managed.
        * `region` (String) Region where this resource is managed.
        ``````

## Custom Import Functions

The built-in import function, and Import ID Handler if defined, should handle parsing the import ID and assigning attributes from the import ID.
In some cases, additional attributes must be set when importing.

### Plugin Framework

Define an `ImportState` function on the resource type, similar to:

```go
func (r *thingResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	r.WithImportByIdentity.ImportState(ctx, request, response)

    // Set additional attributes here
}
```

For example, the resource type `aws_s3tables_table_bucket` sets a default value for the attribute `force_destroy` when importing.
The `ImportState` function is:

```go
func (r *tableBucketResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	r.WithImportByIdentity.ImportState(ctx, request, response)

	// Set force_destroy to false on import to prevent accidental deletion
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrForceDestroy), types.BoolValue(false))...)
}
```

### Plugin SDK

Define an `Importer` with a `StateContext` function on the resource schema, similar to:

```go
Importer: &schema.ResourceImporter{
    StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
         if err := importer.Import(ctx, rd, meta); err != nil {
            return nil, err
        }

        // Set additional attributes here

        return []*schema.ResourceData{rd}, nil
    },
},
```

For example, the resource type `aws_batch_job_definition` sets a default value for the attribute `deregister_on_new_revision` when importing.
The `Importer` is:

```go
Importer: &schema.ResourceImporter{
    StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
        if err := importer.Import(ctx, rd, meta); err != nil {
            return nil, err
        }

        rd.Set("deregister_on_new_revision", true)

        return []*schema.ResourceData{rd}, nil
    },
},
```
