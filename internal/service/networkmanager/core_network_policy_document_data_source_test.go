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
				Config: testAccCoreNetworkPolicyDocumentBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_networkmanager_core_network_policy_document.test", "json",
						testAccPolicyDocumentExpectedJSON(),
					),
				),
			},
		},
	})
}

func TestAccCoreNetworkPolicyDocumentDataSource_edgeLocations(t *testing.T) {
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, networkmanager.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyDocumentEdgeLocations,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_networkmanager_core_network_policy_document.test", "json",
						testAccCoreNetworkPolicyDocumentExpectedJSON(),
					),
				),
			},
		},
	})
}

var testAccCoreNetworkPolicyDocumentBasic = `
data "aws_networkmanager_core_network_policy_document" "full" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555", "4200000000-4294967294"]
    inside_cidr_blocks = [
      "10.1.0.0/16",
      "192.0.0.0/8",
      "2001:4860::/32"
    ]
    edge_locations {
      location = "us-east-1"
      asn      = 64555
      inside_cidr_blocks = [
        "10.1.0.0/24",
        "192.128.0.0/10",
        "2001:4860:F000::/40"
      ]
    }
    edge_locations {
      location = "us-west-2"
      asn      = 4200000001
      inside_cidr_blocks = [
        "192.192.0.0/10",
        "2001:4860:E000::/40"
      ]
    }

  }

  segments {
    name                          = "GoodSegmentSpecification"
    description                   = "A good segment."
    require_attachment_acceptance = true
    isolate_attachments           = false
    edge_locations = [
      "us-east-1",
      "us-west-2"
    ]
  }

  segments {
    name                          = "AnotherGoodSegmentSpecification"
    description                   = "A good segment."
    require_attachment_acceptance = true
    isolate_attachments           = false
    allow_filter                  = ["AllowThisSegment"]
  }
  segments {
    name                          = "AllowThisSegment"
    require_attachment_acceptance = true
    isolate_attachments           = false
    deny_filter                   = ["DenyThisSegment"]
  }
  segments {
    name                          = "DenyThisSegment"
    require_attachment_acceptance = true
    isolate_attachments           = false
  }
  segments {
    name                          = "a"
    require_attachment_acceptance = true
    isolate_attachments           = false
  }
  segments {
    name                          = "b"
    require_attachment_acceptance = true
    isolate_attachments           = false
  }
  segments {
    name                          = "c"
    require_attachment_acceptance = true
    isolate_attachments           = false
  }

  segment_actions {
    action = "create-route"
    destination_cidr_blocks = [
      "1.1.1.1/32",
      "f:f:f::f/128"
    ]
    destinations = [
      "attachment-11111111111111111",
      "attachment-22222222222222222"
    ]
    segment = "GoodSegmentSpecification"
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "AnotherGoodSegmentSpecification"
    share_with = ["*"]
  }
  segment_actions {
    action  = "share"
    mode    = "attachment-route"
    segment = "AnotherGoodSegmentSpecification"
    share_with_except = [
      "a",
      "b",
      "c"
    ]
  }
  segment_actions {
    action  = "share"
    mode    = "attachment-route"
    segment = "GoodSegmentSpecification"
    share_with_except = [
      "a",
      "b",
      "c"
    ]
  }

  attachment_policies {
    rule_number     = 1
    condition_logic = "and"

    conditions {
      type     = "resource-id"
      operator = "not-equals"
      value    = "one"
    }

    conditions {
      type     = "region"
      operator = "equals"
      value    = "us-west-2"
    }

    conditions {
      type     = "attachment-type"
      operator = "equals"
      value    = "connect"
    }

    conditions {
      type     = "account-id"
      operator = "contains"
      value    = "one"
    }

    conditions {
      type  = "tag-exists"
      value = "tag-a"
    }

    conditions {
      type     = "tag-value"
      operator = "contains"
      key      = "tag-b"
      value    = "one"
    }

    action {
      association_method = "tag"
      tag_value_of_key   = "segment"
    }
  }

  attachment_policies {
    rule_number     = 64
    condition_logic = "or"

    conditions {
      type     = "resource-id"
      operator = "not-equals"
      value    = "one"
    }

    conditions {
      type     = "region"
      operator = "equals"
      value    = "us-west-2"
    }

    conditions {
      type     = "attachment-type"
      operator = "equals"
      value    = "vpc"
    }

    conditions {
      type     = "account-id"
      operator = "contains"
      value    = "one"
    }

    conditions {
      type  = "tag-exists"
      value = "tag-a"
    }

    conditions {
      type     = "tag-value"
      operator = "contains"
      key      = "tag-b"
      value    = "one"
    }
    action {
      association_method = "constant"
      segment            = "GoodSegmentSpecification"
      require_acceptance = true
    }
  }

  attachment_policies {
    rule_number     = 72
    condition_logic = "or"

    conditions {
      type = "any"
    }
    action {
      association_method = "constant"
      segment            = "GoodSegmentSpecification"
      require_acceptance = true
    }
  }

  attachment_policies {
    rule_number     = 71
    condition_logic = "or"

    conditions {
      type     = "region"
      operator = "equals"
      value    = "eu-west-1"
    }
    action {
      association_method = "constant"
      segment            = "GoodSegmentSpecification"
      require_acceptance = true
    }
  }
}`

