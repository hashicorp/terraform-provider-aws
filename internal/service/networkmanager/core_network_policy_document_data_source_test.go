// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerCoreNetworkPolicyDocumentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyDocumentDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrEquivalentJSON("data.aws_networkmanager_core_network_policy_document.test", names.AttrJSON, testAccPolicyDocumentBasicExpectedJSON),
				),
			},
		},
	})
}

func TestAccNetworkManagerCoreNetworkPolicyDocumentDataSource_serviceInsertion(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyDocumentDataSourceConfig_serviceInsertion,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrEquivalentJSON("data.aws_networkmanager_core_network_policy_document.test", names.AttrJSON, testAccPolicyDocumentServiceInsertionExpectedJSON),
				),
			},
		},
	})
}

// lintignore:AWSAT003
const testAccCoreNetworkPolicyDocumentDataSourceConfig_basic = `
data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges = [
      "64512-65534",
      "4200000000-4294967294"
    ]
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
      location = "eu-west-1"
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
      "eu-west-1"
    ]
  }

  segments {
    name                          = "AnotherGoodSegmentSpecification"
    description                   = "A good segment."
    require_attachment_acceptance = false
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
    isolate_attachments           = true
  }
  segments {
    name                          = "c"
    isolate_attachments           = false
    require_attachment_acceptance = true
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
    share_with = [
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
      value    = "eu-west-1"
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
      type = "tag-exists"
      key  = "tag-a"
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
      value    = "eu-west-1"
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
      type = "tag-exists"
      key  = "tag-a"
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
      require_acceptance = false
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
}
`

// lintignore:AWSAT003
const testAccPolicyDocumentBasicExpectedJSON = `{
  "version": "2021.12",
  "core-network-configuration": {
    "asn-ranges": [
      "64512-65534",
      "4200000000-4294967294"
    ],
    "vpn-ecmp-support": false,
    "edge-locations": [
      {
        "location": "us-east-1",
        "asn": 64555,
        "inside-cidr-blocks": [
          "2001:4860:F000::/40",
          "192.128.0.0/10",
          "10.1.0.0/24"
        ]
      },
      {
        "location": "eu-west-1",
        "asn": 4200000001,
        "inside-cidr-blocks": [
          "2001:4860:E000::/40",
          "192.192.0.0/10"
        ]
      }
    ],
    "inside-cidr-blocks": [
      "2001:4860::/32",
      "192.0.0.0/8",
      "10.1.0.0/16"
    ]
  },
  "segments": [
    {
      "name": "GoodSegmentSpecification",
      "description": "A good segment.",
      "edge-locations": [
        "us-east-1",
        "eu-west-1"
      ],
      "isolate-attachments": false,
      "require-attachment-acceptance": true
    },
    {
      "name": "AnotherGoodSegmentSpecification",
      "description": "A good segment.",
      "allow-filter": [
        "AllowThisSegment"
      ],
      "isolate-attachments": false,
      "require-attachment-acceptance": false
    },
    {
      "name": "AllowThisSegment",
      "deny-filter": [
        "DenyThisSegment"
      ],
      "isolate-attachments": false,
      "require-attachment-acceptance": true
    },
    {
      "name": "DenyThisSegment",
      "isolate-attachments": false,
      "require-attachment-acceptance": true
    },
    {
      "name": "a",
      "isolate-attachments": false,
      "require-attachment-acceptance": true
    },
    {
      "name": "b",
      "isolate-attachments": true,
      "require-attachment-acceptance": true
    },
    {
      "name": "c",
      "isolate-attachments": false,
      "require-attachment-acceptance": true
    }
  ],
  "attachment-policies": [
    {
      "rule-number": 1,
      "action": {
        "association-method": "tag",
        "tag-value-of-key": "segment"
      },
      "conditions": [
        {
          "type": "resource-id",
          "operator": "not-equals",
          "value": "one"
        },
        {
          "type": "region",
          "operator": "equals",
          "value": "eu-west-1"
        },
        {
          "type": "attachment-type",
          "operator": "equals",
          "value": "connect"
        },
        {
          "type": "account-id",
          "operator": "contains",
          "value": "one"
        },
        {
          "type": "tag-exists",
          "key": "tag-a"
        },
        {
          "type": "tag-value",
          "operator": "contains",
          "key": "tag-b",
          "value": "one"
        }
      ],
      "condition-logic": "and"
    },
    {
      "rule-number": 64,
      "action": {
        "association-method": "constant",
        "segment": "GoodSegmentSpecification",
        "require-acceptance": true
      },
      "conditions": [
        {
          "type": "resource-id",
          "operator": "not-equals",
          "value": "one"
        },
        {
          "type": "region",
          "operator": "equals",
          "value": "eu-west-1"
        },
        {
          "type": "attachment-type",
          "operator": "equals",
          "value": "vpc"
        },
        {
          "type": "account-id",
          "operator": "contains",
          "value": "one"
        },
        {
          "type": "tag-exists",
          "key": "tag-a"
        },
        {
          "type": "tag-value",
          "operator": "contains",
          "key": "tag-b",
          "value": "one"
        }
      ],
      "condition-logic": "or"
    },
    {
      "rule-number": 72,
      "action": {
        "association-method": "constant",
        "segment": "GoodSegmentSpecification"
      },
      "conditions": [
        {
          "type": "any"
        }
      ],
      "condition-logic": "or"
    },
    {
      "rule-number": 71,
      "action": {
        "association-method": "constant",
        "segment": "GoodSegmentSpecification",
        "require-acceptance": true
      },
      "conditions": [
        {
          "type": "region",
          "operator": "equals",
          "value": "eu-west-1"
        }
      ],
      "condition-logic": "or"
    }
  ],
  "segment-actions": [
    {
      "action": "create-route",
      "destinations": [
        "attachment-22222222222222222",
        "attachment-11111111111111111"
      ],
      "destination-cidr-blocks": [
        "f:f:f::f/128",
        "1.1.1.1/32"
      ],
      "segment": "GoodSegmentSpecification"
    },
    {
      "action": "share",
      "mode": "attachment-route",
      "segment": "AnotherGoodSegmentSpecification",
      "share-with": "*"
    },
    {
      "action": "share",
      "mode": "attachment-route",
      "segment": "AnotherGoodSegmentSpecification",
      "share-with": {
        "except": [
          "c",
          "b",
          "a"
        ]
      }
    },
    {
      "action": "share",
      "mode": "attachment-route",
      "segment": "GoodSegmentSpecification",
      "share-with": [
        "c",
        "b",
        "a"
      ]
    }
  ],
  "network-function-groups": []
}`

