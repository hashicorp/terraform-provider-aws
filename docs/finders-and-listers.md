<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

<!-- markdownlint-configure-file { "code-block-style": false } -->
# Finders and Listers

## Finders

A resource's _finder_ is the function called from the resource's Read handler that returns the current state of the resource from the AWS API. If the AWS API indicates that the resource no longer exists (for example it has been deleted outside of Terraform), the finder must return an error that returns `true` from the `retry.NotFound` function. The Read handler then implements logic to inform Terraform that the resource no longer exists.

For example

=== "Terraform Plugin Framework (Preferred)"

    ```go
	output, err := findDBShardGroupByID(ctx, conn, shardGroupID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading RDS Shard Group (%s)", shardGroupID), err.Error())

		return
	}
    ```

=== "Terraform Plugin SDK V2"

    ```go
	ap, err := findAccessPointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EFS Access Point (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Access Point (%s): %s", d.Id(), err)
	}
    ```

### Finders And Acceptance Tests

The finder is also usually used in the resource's [acceptance testing](running-and-writing-acceptance-tests.md) [`CheckDestroy`](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/testcase#checkdestroy) and `Exists` functions.

For example

```go
func testAccCheckAccessPointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EFSClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_efs_access_point" {
				continue
			}

			_, err := tfefs.FindAccessPointByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EFS Access Point %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessPointExists(ctx context.Context, t *testing.T, n string, v *awstypes.AccessPointDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EFSClient(ctx)

		output, err := tfefs.FindAccessPointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
```

To use the finder function in the acceptance test package it must be exported via `exports_test.go`.

For example

```go
// Exports for use in tests only.
var (
	...
	FindAccessPointByID = findAccessPointByID
	...
)
```

### Implementation Patterns

The standard pattern for finder implementation varies depending on whether an AWS service API can returns a single resource or multiple resources from a `Describe` (or `Get`) API.

#### Singular Finder

```go
func findClusterByName(ctx context.Context, conn *eks.Client, name string) (*awstypes.Cluster, error) {
	input := eks.DescribeClusterInput{
		Name: aws.String(name),
	}

	return findCluster(ctx, conn, &input)
}

func findCluster(ctx context.Context, conn *eks.Client, input *eks.DescribeClusterInput) (*awstypes.Cluster, error) {
	output, err := conn.DescribeCluster(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Cluster == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Cluster, nil
}
```

#### Multiple Finder

```go
func findAccessPoint(ctx context.Context, conn *efs.Client, input *efs.DescribeAccessPointsInput) (*awstypes.AccessPointDescription, error) {
	output, err := findAccessPoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAccessPoints(ctx context.Context, conn *efs.Client, input *efs.DescribeAccessPointsInput) ([]awstypes.AccessPointDescription, error) {
	var output []awstypes.AccessPointDescription

	pages := efs.NewDescribeAccessPointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.AccessPointNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.AccessPoints...)
	}

	return output, nil
}

func findAccessPointByID(ctx context.Context, conn *efs.Client, id string) (*awstypes.AccessPointDescription, error) {
	input := efs.DescribeAccessPointsInput{
		AccessPointId: aws.String(id),
	}

	output, err := findAccessPoint(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if state := output.LifeCycleState; state == awstypes.LifeCycleStateDeleted {
		return nil, &retry.NotFoundError{
			Message: string(state),
		}
	}

	return output, nil
}
```

* Checking for AWS resource-not-found errors and mapping to `retry.NotFound` errors is done at the lowest level (closest to the `Describe` API call).
* Checking for logical deletion (e.g checking a resource's current status) is done at the highest level (closest to the resource's Read handler).
