// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/devopsguru/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfdevopsguru "github.com/hashicorp/terraform-provider-aws/internal/service/devopsguru"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccResourceCollection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcecollection types.ResourceCollectionFilter
	resourceName := "aws_devopsguru_resource_collection.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCollectionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceCollectionExists(ctx, resourceName, &resourcecollection),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.ResourceCollectionTypeAwsService)),
					resource.TestCheckResourceAttr(resourceName, "cloudformation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cloudformation.0.stack_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cloudformation.0.stack_names.0", "*"),
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

func testAccResourceCollection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcecollection types.ResourceCollectionFilter
	resourceName := "aws_devopsguru_resource_collection.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCollectionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceCollectionExists(ctx, resourceName, &resourcecollection),
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, acctest.Provider, tfdevopsguru.ResourceResourceCollection, resourceName, resourceCollectionDisappearsStateFunc()),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccResourceCollection_cloudformation(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcecollection types.ResourceCollectionFilter
	resourceName := "aws_devopsguru_resource_collection.test"
	cfnStackResourceName := "aws_cloudformation_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCollectionConfig_cloudformation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceCollectionExists(ctx, resourceName, &resourcecollection),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.ResourceCollectionTypeAwsCloudFormation)),
					resource.TestCheckResourceAttr(resourceName, "cloudformation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cloudformation.0.stack_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudformation.0.stack_names.0", cfnStackResourceName, names.AttrName),
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

func testAccResourceCollection_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcecollection types.ResourceCollectionFilter
	resourceName := "aws_devopsguru_resource_collection.test"
	appBoundaryKey := "DevOps-Guru-tfacctest"
	tagValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCollectionConfig_tags(appBoundaryKey, tagValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceCollectionExists(ctx, resourceName, &resourcecollection),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.ResourceCollectionTypeAwsTags)),
					resource.TestCheckResourceAttr(resourceName, "tags.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.0.app_boundary_key", appBoundaryKey),
					resource.TestCheckResourceAttr(resourceName, "tags.0.tag_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.0.tag_values.0", tagValue),
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

func testAccResourceCollection_tagsAllResources(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcecollection types.ResourceCollectionFilter
	resourceName := "aws_devopsguru_resource_collection.test"
	appBoundaryKey := "DevOps-Guru-tfacctest"
	tagValue := "*" // To include all resources with the specified app boundary key

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCollectionConfig_tags(appBoundaryKey, tagValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceCollectionExists(ctx, resourceName, &resourcecollection),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.ResourceCollectionTypeAwsTags)),
					resource.TestCheckResourceAttr(resourceName, "tags.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.0.app_boundary_key", appBoundaryKey),
					resource.TestCheckResourceAttr(resourceName, "tags.0.tag_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.0.tag_values.0", tagValue),
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

func resourceCollectionDisappearsStateFunc() func(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
	return func(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
		if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root(names.AttrID), is.Attributes[names.AttrID])); err != nil {
			return err
		}

		// The delete operation requires passing in the configured array of stack names
		// with a "REMOVE" action. Manually construct the root cloudformation attribute
		// to match what is created by the _basic test configuration.
		var diags diag.Diagnostics
		attrType := map[string]attr.Type{"stack_names": fwtypes.ListType{ElemType: fwtypes.StringType}}
		obj := map[string]attr.Value{
			"stack_names": flex.FlattenFrameworkStringValueList(ctx, []string{"*"}),
		}
		objVal, d := fwtypes.ObjectValue(attrType, obj)
		diags.Append(d...)

		elemType := fwtypes.ObjectType{AttrTypes: attrType}
		listVal, d := fwtypes.ListValue(elemType, []attr.Value{objVal})
		diags.Append(d...)

		if diags.HasError() {
			return fwdiag.DiagnosticsError(diags)
		}

		if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root("cloudformation"), listVal)); err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckResourceCollectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DevOpsGuruClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devopsguru_resource_collection" {
				continue
			}

			_, err := tfdevopsguru.FindResourceCollectionByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) || tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DevOpsGuru, create.ErrActionCheckingDestroyed, tfdevopsguru.ResNameResourceCollection, rs.Primary.ID, err)
			}

			return create.Error(names.DevOpsGuru, create.ErrActionCheckingDestroyed, tfdevopsguru.ResNameResourceCollection, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResourceCollectionExists(ctx context.Context, name string, resourcecollection *types.ResourceCollectionFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameResourceCollection, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameResourceCollection, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DevOpsGuruClient(ctx)
		resp, err := tfdevopsguru.FindResourceCollectionByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameResourceCollection, rs.Primary.ID, err)
		}

		*resourcecollection = *resp

		return nil
	}
}

func testAccResourceCollectionConfig_basic() string {
	return `
resource "aws_devopsguru_resource_collection" "test" {
  type = "AWS_SERVICE"
  cloudformation {
    stack_names = ["*"]
  }
}
`
}

func testAccResourceCollectionConfig_cloudformation(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name       = %[1]q
  on_failure = "DO_NOTHING"

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Resources = {
      S3Bucket = {
        Type = "AWS::S3::Bucket"
      }
    }
  })
}

resource "aws_devopsguru_resource_collection" "test" {
  type = "AWS_CLOUD_FORMATION"
  cloudformation {
    stack_names = [aws_cloudformation_stack.test.name]
  }
}
`, rName)
}

func testAccResourceCollectionConfig_tags(appBoundaryKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_devopsguru_resource_collection" "test" {
  type = "AWS_TAGS"
  tags {
    app_boundary_key = %[1]q
    tag_values       = [%[2]q]
  }
}
`, appBoundaryKey, tagValue)
}
