package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2InternetGateway_basic(t *testing.T) {
	var v, v2 ec2.InternetGateway
	resourceName := "aws_internet_gateway.test"

	testNotEqual := func(*terraform.State) error {
		if len(v.Attachments) == 0 {
			return fmt.Errorf("IG A is not attached")
		}
		if len(v2.Attachments) == 0 {
			return fmt.Errorf("IG B is not attached")
		}

		id1 := v.Attachments[0].VpcId
		id2 := v2.Attachments[0].VpcId
		if id1 == id2 {
			return fmt.Errorf("Both attachment IDs are the same")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`internet-gateway/igw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInternetGatewayConfigChangeVPC,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v2),
					testNotEqual,
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccEC2InternetGateway_delete(t *testing.T) {
	var ig ec2.InternetGateway
	resourceName := "aws_internet_gateway.test"

	testDeleted := func(r string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			_, ok := s.RootModule().Resources[r]
			if ok {
				return fmt.Errorf("Internet Gateway %q should have been deleted", r)
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &ig)),
			},
			{
				Config: testAccNoInternetGatewayConfig,
				Check:  resource.ComposeTestCheckFunc(testDeleted(resourceName)),
			},
		},
	})
}

func TestAccEC2InternetGateway_tags(t *testing.T) {
	var v ec2.InternetGateway
	resourceName := "aws_internet_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v),
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
				Config: testAccInternetGatewayTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccInternetGatewayTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2InternetGateway_disappears(t *testing.T) {
	var v ec2.InternetGateway
	resourceName := "aws_internet_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceInternetGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInternetGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_internet_gateway" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			if len(resp.InternetGateways) > 0 {
				return fmt.Errorf("still exists")
			}

			return nil
		}

		// Verify the error is what we want
		if !tfawserr.ErrMessageContains(err, "InvalidInternetGatewayID.NotFound", "") {
			return err
		}
	}

	return nil
}

func testAccCheckInternetGatewayExists(n string, ig *ec2.InternetGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		resp, err := conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.InternetGateways) == 0 {
			return fmt.Errorf("InternetGateway not found")
		}

		*ig = *resp.InternetGateways[0]

		return nil
	}
}

const testAccNoInternetGatewayConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-no-internet-gateway"
  }
}
`

const testAccInternetGatewayConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-internet-gateway"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}
`

const testAccInternetGatewayConfigChangeVPC = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-internet-gateway-change-vpc"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-internet-gateway-change-vpc-other"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.bar.id

  tags = {
    Name = "terraform-testacc-internet-gateway-change-vpc-other"
  }
}
`

func testAccInternetGatewayTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccInternetGatewayTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
