// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccGlobalAcceleratorCrossAccountAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v globalaccelerator.Attachment

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "aws_globalaccelerator_cross_account_attachment"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "principals.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "resource.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrossAccountAttachmentConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}
func TestAccGlobalAcceleratorCrossAccountAttachment_principals(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v globalaccelerator.Attachment
	rAccountID := sdkacctest.RandStringFromCharSet(12, "012346789")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "aws_globalaccelerator_cross_account_attachment"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_principals(rName, rAccountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckTypeSetElemAttr(resourceName, "principals.*", rAccountID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

/*
func TestAccGlobalAcceleratorCrossAccountAttachment_resources(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	partition := acctest.Partition()
	region := acctest.Region()
	alternateRegion := acctest.AlternateRegion()
	endpointID := fmt.Sprintf("arn:%s:ec2:%s:171405876253:elastic-ip/eipalloc-1234567890abcdef0", partition, region)
	endpointID2 := fmt.Sprintf("arn:%s:ec2:%s:171405876253:elastic-ip/eipalloc-1234567890abcdef1", partition, alternateRegion)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "aws_globalaccelerator_cross_account_attachment"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_resources(rName, []tfglobalaccelerator.ResourceData{
					{EndpointID: types.StringValue(endpointID), Region: types.StringValue(region)},
					{EndpointID: types.StringValue(endpointID2), Region: types.StringValue(alternateRegion)},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resources.*", map[string]string{
						"endpoint_id": endpointID,
						"region":      region,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resources.*", map[string]string{
						"endpoint_id": endpointID2,
						"region":      alternateRegion,
					}),
				),
			},
		},
	})
}
*/

// TODO: tags
func TestAccGlobalAcceleratorCrossAccountAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v globalaccelerator.Attachment

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "aws_globalaccelerator_cross_account_attachment"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfglobalaccelerator.ResourceCrossAccountAttachment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCrossAccountAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_globalaccelerator_cross_account_attachment" {
				continue
			}

			_, err := tfglobalaccelerator.FindCrossAccountAttachmentByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Global Accelerator ross-account Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCrossAccountAttachmentExists(ctx context.Context, n string, v *globalaccelerator.Attachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn(ctx)

		output, err := tfglobalaccelerator.FindCrossAccountAttachmentByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCrossAccountAttachmentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name = %[1]q
}
`, rName)
}

func testAccCrossAccountAttachmentConfig_principals(rName string, accountID string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name       = %[1]q
  principals = [%[2]q]
}
`, rName, accountID)
}

/*
func testAccCrossAccountAttachmentConfig_resources(rName string, resources []tfglobalaccelerator.ResourceData) string {
	var resourcesStr []string
	for _, r := range resources {
		resourcesStr = append(resourcesStr, fmt.Sprintf(`{ endpoint_id = "%s", region = "%s" }`, r.EndpointID.ValueString(), r.Region.ValueString()))
	}
	return fmt.Sprintf(`
resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name      = "%s"
  resources = [%s]
}
`, rName, strings.Join(resourcesStr, ", "))
}
*/
