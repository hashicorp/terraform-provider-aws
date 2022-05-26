package shield_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfshield "github.com/hashicorp/terraform-provider-aws/internal/service/shield"
)

func TestAccShieldProtectionGroup_basic(t *testing.T) {
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(shield.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, shield.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation", shield.ProtectionGroupAggregationMax),
					resource.TestCheckNoResourceAttr(resourceName, "members"),
					resource.TestCheckResourceAttr(resourceName, "pattern", shield.ProtectionGroupPatternAll),
					resource.TestCheckNoResourceAttr(resourceName, "resource_type"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
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
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(shield.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, shield.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfshield.ResourceProtectionGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccShieldProtectionGroup_aggregation(t *testing.T) {
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(shield.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, shield.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_aggregation(rName, shield.ProtectionGroupAggregationMean),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation", shield.ProtectionGroupAggregationMean),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProtectionGroupConfig_aggregation(rName, shield.ProtectionGroupAggregationSum),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation", shield.ProtectionGroupAggregationSum),
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
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(shield.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, shield.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_members(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "pattern", shield.ProtectionGroupPatternArbitrary),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "members.0", "ec2", regexp.MustCompile(`eip-allocation/eipalloc-.+`)),
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
	resourceName := "aws_shield_protection_group.test"
	testID1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testID2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(shield.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, shield.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_basic(testID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(resourceName),
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
					testAccCheckProtectionGroupExists(resourceName),
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
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(shield.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, shield.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_resourceType(rName, shield.ProtectedResourceTypeElasticIpAllocation),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "pattern", shield.ProtectionGroupPatternByResourceType),
					resource.TestCheckResourceAttr(resourceName, "resource_type", shield.ProtectedResourceTypeElasticIpAllocation),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProtectionGroupConfig_resourceType(rName, shield.ProtectedResourceTypeApplicationLoadBalancer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "pattern", shield.ProtectionGroupPatternByResourceType),
					resource.TestCheckResourceAttr(resourceName, "resource_type", shield.ProtectedResourceTypeApplicationLoadBalancer),
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
	resourceName := "aws_shield_protection_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(shield.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, shield.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionGroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProtectionGroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccProtectionGroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckProtectionGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_shield_protection_group" {
			continue
		}

		input := &shield.DescribeProtectionGroupInput{
			ProtectionGroupId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeProtectionGroup(input)

		if tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.ProtectionGroup != nil && aws.StringValue(resp.ProtectionGroup.ProtectionGroupId) == rs.Primary.ID {
			return fmt.Errorf("The Shield protection group with ID %v still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckProtectionGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldConn

		input := &shield.DescribeProtectionGroupInput{
			ProtectionGroupId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeProtectionGroup(input)

		if err != nil {
			return err
		}

		return nil
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
