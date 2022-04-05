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
data "aws_networkmanager_core_network_policy_document" test {

  core_network_configuration {
      vpn_ecmp_support = true
      asn_ranges = ["64512-64555"]
      edge_locations {
          location = "us-east-1"
          asn = 64512
      }
      edge_locations {
          location = "eu-central-1"
          asn = 64513
      }
  }

  segments {
    name = "test"
  }
  segments {
    name = "test2"
    require_attachment_acceptance = true
  }

  segment_actions {
    action = "share"
    mode =  "attachment-route"
    segment = "shared"
    # share-with = "*"
  }

  attachment_policies {
    rule_number = 100
    condition_logic = "or"

    conditions {
      type = "tag-value"
      operator = "equals"
      key = "segment"
      value = "prod"
    }
    conditions {
      type = "any"
    }
    conditions {
      type = "attachment-type"
      operator = "equals"
      value = "prod"
    }
    action {
      association_method = "constant"
      segment = "prod"
    }
  }
}`

func testAccPolicyDocumentExpectedJSON() string {
	return fmt.Sprint(`{
  "Version": "2021.12",
  "AttachmentPolicies": [
    {
      "RuleNumber": 100,
      "Action": {
        "AssociationMethod": "constant",
        "Segment": "prod"
      },
      "Conditions": [
        {
          "Type": "tag-value",
          "Operator": "equals",
          "Key": "segment",
          "Value": "prod"
        },
        {
          "Type": "any"
        },
        {
          "Type": "attachment-type",
          "Operator": "equals",
          "Value": "prod"
        }
      ],
      "ConditionLogic": "or"
    }
  ],
  "Segments": [
    {
      "Name": "test"
    },
    {
      "Name": "test2",
      "RequireAttachmentAcceptance": true
    }
  ],
  "CoreNetworkConfiguration": {
    "AsnRanges": "64512-64555",
    "VpnEcmpSupport": true,
    "EdgeLocations": [
      {
        "Location": "us-east-1",
        "Asn": 64512
      },
      {
        "Location": "eu-central-1",
        "Asn": 64513
      }
    ]
  }
}`)
}
