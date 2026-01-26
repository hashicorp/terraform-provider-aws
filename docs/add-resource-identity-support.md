<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

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
