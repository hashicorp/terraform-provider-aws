// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontConnectionGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionGroup awstypes.ConnectionGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, t, resourceName, &connectionGroup),
					testAccCheckConnectionGroupExistsByRoutingEndpoint(ctx, t, resourceName, &connectionGroup),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "cloudfront", regexache.MustCompile(`connection-group/cg_[0-9A-Za-z]+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "ipv6_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "is_default", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_deployment",
					names.AttrStatus,
				},
			},
		},
	})
}

func TestAccCloudFrontConnectionGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionGroup awstypes.ConnectionGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, t, resourceName, &connectionGroup),
					testAccCheckConnectionGroupExistsByRoutingEndpoint(ctx, t, resourceName, &connectionGroup),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudfront.ResourceConnectionGroup, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontConnectionGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var connectionGroup awstypes.ConnectionGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, t, resourceName, &connectionGroup),
					testAccCheckConnectionGroupExistsByRoutingEndpoint(ctx, t, resourceName, &connectionGroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccConnectionGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, t, resourceName, &connectionGroup),
					testAccCheckConnectionGroupExistsByRoutingEndpoint(ctx, t, resourceName, &connectionGroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConnectionGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, t, resourceName, &connectionGroup),
					testAccCheckConnectionGroupExistsByRoutingEndpoint(ctx, t, resourceName, &connectionGroup),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_connection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionGroupConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionGroupExists(ctx, t, resourceName, &connectionGroup),
					testAccCheckConnectionGroupExistsByRoutingEndpoint(ctx, t, resourceName, &connectionGroup),
					resource.TestCheckResourceAttr(resourceName, "ipv6_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckConnectionGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_connection_group" {
				continue
			}

			_, err := tfcloudfront.FindConnectionGroupById(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckConnectionGroupExists(ctx context.Context, t *testing.T, n string, v *awstypes.ConnectionGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindConnectionGroupById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output.ConnectionGroup

		return nil
	}
}

func testAccCheckConnectionGroupExistsByRoutingEndpoint(ctx context.Context, t *testing.T, n string, v *awstypes.ConnectionGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindConnectionGroupByRoutingEndpoint(ctx, conn, rs.Primary.Attributes["routing_endpoint"])

		if err != nil {
			return err
		}

		*v = *output.ConnectionGroup

		return nil
	}
}

func testAccConnectionGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccConnectionGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccConnectionGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccConnectionGroupConfig_ipv6(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_connection_group" "test" {
  name         = %[1]q
  ipv6_enabled = false
}
`, rName)
}
