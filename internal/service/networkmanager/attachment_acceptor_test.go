package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
)

// This test file serves as tests for aws_networkmanager_attachment_acceptor and the following attachment types
// aws_networkmanager_vpc_attachment
func TestAccNetworkManagerAttachmentAcceptor_vpcAttachmentBasic(t *testing.T) {
	resourceName := "aws_networkmanager_vpc_attachment.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"awscc": {
			Source:            "hashicorp/awscc",
			VersionConstraint: "0.29.0",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkConfig_basic("*", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", "false"),
				),
			},
			{
				Config: testAccCoreNetworkConfig_basic("0", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", "true"),
				),
			},
			{
				Config: testAccCoreNetworkConfig_basic("*", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", "false"),
				),
			},
			// Cannot currently update ipv6 on its own, must also update subnet_arn
			// {
			// 	Config: testAccCoreNetworkConfig_basic("*", true),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "2"),
			// 		resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", "true"),
			// 	),
			// },
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerAttachmentAcceptor_vpcAttachmentTags(t *testing.T) {
	resourceName := "aws_networkmanager_vpc_attachment.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"awscc": {
			Source:            "hashicorp/awscc",
			VersionConstraint: "0.29.0",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkConfig_oneTag("segment", "shared"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
				),
			},
			{
				Config: testAccCoreNetworkConfig_twoTag("segment", "shared", "Name", "test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test"),
				),
			},
			{
				Config: testAccCoreNetworkConfig_oneTag("segment", "shared"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
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

func testAccCheckVPCAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_vpc_attachment" {
			continue
		}

		_, err := tfnetworkmanager.FindVPCAttachmentByID(context.Background(), conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Network Manager Attachment ID %s still exists", rs.Primary.ID)
	}

	return nil
}

const TestAccVPCConfig_multipleSubnets = `
data "aws_availability_zones" "test" {}
data "aws_region" "test" {}

resource "aws_vpc" "test" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true
}

locals {
  ipv6_cidrs = cidrsubnets(aws_vpc.test.ipv6_cidr_block, 8, 8)
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
}

resource "aws_subnet" "test" {
  count = length(var.subnets)

  vpc_id            = aws_vpc.test.id
  cidr_block        = element(var.subnets, count.index)
  availability_zone = data.aws_availability_zones.test.names[count.index]

  assign_ipv6_address_on_creation = true
  ipv6_cidr_block                 = local.ipv6_cidrs[count.index]
}
`

const TestAccCoreNetworkConfig = `
resource "awscc_networkmanager_global_network" "test" {}

resource "awscc_networkmanager_core_network" "test" {
  global_network_id = awscc_networkmanager_global_network.test.id
  policy_document   = jsonencode(jsondecode(data.aws_networkmanager_core_network_policy_document.test.json))
}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = data.aws_region.test.name
      asn      = 64512
    }
  }

  segments {
    name                          = "shared"
    description                   = "SegmentForSharedServices"
    require_attachment_acceptance = true
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
    share_with = ["*"]
  }

  attachment_policies {
    rule_number     = 1
    condition_logic = "or"

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "segment"
      value    = "shared"
    }

    action {
      association_method = "constant"
      segment            = "shared"
    }
  }
}
`

func testAccCoreNetworkConfig_basic(azs string, ipv6Support bool) string {
	return TestAccVPCConfig_multipleSubnets +
		TestAccCoreNetworkConfig + fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = flatten([aws_subnet.test.%[1]s.arn])
  core_network_id = awscc_networkmanager_core_network.test.id
  vpc_arn         = aws_vpc.test.arn

  options {
    ipv6_support = %[2]t
  }
}

resource "aws_networkmanager_attachment_acceptor" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.test.id
  attachment_type = aws_networkmanager_vpc_attachment.test.attachment_type
}
`, azs, ipv6Support)
}

func testAccCoreNetworkConfig_oneTag(tagKey1, tagValue1 string) string {
	return TestAccVPCConfig_multipleSubnets +
		TestAccCoreNetworkConfig + fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = [aws_subnet.test[0].arn]
  core_network_id = awscc_networkmanager_core_network.test.id
  vpc_arn         = aws_vpc.test.arn

  options {
    ipv6_support = false
  }

  tags = {
    %[1]q = %[2]q
  }
}

resource "aws_networkmanager_attachment_acceptor" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.test.id
  attachment_type = aws_networkmanager_vpc_attachment.test.attachment_type
}
`, tagKey1, tagValue1)
}

func testAccCoreNetworkConfig_twoTag(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return TestAccVPCConfig_multipleSubnets +
		TestAccCoreNetworkConfig + fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = [aws_subnet.test[0].arn]
  core_network_id = awscc_networkmanager_core_network.test.id
  vpc_arn         = aws_vpc.test.arn

  options {
    ipv6_support = false
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}

resource "aws_networkmanager_attachment_acceptor" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.test.id
  attachment_type = aws_networkmanager_vpc_attachment.test.attachment_type
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
