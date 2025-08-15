# Adding Resource Identity to parameterized Resources

You are working on the [Terraform AWS Provider](https://github.com/hashicorp/terraform-provider-aws), specifically focused on adding [resource identity](https://developer.hashicorp.com/terraform/plugin/sdkv2/resources/identity) to Plugin SDKV2 resources whose identity is composed from multiple parameters (parameterized).
[This Github meta issue](https://github.com/hashicorp/terraform-provider-aws/issues/42983) contains details and sub-issues related to adding resource identity support.

When adding resource identity, a pull request may include all resources in a service or a single resource.
Follow the steps below to complete this task.

## 1. Prepare the branch

- The feature branch name should begin with `f-ri` and be suffixed with the name of the service being updated, e.g. `f-ri-elbv2`. If the current branch does not match this convention, create one.
- Ensure the feature branch is rebased with the `main` branch.

## 2. Add resource identity to each resource

The changes for each individual resource should be done in its own commit.
Use the following steps to add resource identity to an existing resource:

- Determine which arguments the resource identity is composed from. This may be a single argument mapping to an AWS-generated identifier, or a combination of multiple arguments. Check for places where the resource ID is set (e.g. `d.SetId(<value>)`) and infer the relevant parameters.
- Add an `@IdentityAttribute("<argument_name>")` annotation to the target resource. For resources where the ID is composed from multiple arguments, add one annotation for each argument.
- If the `id` attribute is set to the same value as an identity attribute, add an `@Testing(idAttrDuplicates="<argument_name>")` annotation.
- If the resource's test file uses a `CheckExists` helper function that accepts 3 parameters rather than 2 (you can check this in the resource's test file), add another annotation to the resource file in the format `// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types;types.TrustStore")`, but replacing the type with the correct one for the resource in question. The type should match the third parameter of the CheckExists function.
- Since we are newly adding identity to this resource, add an annotation indicating the most recent pre-identity version, e.g. `@Testing(preIdentityVersion="v6.3.0")`. Use `CHANGELOG.md` at the project root to determine the most recently released version (which will be the last before identity is added).
- Some resources will have an importer function defined.
    - If that function uses `schema.ImportStatePassthroughContext` as `StateContext` value then remove that importer function declaration as it is no longer necessary.
    - If a custom import function is defined, add a `// @CustomImport` annotation and include the following at the beginning the custom `StateContext` function:

```go
				identitySpec := importer.IdentitySpec(ctx)
				if err := importer.RegionalSingleParameterized(ctx, d, identitySpec, meta.(importer.AWSClient)); err != nil {
					return nil, err
				}
```

- If the service does not use generated tag tests, you will need to create template files in the `testdata/tmpl` directory. For each resource, create a file named `<resource>_tags.gtpl` (e.g., `trust_store_tags.gtpl`).
- Populate each template file with the configuration from the resource's `_basic` test. If populating from the `_basic` configuration, be sure to replace any string format directives (e.g. `name = %[1]q`) with a corresponding reference to a variable (e.g. `name = var.rName`).
- The generators will use the template files to generate the resource identity test configuration. These will be located in the `testdata` directory for the service. **Do not manually create test directories or files as they will be generated.**
- The region template must be included inside each resource block in the template files. Add it as the first line after the resource declaration:

```hcl
resource "aws_service_thing" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" }}
}
```

- If the resource already has a tags template declaration different than the example above, e.g. `{{- template "tags" . }}`, leave it unchanged.
- If the test configuration references an `aws_region` data source, the region template should also be embedded here.

```hcl
data "aws_region" "current" {
{{- template "region" }}
}
```

## 3. Generate and test the changes

- Run the generators for this service. This can be done with the following command (e.g. for the elbv2 package): `go generate ./internal/service/elbv2/...`. This will generate tests for Resource Identity and any required test files.
- Run the tests in this order:
    - First run the basic identity test: `make testacc PKG=<service> TESTS=TestAcc<Resource>_Identity_Basic`
    - Run all identity tests: `make testacc PKG=<service> TESTS=TestAcc<Resource>_Identity`
    - Finally, run all tests for the resource: `make testacc PKG=<service> TESTS=TestAcc<Resource>_`. **Always include the `PKG` parameter to properly scope the tests to the intended service package.**
- Ensure the template modifications have not introduced any structural changes that would fail `terraform fmt`. To verify, run `terraform fmt -recursive -check`, and confirm there is no output.
- If all the preceding steps complete successfully commit the changes with an appropriate message, e.g. `r/aws_lb_target_group: add resource identity`. Ensure the commit message body includes the results of the acceptance test run in the previous step.

Repeat steps 2 and 3 for each resource in the service. When all resources are complete, proceed to the next section.

## 4. Submit a pull request

**!!!Important!!!**: Ask for confirmation before proceeding with this step.

- Push the changes.
- Create a draft pull request with the following details:
    - Title: "Add parameterized resource identity to `<service-name>`", e.g. "Add parameterized resource identity to `elbv2`". If only a single resource is included, replace service-name with the full Terraform resource name.
    - Use the following template for the body. Be sure to replace the acceptance test results section with the results from the full acceptance test suite run.

```
### Description
Add resource identity to parameterized resources in `<service-name>`. This includes:

<list Terraform resource names here>

### Relations
Relates #42983
Relates #42988

### Output from Acceptance Testing

<insert acceptance test results here>

```

- Once the pull request is created, fetch the PR number to add changelog entries. Create a new file, `.changelog/<pr-number>.txt`, and include one enhancement entry per resource. Refer to `.changelog/43503.txt` for the appropriate formatting.
- Provide a summary of the completed changes.

## Common Issues and Troubleshooting

### Test Failures

- Ensure `PKG` parameter is included in test commands
- Verify template file names match exactly (`<resource>_tags.gtpl`)
- Check region template placement is inside resource blocks
- Don't create test directories manually - let the generator create them
- If a generated test panics because a `testAccCheck*Exists` helper function has incorrect arguments, add a `@Testing(existsType="")` annotation. NEVER modify the function signature of an existing "exists" helper function

### Generator Issues

- Remove any manually created test directories before running the generator
- Ensure template files are in the correct location (`testdata/tmpl`)
- Verify template file names match the resource name
- If identity tests are not generated, verify that the `identitytests` generator is being called within the service's `generate.go` file. If it isn't, add the following line to `generate.go` next to the existing `go:generate` directives.
- If a generated test does not reference the `var.rName` variable, add an `// @Testing(generator=false)` annotation to remove it from the generated configuration.

```go
//go:generate go run ../../generate/identitytests/main.go
```

### Resource Updates

- Check if the resource's check exists helper takes 3 parameters
- Verify the correct type is used in the `existsType` annotation
- Ensure importer is only removed if using `ImportStatePassthroughContext`

### Import Test Failures

- If identity tests are failing because they expect an update during import but get a no-op, add an `// @Testing(plannableImportAction="NoOp")` annotation and re-generate the test files.
- If identity tests are failing import verification due to missing attribute values, check the `_basic` test implementation for the presence of an `ImportStateVerifyIgnore` field in the import test step. If present, add an `// @Testing(importIgnore="arg1")` annotation where `arg1` is replaced with the argument name(s) from the verify ignore slice. If mutiple fields are ignored, separate field names with a `;`, e.g. `arg1;arg2`.
- If a region override test is failing and a custom import fuction is configured, ensure the appropriate helper function from the `importer` package is used.
    - `RegionalSingleParameterized` - regional resources whose identity is made up of a single parameter.
    - `GlobalSingleParameterized` - global resources whose identity is made up of a single parameter.
    - `RegionalMultipleParameterized` - regional resources whose identity is made up of multiple parameters.
    - `GlobalMultipleParameterized` - global resources whose identity is made up of multiple parameters.
