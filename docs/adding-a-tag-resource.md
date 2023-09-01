# Adding a New Tag Resource

Adding a tag resource, similar to the `aws_ecs_tag` resource, has its own implementation procedure since the resource code and initial acceptance testing functions are automatically generated. The rest of the resource acceptance testing and resource documentation must still be manually created.

- In `internal/generate`: Ensure the service is supported by all generators. Run `make gen` after any modifications.
- In `internal/service/{service}/generate.go`: Add the new `//go:generate` call with the correct generator directives. Run `make gen` after any modifications.
- In `internal/provider/provider.go`: Add the new resource.
- Run `make test` and ensure there are no failures.
- Create `internal/service/{service}/tag_gen_test.go` with initial acceptance testing similar to the following (where the parent resource is simple to provision):

```go

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/{Service}"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc{Service}Tag_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_{service}_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, {Service}.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheck{Service}TagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAcc{Service}TagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck{Service}TagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc{Service}Tag_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_{service}_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, {Service}.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheck{Service}TagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAcc{Service}TagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck{Service}TagExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, resourceAws{Service}Tag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAcc{Service}Tag_Value(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_{service}_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, {Service}.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheck{Service}TagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAcc{Service}TagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck{Service}TagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAcc{Service}TagConfig(rName, "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck{Service}TagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1updated"),
				),
			},
		},
	})
}

func testAcc{Service}TagConfig(rName string, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_{service}_{thing}" "test" {
  name = %[1]q

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_{service}_tag" "test" {
  resource_arn = aws_{service}_{thing}.test.arn
  key          = %[2]q
  value        = %[3]q
}
`, rName, key, value)
}
```

- Run `make testacc TESTS=TestAcc{Service}Tags_ PKG={Service}` and ensure there are no failures.
- Create `website/docs/r/{service}_tag.html.markdown` with initial documentation similar to the following:

``````markdown
---
subcategory: "{SERVICE}"
layout: "aws"
page_title: "AWS: aws_{service}_tag"
description: |-
  Manages an individual {SERVICE} resource tag
---

# Resource: aws_{service}_tag

Manages an individual {SERVICE} resource tag. This resource should only be used in cases where {SERVICE} resources are created outside Terraform (e.g., {SERVICE} {THING}s implicitly created by {OTHER SERVICE THING}).

~> **NOTE:** This tagging resource should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_{service}_{thing}` and `aws_{service}_tag` to manage tags of the same {SERVICE} {THING} will cause a perpetual difference where the `aws_{service}_{thing}` resource will try to remove the tag being added by the `aws_{service}_tag` resource.

~> **NOTE:** This tagging resource does not use the [provider `ignore_tags` configuration](/docs/providers/aws/index.html#ignore_tags).

## Example Usage

```terraform
resource "aws_{service}_tag" "example" {
  resource_arn = "..."
  key          = "Name"
  value        = "Hello World"
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required) ARN of the {SERVICE} resource to tag.
* `key` - (Required) Tag name.
* `value` - (Required) Tag value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - {SERVICE} resource identifier and key, separated by a comma (`,`)

## Import

Import `aws_{service}_tag` using the {SERVICE} resource identifier and key, separated by a comma (`,`). For example:

```console
$ terraform import aws_{service}_tag.example arn:aws:{service}:us-east-1:123456789012:{thing}/example,Name
```
``````
