package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccWavelengthCarrierGateway_basic(t *testing.T) {
	var v ec2.CarrierGateway
	resourceName := "aws_ec2_carrier_gateway.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWavelengthZoneAvailable(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CarrierGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CarrierGatewayConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CarrierGatewayExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`carrier-gateway/cagw-.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccWavelengthCarrierGateway_disappears(t *testing.T) {
	var v ec2.CarrierGateway
	resourceName := "aws_ec2_carrier_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWavelengthZoneAvailable(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CarrierGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CarrierGatewayConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CarrierGatewayExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceCarrierGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWavelengthCarrierGateway_tags(t *testing.T) {
	var v ec2.CarrierGateway
	resourceName := "aws_ec2_carrier_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWavelengthZoneAvailable(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CarrierGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CarrierGatewayConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CarrierGatewayExists(resourceName, &v),
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
				Config: testAccEc2CarrierGatewayConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CarrierGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEc2CarrierGatewayConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CarrierGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckEc2CarrierGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_carrier_gateway" {
			continue
		}

		out, err := tfec2.FindCarrierGatewayByID(conn, rs.Primary.ID)
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidCarrierGatewayIDNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if out == nil {
			continue
		}
		if state := aws.StringValue(out.State); state != ec2.CarrierGatewayStateDeleted {
			return fmt.Errorf("EC2 Carrier Gateway in incorrect state. Expected: %s, got: %s", ec2.CarrierGatewayStateDeleted, state)
		}

		return err
	}

	return nil
}

func testAccCheckEc2CarrierGatewayExists(n string, v *ec2.CarrierGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		out, err := tfec2.FindCarrierGatewayByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if out == nil {
			return fmt.Errorf("EC2 Carrier Gateway not found")
		}
		if state := aws.StringValue(out.State); state != ec2.CarrierGatewayStateAvailable {
			return fmt.Errorf("EC2 Carrier Gateway in incorrect state. Expected: %s, got: %s", ec2.CarrierGatewayStateAvailable, state)
		}

		*v = *out

		return nil
	}
}

func testAccPreCheckWavelengthZoneAvailable(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeAvailabilityZonesInput{
		Filters: tfec2.BuildAttributeFilterList(map[string]string{
			"zone-type":     "wavelength-zone",
			"opt-in-status": "opted-in",
		}),
	}

	output, err := tfec2.FindAvailabilityZones(conn, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if len(output) == 0 {
		t.Skip("skipping since no Wavelength Zones are available")
	}
}

func testAccEc2CarrierGatewayConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_carrier_gateway" "test" {
  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccEc2CarrierGatewayConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_carrier_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccEc2CarrierGatewayConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_carrier_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
