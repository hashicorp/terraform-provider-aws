package networkmanager_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
)

func TestAccNetworkManagerCoreNetworkPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	originalSegmentValue := "segmentValue1"
	updatedSegmentValue := "segmentValue2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCoreNetworkPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_basic(originalSegmentValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"65022-65534\"],\"edge-locations\":[{\"location\":\"%s\"}],\"vpn-ecmp-support\":true},\"segments\":[{\"isolate-attachments\":false,\"name\":\"%s\",\"require-attachment-acceptance\":true}],\"version\":\"2021.12\"}", acctest.Region(), originalSegmentValue)),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", "aws_networkmanager_core_network.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "state", networkmanager.CoreNetworkStateAvailable),
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
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", "aws_networkmanager_core_network.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "state", networkmanager.CoreNetworkStateAvailable),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCoreNetworkPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile(fmt.Sprintf(`{"core-network-configuration":{"asn-ranges":\["65022-65534"\],"edge-locations":\[{"location":"%s"}\],"vpn-ecmp-support":true},"segment-actions":\[{"action":"create-route","destination-cidr-blocks":\["0.0.0.0/0"\],"destinations":\["attachment-.+"\],"segment":"segment"}\],"segments":\[{"isolate-attachments":false,"name":"segment","require-attachment-acceptance":true}\],"version":"2021.12"}`, acctest.Region()))),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", "aws_networkmanager_core_network.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "state", networkmanager.CoreNetworkStateAvailable),
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
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", "aws_networkmanager_core_network.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "state", networkmanager.CoreNetworkStateAvailable),
				),
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn()

		_, err := tfnetworkmanager.FindCoreNetworkPolicyByID(ctx, conn, rs.Primary.ID)

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
