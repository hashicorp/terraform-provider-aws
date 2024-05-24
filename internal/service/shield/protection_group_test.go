// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfshield "github.com/hashicorp/terraform-provider-aws/internal/service/shield"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccShieldProtectionGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation", string(awstypes.ProtectionGroupAggregationMax)),
					resource.TestCheckNoResourceAttr(resourceName, "members"),
					resource.TestCheckResourceAttr(resourceName, "pattern", string(awstypes.ProtectionGroupPatternAll)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
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

func TestAccShieldProtectionGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfshield.ResourceProtectionGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccShieldProtectionGroup_aggregation(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_aggregation(rName, string(awstypes.ProtectionGroupAggregationMean)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation", string(awstypes.ProtectionGroupAggregationMean)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProtectionGroupConfig_aggregation(rName, string(awstypes.ProtectionGroupAggregationSum)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation", string(awstypes.ProtectionGroupAggregationSum)),
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

func TestAccShieldProtectionGroup_members(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_members(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "pattern", string(awstypes.ProtectionGroupPatternArbitrary)),
					resource.TestCheckResourceAttr(resourceName, "members.#", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "members.0", "ec2", regexache.MustCompile(`eip-allocation/eipalloc-.+`)),
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

func TestAccShieldProtectionGroup_protectionGroupID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection_group.test"
	testID1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testID2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_basic(testID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "protection_group_id", testID1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProtectionGroupConfig_basic(testID2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "protection_group_id", testID2),
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

func TestAccShieldProtectionGroup_resourceType(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_resourceType(rName, string(awstypes.ProtectedResourceTypeElasticIpAllocation)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "pattern", string(awstypes.ProtectionGroupPatternByResourceType)),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceType, string(awstypes.ProtectedResourceTypeElasticIpAllocation)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProtectionGroupConfig_resourceType(rName, string(awstypes.ProtectedResourceTypeApplicationLoadBalancer)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "pattern", string(awstypes.ProtectionGroupPatternByResourceType)),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceType, string(awstypes.ProtectedResourceTypeApplicationLoadBalancer)),
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

func TestAccShieldProtectionGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProtectionGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccProtectionGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectionGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckProtectionGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_shield_protection_group" {
				continue
			}

			input := &shield.DescribeProtectionGroupInput{
				ProtectionGroupId: aws.String(rs.Primary.ID),
			}

			resp, err := conn.DescribeProtectionGroup(ctx, input)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			if resp != nil && resp.ProtectionGroup != nil && aws.ToString(resp.ProtectionGroup.ProtectionGroupId) == rs.Primary.ID {
				return fmt.Errorf("The Shield protection group with ID %v still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckProtectionGroupExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

		input := &shield.DescribeProtectionGroupInput{
			ProtectionGroupId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeProtectionGroup(ctx, input)

		return err
	}
}

func testAccProtectionGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_shield_protection_group" "test" {
  protection_group_id = "%s"
  aggregation         = "MAX"
  pattern             = "ALL"
}
`, rName)
}

func testAccProtectionGroupConfig_aggregation(rName string, aggregation string) string {
	return fmt.Sprintf(`
resource "aws_shield_protection_group" "test" {
  protection_group_id = "%[1]s"
  aggregation         = "%[2]s"
  pattern             = "ALL"
}
`, rName, aggregation)
}

func testAccProtectionGroupConfig_resourceType(rName string, resType string) string {
	return fmt.Sprintf(`
resource "aws_shield_protection_group" "test" {
  protection_group_id = "%[1]s"
  aggregation         = "MAX"
  pattern             = "BY_RESOURCE_TYPE"
  resource_type       = "%[2]s"
}
`, rName, resType)
}

func testAccProtectionGroupConfig_members(rName string) string {
	return acctest.ConfigCompose(testAccProtectionConfig_elasticIPAddress(rName), fmt.Sprintf(`
resource "aws_shield_protection_group" "test" {
  depends_on = [aws_shield_protection.test]

  protection_group_id = "%[1]s"
  aggregation         = "MAX"
  pattern             = "ARBITRARY"
  members             = ["arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:eip-allocation/${aws_eip.test.id}"]
}
`, rName))
}

func testAccProtectionGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_shield_protection_group" "test" {
  protection_group_id = %[1]q
  aggregation         = "MAX"
  pattern             = "ALL"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccProtectionGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_shield_protection_group" "test" {
  protection_group_id = %[1]q
  aggregation         = "MAX"
  pattern             = "ALL"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
