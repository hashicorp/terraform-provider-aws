package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCoreNetworkPolicyDocumentDataSource_basic(t *testing.T) {
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, networkmanager.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_networkmanager_core_network_policy_document.test", "json",
						testAccPolicyDocumentExpectedJSON(),
					),
				),
			},
		},
	})
}

var testAccPolicyDocumentConfig = `
data "aws_networkmanager_core_network_policy_document" "test" {
  segments {
    name = "test"
  }
  segments {
    name = "test2"
    require_attachment_acceptance = true
  }
}
`

func testAccPolicyDocumentExpectedJSON() string {
	return fmt.Sprint(`{
  "Version": "2021.12",
  "Segments": [
    {
      "Name": "test"
    },
    {
      "Name": "test2",
      "RequireAttachmentAcceptance": true
    }
  ],
  "CoreNetworkConfiguration": null
}`)
}
