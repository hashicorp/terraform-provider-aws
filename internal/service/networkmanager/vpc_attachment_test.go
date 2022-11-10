package networkmanager_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccNetworkManagerVPCAttachment_basic(t *testing.T) {
	var v networkmanager.VpcAttachment
	resourceName := "aws_networkmanager_vpc_attachment.test"
	coreNetworkResourceName := "awscc_networkmanager_core_network.test"
	vpcResourceName := "aws_vpc.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"awscc": {
			Source:            "hashicorp/awscc",
			VersionConstraint: "0.29.0",
		},
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAttachmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCAttachmentExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "networkmanager", regexp.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_policy_rule_number", "0"),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "VPC"),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_arn", coreNetworkResourceName, "core_network_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_id"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", "false"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", vpcResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "segment_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_arn", vpcResourceName, "arn"),
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

func TestAccNetworkManagerVPCAttachment_disappears(t *testing.T) {
	var v networkmanager.VpcAttachment
	resourceName := "aws_networkmanager_vpc_attachment.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"awscc": {
			Source:            "hashicorp/awscc",
			VersionConstraint: "0.29.0",
		},
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAttachmentExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfnetworkmanager.ResourceVPCAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerVPCAttachment_tags(t *testing.T) {
	var v networkmanager.VpcAttachment
	resourceName := "aws_networkmanager_vpc_attachment.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"awscc": {
			Source:            "hashicorp/awscc",
			VersionConstraint: "0.29.0",
		},
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAttachmentConfig_tags1(rName, "segment", "shared"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAttachmentExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
				),
			},
			{
				Config: testAccVPCAttachmentConfig_tags2(rName, "segment", "shared", "Name", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAttachmentExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test"),
				),
			},
			{
				Config: testAccVPCAttachmentConfig_tags1(rName, "segment", "shared"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAttachmentExists(resourceName, &v),
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

func TestAccNetworkManagerVPCAttachment_update(t *testing.T) {
	var v networkmanager.VpcAttachment
	resourceName := "aws_networkmanager_vpc_attachment.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"awscc": {
			Source:            "hashicorp/awscc",
			VersionConstraint: "0.29.0",
		},
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAttachmentConfig_updates(rName, 2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAttachmentExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", "false"),
				),
			},
			{
				Config: testAccVPCAttachmentConfig_updates(rName, 1, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", "true"),
				),
			},
			{
				Config: testAccVPCAttachmentConfig_updates(rName, 2, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", "false"),
				),
			},
			// Cannot currently update ipv6 on its own, must also update subnet_arn
			// {
			// 	Config: testAccVPCAttachmentConfig_updates(rName, 2, true),
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

func testAccCheckVPCAttachmentExists(n string, v *networkmanager.VpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager VPC Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

		output, err := tfnetworkmanager.FindVPCAttachmentByID(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVPCAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_vpc_attachment" {
			continue
		}

		_, err := tfnetworkmanager.FindVPCAttachmentByID(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Network Manager VPC Attachment %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccVPCAttachmentConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "awscc_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  policy_document   = jsonencode(jsondecode(data.aws_networkmanager_core_network_policy_document.test.json))
}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = data.aws_region.current.name
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
`, rName))
}

func testAccVPCAttachmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCAttachmentConfig_base(rName), `
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = aws_subnet.test[*].arn
  core_network_id = awscc_networkmanager_core_network.test.id
  vpc_arn         = aws_vpc.test.arn
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.test.id
  attachment_type = aws_networkmanager_vpc_attachment.test.attachment_type
}
`)
}

func testAccVPCAttachmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVPCAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = [aws_subnet.test[0].arn]
  core_network_id = awscc_networkmanager_core_network.test.id
  vpc_arn         = aws_vpc.test.arn

  tags = {
    %[1]q = %[2]q
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.test.id
  attachment_type = aws_networkmanager_vpc_attachment.test.attachment_type
}
`, tagKey1, tagValue1))
}

func testAccVPCAttachmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVPCAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = [aws_subnet.test[0].arn]
  core_network_id = awscc_networkmanager_core_network.test.id
  vpc_arn         = aws_vpc.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.test.id
  attachment_type = aws_networkmanager_vpc_attachment.test.attachment_type
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCAttachmentConfig_updates(rName string, nSubnets int, ipv6Support bool) string {
	return acctest.ConfigCompose(testAccVPCAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = slice(aws_subnet.test[*].arn, 0, %[2]d)
  core_network_id = awscc_networkmanager_core_network.test.id
  vpc_arn         = aws_vpc.test.arn

  options {
    ipv6_support = %[3]t
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.test.id
  attachment_type = aws_networkmanager_vpc_attachment.test.attachment_type
}
`, rName, nSubnets, ipv6Support))
}
