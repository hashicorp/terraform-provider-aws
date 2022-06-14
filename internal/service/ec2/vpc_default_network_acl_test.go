package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccVPCDefaultNetworkACL_basic(t *testing.T) {
	var v ec2.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`network-acl/acl-.+`)),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
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

func TestAccVPCDefaultNetworkACL_basicIPv6VPC(t *testing.T) {
	var v ec2.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "0"),
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

func TestAccVPCDefaultNetworkACL_tags(t *testing.T) {
	var v ec2.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(resourceName, &v),
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
				Config: testAccVPCDefaultNetworkACLConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCDefaultNetworkACLConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCDefaultNetworkACL_Deny_ingress(t *testing.T) {
	var v ec2.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_denyIngress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"protocol":   "-1",
						"rule_no":    "100",
						"from_port":  "0",
						"to_port":    "0",
						"action":     "allow",
						"cidr_block": "0.0.0.0/0",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "0"),
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

func TestAccVPCDefaultNetworkACL_withIPv6Ingress(t *testing.T) {
	var v ec2.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_includingIPv6Rule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":        "-1",
						"rule_no":         "101",
						"from_port":       "0",
						"to_port":         "0",
						"action":          "allow",
						"ipv6_cidr_block": "::/0",
					}),
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

func TestAccVPCDefaultNetworkACL_subnetRemoval(t *testing.T) {
	var v ec2.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_subnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test1", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test2", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Here the Subnets have been removed from the Default Network ACL Config,
			// but have not been reassigned. The result is that the Subnets are still
			// there, and we have a non-empty plan
			{
				Config: testAccVPCDefaultNetworkACLConfig_subnetsRemove(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(resourceName, &v),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCDefaultNetworkACL_subnetReassign(t *testing.T) {
	var v ec2.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_subnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test1", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test2", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Here we've reassigned the subnets to a different ACL.
			// Without any otherwise association between the `aws_network_acl` and
			// `aws_default_network_acl` resources, we cannot guarantee that the
			// reassignment of the two subnets to the `aws_network_acl` will happen
			// before the update/read on the `aws_default_network_acl` resource.
			// Because of this, there could be a non-empty plan if a READ is done on
			// the default before the reassignment occurs on the other resource.
			//
			// For the sake of testing, here we introduce a depends_on attribute from
			// the default resource to the other acl resource, to ensure the latter's
			// update occurs first, and the former's READ will correctly read zero
			// subnets
			{
				Config: testAccVPCDefaultNetworkACLConfig_subnetsMove(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
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

func testAccCheckDefaultNetworkACLDestroy(s *terraform.State) error {
	// The default NACL is not deleted.
	return nil
}

func testAccCheckDefaultNetworkACLExists(n string, v *ec2.NetworkAcl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Default Network ACL ID is set: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindNetworkACLByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if !aws.BoolValue(output.IsDefault) {
			return fmt.Errorf("EC2 Network ACL %s is not a default NACL", rs.Primary.ID)
		}

		*v = *output

		return nil
	}
}

func testAccVPCDefaultNetworkACLConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id
}
`, rName)
}

func testAccVPCDefaultNetworkACLConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVPCDefaultNetworkACLConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVPCDefaultNetworkACLConfig_includingIPv6Rule(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  ingress {
    protocol        = -1
    rule_no         = 101
    action          = "allow"
    ipv6_cidr_block = "::/0"
    from_port       = 0
    to_port         = 0
  }
}
`, rName)
}

func testAccVPCDefaultNetworkACLConfig_denyIngress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  egress {
    protocol   = -1
    rule_no    = 100
    action     = "allow"
    cidr_block = "0.0.0.0/0"
    from_port  = 0
    to_port    = 0
  }
}
`, rName)
}

func testAccDefaultNetworkACLSubnetsBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCDefaultNetworkACLConfig_subnets(rName string) string {
	return acctest.ConfigCompose(testAccDefaultNetworkACLSubnetsBaseConfig(rName), fmt.Sprintf(`
resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
}
`, rName))
}

func testAccVPCDefaultNetworkACLConfig_subnetsRemove(rName string) string {
	return acctest.ConfigCompose(testAccDefaultNetworkACLSubnetsBaseConfig(rName), fmt.Sprintf(`
resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  depends_on = [aws_network_acl.test]
}
`, rName))
}

func testAccVPCDefaultNetworkACLConfig_subnetsMove(rName string) string {
	return acctest.ConfigCompose(testAccDefaultNetworkACLSubnetsBaseConfig(rName), fmt.Sprintf(`
resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  depends_on = [aws_network_acl.test]
}
`, rName))
}

func testAccVPCDefaultNetworkACLConfig_ipv6(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id
}
`, rName)
}
