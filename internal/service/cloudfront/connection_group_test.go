// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	ret "github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontConnectionGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionGroup awstypes.ConnectionGroup
	resourceName := "aws_cloudfront_connection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, resourceName, &connectionGroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "tftest"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "is_default", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
				),
			},
			{
				Config:            testAccConnectionGroupConfig_basic(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontConnectionGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionGroup awstypes.ConnectionGroup
	resourceName := "aws_cloudfront_connection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, resourceName, &connectionGroup),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceConnectionGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontConnectionGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionGroup awstypes.ConnectionGroup
	resourceName := "aws_cloudfront_connection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, resourceName, &connectionGroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccConnectionGroupConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, resourceName, &connectionGroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConnectionGroupConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, resourceName, &connectionGroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCloudFrontConnectionGroup_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionGroup awstypes.ConnectionGroup
	resourceName := "aws_cloudfront_connection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupConfig_ipv6(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, resourceName, &connectionGroup),
					resource.TestCheckResourceAttr(resourceName, "ipv6_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCloudFrontConnectionGroup_anycastIpList(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionGroup awstypes.ConnectionGroup
	resourceName := "aws_cloudfront_connection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupConfig_anycastIpList(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, resourceName, &connectionGroup),
					resource.TestCheckResourceAttr(resourceName, "anycast_ip_list_id", "aip_3jpJwsoxxDsGJLm3JnLdvG"),
				),
			},
		},
	})
}

func testAccCheckConnectionGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_connection_group" {
				continue
			}

			_, err := tfcloudfront.FindConnectionGroupById(ctx, conn, rs.Primary.ID)

			if ret.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Connection Group (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckConnectionGroupExists(ctx context.Context, n string, v *awstypes.ConnectionGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindConnectionGroupById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output.ConnectionGroup

		return nil
	}
}

func testAccConnectionGroupConfig_basic() string {
	return `
resource "aws_cloudfront_connection_group" "test" {
  name = "tftest"
}
`
}

func testAccConnectionGroupConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_group" "test" {
  name = "tagstest"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccConnectionGroupConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_group" "test" {
  name = "tagstest"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccConnectionGroupConfig_ipv6() string {
	return `
resource "aws_cloudfront_connection_group" "test" {
  name = "ipv6test"
  ipv6_enabled = false
}
`
}

func testAccConnectionGroupConfig_anycastIpList() string {
	return `
resource "aws_cloudfront_connection_group" "testip" {
  name = "iptest"
  anycast_ip_list_id = "aip_3jpJwsoxxDsGJLm3JnLdvG"
}
`
}
