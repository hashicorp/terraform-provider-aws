package networkmanager_test

import (
	"context"
	"fmt"
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
					testAccCheckPolicyDocument(resourceName, originalSegmentValue),
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
					testAccCheckPolicyDocument(resourceName, updatedSegmentValue),
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
data "aws_region" "current" {}

resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = data.aws_region.current.name
    }
  }

  segments {
    name = %[1]q
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  lifecycle {
    ignore_changes = [policy_document]
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, segmentValue)
}