// lintignore:AWSAT003
const testAccCoreNetworkPolicyDocumentDataSourceConfig_serviceInsertion = `
data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = true
    asn_ranges = [
      "64512-65534"
    ]
    inside_cidr_blocks = [
      "10.0.0.0/16"
    ]
    edge_locations {
      location = "us-east-2"
    }
    edge_locations {
      location = "us-west-2"
    }
  }

  segments {
    name                          = "development"
    require_attachment_acceptance = true
    isolate_attachments           = true
    edge_locations = [
      "us-east-2"
    ]
  }

  segments {
    name                          = "production"
    require_attachment_acceptance = true
    isolate_attachments           = true
    edge_locations = [
      "us-east-2"
    ]
  }

  segment_actions {
    action  = "send-via"
    segment = "development"
    mode    = "single-hop"

    when_sent_to {
      segments = [
        "production",
      ]
    }

    via {
      network_function_groups = ["InspectionVPC"]
    }
  }

  attachment_policies {
    rule_number     = 125
    condition_logic = "and"

    conditions {
      type = "tag-exists"
      key  = "InspectionVpcs"
    }

    action {
      add_to_network_function_group = "InspectionVPC"
    }
  }

  network_function_groups {
    name                          = "InspectionVPC"
    description                   = "Route segment traffic to the inspection VPC"
    require_attachment_acceptance = true
  }
}
`

// lintignore:AWSAT003
const testAccPolicyDocumentServiceInsertionExpectedJSON = `{
  "version": "2021.12",
  "core-network-configuration": {
    "vpn-ecmp-support": true,
    "inside-cidr-blocks": [
      "10.0.0.0/16"
    ],
    "asn-ranges": [
      "64512-65534"
    ],
    "edge-locations": [
      {
        "location": "us-east-2"
      },
      {
        "location": "us-west-2"
      }
    ]
  },
  "segments": [
    {
      "name": "development",
      "edge-locations": [
        "us-east-2"
      ],
      "require-attachment-acceptance": true,
      "isolate-attachments": true
    },
    {
      "name": "production",
      "edge-locations": [
        "us-east-2"
      ],
      "require-attachment-acceptance": true,
      "isolate-attachments": true
    }
  ],
  "network-function-groups": [
    {
      "name": "InspectionVPC",
      "description": "Route segment traffic to the inspection VPC",
      "require-attachment-acceptance": true
    }
  ],
  "segment-actions": [
    {
      "action": "send-via",
      "segment": "development",
      "mode": "single-hop",
      "when-sent-to": {
        "segments": [
          "production"
        ]
      },
      "via": {
        "network-function-groups": [
          "InspectionVPC"
        ]
      }
    }
  ],
  "attachment-policies": [
    {
      "rule-number": 125,
      "condition-logic": "and",
      "conditions": [
        {
          "type": "tag-exists",
          "key": "InspectionVpcs"
        }
      ],
      "action": {
        "add-to-network-function-group": "InspectionVPC"
      }
    }
  ]
}`