var testAccCoreNetworkPolicyDocumentEdgeLocations = `
data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location           = "us-east-1"
      asn                = 64512
      inside_cidr_blocks = ["10.1.0.0/24"]
    }
    inside_cidr_blocks = ["10.1.0.0/16"]
  }

  segments {
    name           = "development"
    edge_locations = ["us-east-1"]
  }
  segments {
    name           = "production"
    edge_locations = ["us-east-1"]
  }

  attachment_policies {
    rule_number     = 1

    conditions {
      type     = "tag-value"
      operator = "contains"
      key      = "segment"
      value    = "development"
    }
    action {
      association_method = "constant"
      segment            = "development"
    }
  }
  attachment_policies {
    rule_number = 2

    conditions {
      type     = "tag-value"
      operator = "contains"
      key      = "segment"
      value    = "production"
    }
    action {
      association_method = "constant"
      segment            = "production"
    }
  }
}`

func testAccPolicyDocumentExpectedJSON() string {
	return fmt.Sprint(`{
  "version": "2021.12",
  "core-network-configuration": {
    "asn-ranges": [
      "64512-64555"
    ],
    "vpn-ecmp-support": false,
    "edge-locations": [
      {
        "location": "us-east-1",
        "asn": 64512
      },
      {
        "location": "eu-central-1",
        "asn": 64513
      }
    ]
  },
  "segments": [
    {
      "name": "shared",
      "description": "Segment for shared services",
      "require-attachment-acceptance": true
    },
    {
      "name": "prod",
      "description": "Segment for prod services",
      "require-attachment-acceptance": true
    },
    {
      "name": "finance",
      "description": "Segment for finance services",
      "require-attachment-acceptance": true
    },
    {
      "name": "hr",
      "description": "Segment for hr services",
      "require-attachment-acceptance": true
    },
    {
      "name": "vpn",
      "description": "Segment for vpn services",
      "require-attachment-acceptance": true
    }
  ],
  "attachment-policies": [
    {
      "rule-number": 100,
      "action": {
        "association-method": "constant",
        "segment": "shared"
      },
      "conditions": [
        {
          "type": "tag-value",
          "operator": "equals",
          "key": "segment",
          "value": "shared"
        }
      ],
      "condition-logic": "or"
    },
    {
      "rule-number": 200,
      "action": {
        "association-method": "constant",
        "segment": "prod"
      },
      "conditions": [
        {
          "type": "tag-value",
          "operator": "equals",
          "key": "segment",
          "value": "prod"
        }
      ],
      "condition-logic": "or"
    },
    {
      "rule-number": 300,
      "action": {
        "association-method": "constant",
        "segment": "finance"
      },
      "conditions": [
        {
          "type": "tag-value",
          "operator": "equals",
          "key": "segment",
          "value": "finance"
        }
      ],
      "condition-logic": "or"
    },
    {
      "rule-number": 400,
      "action": {
        "association-method": "constant",
        "segment": "hr"
      },
      "conditions": [
        {
          "type": "tag-value",
          "operator": "equals",
          "key": "segment",
          "value": "hr"
        }
      ],
      "condition-logic": "or"
    },
    {
      "rule-number": 500,
      "action": {
        "association-method": "constant",
        "segment": "vpn"
      },
      "conditions": [
        {
          "type": "tag-value",
          "operator": "equals",
          "key": "segment",
          "value": "vpn"
        }
      ],
      "condition-logic": "or"
    }
  ],
  "segment-actions": [
    {
      "action": "share",
      "mode": "attachment-route",
      "segment": "shared",
      "share-with": "*"
    },
    {
      "action": "share",
      "mode": "attachment-route",
      "segment": "vpn",
      "share-with": "*"
    }
  ]
}`)
}

func testAccCoreNetworkPolicyDocumentExpectedJSON() string {
	return fmt.Sprint(`{
  "version": "2021.12",
  "core-network-configuration": {
    "asn-ranges": [
      "64512-64555"
    ],
    "vpn-ecmp-support": false,
    "edge-locations": [
      {
        "location": "us-east-1",
        "asn": 64512,
        "inside-cidr-blocks": [
          "10.1.0.0/24"
        ]
      }
    ],
    "inside-cidr-blocks": [
      "10.1.0.0/16"
    ]
  },
  "segments": [
    {
      "name": "development",
      "edge-locations": [
        "us-east-1"
      ]
    },
    {
      "name": "production",
      "edge-locations": [
        "us-east-1"
      ]
    }
  ],
  "attachment-policies": [
    {
      "rule-number": 1,
      "action": {
        "association-method": "constant",
        "segment": "development"
      },
      "conditions": [
        {
          "type": "tag-value",
          "operator": "contains",
          "key": "segment",
          "value": "development"
        }
      ]
    },
    {
      "rule-number": 2,
      "action": {
        "association-method": "constant",
        "segment": "production"
      },
      "conditions": [
        {
          "type": "tag-value",
          "operator": "contains",
          "key": "segment",
          "value": "production"
        }
      ]
    }
  ]
}`)
}
