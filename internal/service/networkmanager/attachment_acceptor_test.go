package networkmanager_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccNetworkManagerAttachmentAccepter_vpcAttachmentBasic(t *testing.T) {
	resourceName := "aws_networkmanager_attachment_accepter.test"
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
				Config: testAccCoreNetworkConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "VPC"),
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
