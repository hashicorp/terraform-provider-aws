// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerCoreNetworkPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	originalSegmentValue := "segmentValue1"
	updatedSegmentValue := "segmentValue2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCoreNetworkPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_basic(originalSegmentValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"65022-65534\"],\"edge-locations\":[{\"location\":\"%s\"}],\"vpn-ecmp-support\":true},\"segments\":[{\"isolate-attachments\":false,\"name\":\"%s\",\"require-attachment-acceptance\":true}],\"version\":\"2021.12\"}", acctest.Region(), originalSegmentValue)),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, networkmanager.CoreNetworkStateAvailable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_basic(updatedSegmentValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"65022-65534\"],\"edge-locations\":[{\"location\":\"%s\"}],\"vpn-ecmp-support\":true},\"segments\":[{\"isolate-attachments\":false,\"name\":\"%s\",\"require-attachment-acceptance\":true}],\"version\":\"2021.12\"}", acctest.Region(), updatedSegmentValue)),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, networkmanager.CoreNetworkStateAvailable),
				),
			},
		},
	})
}

func TestAccNetworkManagerCoreNetworkPolicyAttachment_vpcAttachment(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	segmentValue := "segmentValue"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCoreNetworkPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(fmt.Sprintf(`{"core-network-configuration":{"asn-ranges":\["65022-65534"\],"edge-locations":\[{"location":"%s"}\],"vpn-ecmp-support":true},"segment-actions":\[{"action":"create-route","destination-cidr-blocks":\["0.0.0.0/0"\],"destinations":\["attachment-.+"\],"segment":"segment"}\],"segments":\[{"isolate-attachments":false,"name":"segment","require-attachment-acceptance":true}\],"version":"2021.12"}`, acctest.Region()))),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, networkmanager.CoreNetworkStateAvailable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_basic(segmentValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"65022-65534\"],\"edge-locations\":[{\"location\":\"%s\"}],\"vpn-ecmp-support\":true},\"segments\":[{\"isolate-attachments\":false,\"name\":\"%s\",\"require-attachment-acceptance\":true}],\"version\":\"2021.12\"}", acctest.Region(), segmentValue)),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, networkmanager.CoreNetworkStateAvailable),
				),
			},
		},
	})
}

func TestAccNetworkManagerCoreNetworkPolicyAttachment_vpcAttachmentMultiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var providers []*schema.Provider
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckCoreNetworkPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentMultiRegionCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(fmt.Sprintf(`{"core-network-configuration":{"asn-ranges":\["65022-65534"\],"edge-locations":\[{"location":"%s"},{"location":"%s"}\],"vpn-ecmp-support":true},"segment-actions":\[{"action":"create-route","destination-cidr-blocks":\["10.0.0.0/16"\],"destinations":\["attachment-.+"\],"segment":"segment"},{"action":"create-route","destination-cidr-blocks":\["10.1.0.0/16"\],"destinations":\["attachment-.+"\],"segment":"segment2"}\],"segments":\[{"isolate-attachments":false,"name":"segment","require-attachment-acceptance":true},{"isolate-attachments":false,"name":"segment2","require-attachment-acceptance":true}\],"version":"2021.12"}`, acctest.Region(), acctest.AlternateRegion()))),
					// use test below if the order of locations is unordered
					// resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"65022-65534\"],\"edge-locations\":[{\"location\":\"%s\"},{\"location\":\"%s\"}],\"vpn-ecmp-support\":true},\"segments\":[{\"description\":\"base-policy\",\"isolate-attachments\":false,\"name\":\"segment\",\"require-attachment-acceptance\":false}],\"version\":\"2021.12\"}", acctest.AlternateRegion(), acctest.Region())),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, networkmanager.CoreNetworkStateAvailable),
				),
			},
			{
				Config:            testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentMultiRegionCreate(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerCoreNetworkPolicyAttachment_expectPolicyErrorInvalidASNRange(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCoreNetworkPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccCoreNetworkPolicyAttachmentConfig_expectPolicyErrorInvalidASNRange(),
				ExpectError: regexache.MustCompile("INVALID_ASN_RANGE"),
			},
		},
	})
}

func testAccCheckCoreNetworkPolicyAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	// policy document will not be reverted to empty if the attachment is deleted
	return nil
}

func testAccCheckCoreNetworkPolicyAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Core Network ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn(ctx)

		// pass in latestPolicyVersionId to get the latest version id by default
		const latestPolicyVersionId = -1
		_, err := tfnetworkmanager.FindCoreNetworkPolicyByTwoPartKey(ctx, conn, rs.Primary.ID, latestPolicyVersionId)

		return err
	}
}

func testAccCoreNetworkPolicyAttachmentConfig_basic(segmentValue string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = %[2]q
    }
  }

  segments {
    name = %[1]q
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, segmentValue, acctest.Region())
}

func testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentCreate() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  segment_actions {
    action  = "create-route"
    segment = "segment"
    destination_cidr_blocks = [
      "0.0.0.0/0"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.test.id,
    ]
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id  = aws_networkmanager_global_network.test.id
  create_base_policy = true
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

resource "aws_networkmanager_vpc_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  subnet_arns     = aws_subnet.test[*].arn
  vpc_arn         = aws_vpc.test.arn
}
`, acctest.Region()))
}

func testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentMultiRegionCreate() string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = %[1]q
    }

    edge_locations {
      location = %[2]q
    }
  }

  segments {
    name = "segment"
  }

  segments {
    name = "segment2"
  }

  segment_actions {
    action  = "create-route"
    segment = "segment"
    destination_cidr_blocks = [
      "10.0.0.0/16"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.test.id,
    ]
  }

  segment_actions {
    action  = "create-route"
    segment = "segment2"
    destination_cidr_blocks = [
      "10.1.0.0/16"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.alternate_region.id,
    ]
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  base_policy_regions = [%[1]q, %[2]q]
  create_base_policy  = true
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

resource "aws_networkmanager_vpc_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  subnet_arns     = aws_subnet.test[*].arn
  vpc_arn         = aws_vpc.test.arn
}

# Alternate region
data "aws_availability_zones" "alternate_region_available" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alternate_region" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_subnet" "alternate_region" {
  provider = "awsalternate"

  count = 2

  availability_zone = data.aws_availability_zones.alternate_region_available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.alternate_region.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.alternate_region.id

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_networkmanager_vpc_attachment" "alternate_region" {
  provider = "awsalternate"

  core_network_id = aws_networkmanager_core_network.test.id
  subnet_arns     = aws_subnet.alternate_region[*].arn
  vpc_arn         = aws_vpc.alternate_region.arn
}
`, acctest.Region(), acctest.AlternateRegion()))
}

func testAccCoreNetworkPolicyAttachmentConfig_expectPolicyErrorInvalidASNRange() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534123"] # not a valid range

    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "test"
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}
