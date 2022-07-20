---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_core_network_policy_document"
description: |-
  Generates an Core Network policy document in JSON format
---

# Data Source: aws_networkmanager_core_network_policy_document

Generates a Core Network policy document in JSON format for use with resources that expect core network policy documents such as `awscc_networkmanager_core_network`. It follows the API definition from the [core-network-policy documentation](https://docs.aws.amazon.com/vpc/latest/cloudwan/cloudwan-policies-json.html).

Using this data source to generate policy documents is *optional*. It is also valid to use literal JSON strings in your configuration or to use the `file` interpolation function to read a raw JSON policy document from a file.

-> For more information about building AWS Core Network policy documents with Terraform, see the [Using AWS & AWSCC Provider Together Guide](/docs/providers/aws/guides/using-aws-with-awscc-provider.html)

## Example Usage

### Basic Example

```terraform
data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = "us-east-1"
      asn      = 64512
    }
    edge_locations {
      location = "eu-central-1"
      asn      = 64513
    }
  }

  segments {
    name                          = "shared"
    description                   = "Segment for shared services"
    require_attachment_acceptance = true
  }
  segments {
    name                          = "prod"
    description                   = "Segment for prod services"
    require_attachment_acceptance = true
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
    share_with = ["*"]
  }

  attachment_policies {
    rule_number     = 100
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
  attachment_policies {
    rule_number     = 200
    condition_logic = "or"

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "segment"
      value    = "prod"
    }
    action {
      association_method = "constant"
      segment            = "prod"
    }
  }
}
```

`data.aws_networkmanager_core_network_policy_document.test.json` will evaluate to:

```json
{
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
    }
  ],
  "segment-actions": [
    {
      "action": "share",
      "mode": "attachment-route",
      "segment": "shared",
      "share-with": "*"
    }
  ]
}
```

## Argument Reference

The following arguments are available:

* `attachment_policies` (Optional) - In a core network, all attachments use the block argument `attachment_policies` section to map an attachment to a segment. Instead of manually associating a segment to each attachment, attachments use tags, and then the tags are used to associate the attachment to the specified segment. Detailed below.
* `core_network_configuration` (Required) - The core network configuration section defines the Regions where a core network should operate. For AWS Regions that are defined in the policy, the core network creates a Core Network Edge where you can connect attachments. After it's created, each Core Network Edge is peered with every other defined Region and is configured with consistent segment and routing across all Regions. Regions cannot be removed until the associated attachments are deleted. Detailed below.
* `segments` (Required) - A block argument that defines the different segments in the network. Here you can provide descriptions, change defaults, and provide explicit Regional operational and route filters. The names defined for each segment are used in the `segment_actions` and `attachment_policies` section. Each segment is created, and operates, as a completely separated routing domain. By default, attachments can only communicate with other attachments in the same segment. Detailed below.
* `segment_actions` (Optional) - A block argument, `segment_actions` define how routing works between segments. By default, attachments can only communicate with other attachments in the same segment. Detailed below.

### `attachment_policies`

The following arguments are available:

* `action` (Required) - The action to take when a condition is true. Detailed Below.
* `condition_logic` (Optional) - Valid values include `and` or `or`. This is a mandatory parameter only if you have more than one condition. The `condition_logic` apply to all of the conditions for a rule, which also means nested conditions of `and` or `or` are not supported. Use `or` if you want to associate the attachment with the segment by either the segment name or attachment tag value, or by the chosen conditions. Use `and` if you want to associate the attachment with the segment by either the segment name or attachment tag value and by the chosen conditions. Detailed Below.
* `conditions` (Required) - A block argument. Detailed Below.
* `description` (Optional) - A user-defined description that further helps identify the rule.
* `rule_number` (Required) - An integer from `1` to `65535` indicating the rule's order number. Rules are processed in order from the lowest numbered rule to the highest. Rules stop processing when a rule is matched. It's important to make sure that you number your rules in the exact order that you want them processed.

### `action`

The following arguments are available:

* `association_method` (Required) - Defines how a segment is mapped. Values can be `constant` or `tag`. `constant` statically defines the segment to associate the attachment to. `tag` uses the value of a tag to dynamically try to map to a segment.reference_policies_elements_condition_operators.html) to evaluate.
* `segment` (Optional) - The name of the `segment` to share as defined in the `segments` section. This is used only when the `association_method` is `constant`.
* `tag_value_of_key` (Optional) - Maps the attachment to the value of a known key. This is used with the `association_method` is `tag`. For example a `tag` of `stage = “test”`, will map to a segment named `test`. The value must exactly match the name of a segment. This allows you to have many segments, but use only a single rule without having to define multiple nearly identical conditions. This prevents creating many similar conditions that all use the same keys to map to segments.
* `require_acceptance` (Optional) - Determines if this mapping should override the segment value for `require_attachment_acceptance`. You can only set this to `true`, indicating that this setting applies only to segments that have `require_attachment_acceptance` set to `false`. If the segment already has the default `require_attachment_acceptance`, you can set this to inherit segment’s acceptance value.

### `conditions`

The conditions block has 4 arguments `type`, `operator`, `key`, `value`. Setting or omitting each argument requires a combination of logic based on the value set to `type`. For that reason, please refer to the [AWS documentation](https://docs.aws.amazon.com/vpc/latest/cloudwan/cloudwan-policies-json.html) for complete usage docs.

The following arguments are available:

* `type` (Required) - Valid values include: `account-id`, `any`, `tag-value`, `tag-exists`, `resource-id`, `region`, `attachment-type`.
* `operator` (Optional) - Valid values include: `equals`, `not-equals`, `contains`, `begins-with`.
* `key` (Optional) - string value
* `value` (Optional) - string value

### `core_network_configuration`

The following arguments are available:

* `asn_ranges` (Required) - List of strings containing Autonomous System Numbers (ASNs) to assign to Core Network Edges. By default, the core network automatically assigns an ASN for each Core Network Edge but you can optionally define the ASN in the edge-locations for each Region. The ASN uses an array of integer ranges only from `64512` to `65534` and `4200000000` to `4294967294` expressed as a string like `"64512-65534"`. No other ASN ranges can be used.
* `inside_cidr_blocks` (Optional) - The Classless Inter-Domain Routing (CIDR) block range used to create tunnels for AWS Transit Gateway Connect. The format is standard AWS CIDR range (for example, `10.0.1.0/24`). You can optionally define the inside CIDR in the Core Network Edges section per Region. The minimum is a `/24` for IPv4 or `/64` for IPv6. You can provide multiple `/24` subnets or a larger CIDR range. If you define a larger CIDR range, new Core Network Edges will be automatically assigned `/24` and `/64` subnets from the larger CIDR. an Inside CIDR block is required for attaching Connect attachments to a Core Network Edge.
* `vpn_ecmp_support` (Optional) - Indicates whether the core network forwards traffic over multiple equal-cost routes using VPN. The value can be either `true` or `false`. The default is `true`.
* `edge_locations` (Required) - A block value of AWS Region locations where you're creating Core Network Edges. Detailed below.

### `edge_locations`

The following arguments are available:

* `locations` (Required) - An AWS Region code, such as `us-east-1`.
* `asn` (Optional) - The ASN of the Core Network Edge in an AWS Region. By default, the ASN will be a single integer automatically assigned from `asn_ranges`
* `inside_cidr_blocks` (Optional) - The local CIDR blocks for this Core Network Edge for AWS Transit Gateway Connect attachments. By default, this CIDR block will be one or more optional IPv4 and IPv6 CIDR prefixes auto-assigned from `inside_cidr_blocks`.

### `segments`

The following arguments are available:

* `allow_filter` (Optional) -  List of strings of segment names that explicitly allows only routes from the segments that are listed in the array. Use the `allow_filter` setting if a segment has a well-defined group of other segments that connectivity should be restricted to. It is applied after routes have been shared in `segment_actions`. If a segment is listed in `allow_filter`, attachments between the two segments will have routes if they are also shared in the segment-actions area. For example, you might have a segment named "video-producer" that should only ever share routes with a "video-distributor" segment, no matter how many other share statements are created.
* `deny_filter` (Optional) - An array of segments that disallows routes from the segments listed in the array. It is applied only after routes have been shared in `segment_actions`. If a segment is listed in the `deny_filter`, attachments between the two segments will never have routes shared across them. For example, you might have a "financial" payment segment that should never share routes with a "development" segment, regardless of how many other share statements are created. Adding the payments segment to the deny-filter parameter prevents any shared routes from being created with other segments.
* `description` (Optional) - A user-defined string describing the segment.
* `edge_locations` (Optional) - A list of strings of AWS Region names. Allows you to define a more restrictive set of Regions for a segment. The edge location must be a subset of the locations that are defined for `edge_locations` in the `core_network_configuration`.
* `isolate_attachments` (Optional) - This Boolean setting determines whether attachments on the same segment can communicate with each other. If set to `true`, the only routes available will be either shared routes through the share actions, which are attachments in other segments, or static routes. The default value is `false`. For example, you might have a segment dedicated to "development" that should never allow VPCs to talk to each other, even if they’re on the same segment. In this example, you would keep the default parameter of `false`.
* `name` (Required) - A unique name for a segment. The name is a string used in other parts of the policy document, as well as in the console for metrics and other reference points. Valid characters are a–z, and 0–9.
* `require_attachment_acceptance` (Optional) - This Boolean setting determines whether attachment requests are automatically approved or require acceptance. The default is `true`, indicating that attachment requests require acceptance. For example, you might use this setting to allow a "sandbox" segment to allow any attachment request so that a core network or attachment administrator does not need to review and approve attachment requests. In this example, `require_attachment_acceptance` is set to `false`.

### `segment_actions`

`segment_actions` have differnet outcomes based on their `action` argument value. There are 2 valid values for `action`: `create-route` & `share`. Behaviors of the below arguments changed depending on the `action` you specify. For more details on their use see the [AWS documentation](https://docs.aws.amazon.com/vpc/latest/cloudwan/cloudwan-policies-json.html#cloudwan-segment-actions-json).

~> **NOTE**: `share_with` and `share_with_except` break from the AWS API specification. The API has 1 argument `share-with` and it can accept 3 input types as valid (`"*"`, `["<segment-name>"]`, or `{ except: ["<segment-name>"]}`). To emulate this behavior, `share_with` is always a list that can accept the argument `["*"]` as valid for `"*"` and `share_with_except` is a that can accept `["<segment-name>"]` as valid for `{ except: ["<segment-name>"]}`. You may only specify one of: `share_with` or `share_with_except`.


The following arguments are available:

* `action` (Required) - The action to take for the chosen segment. Valid values `create-route` or `share`.
* `description` (Optional) - A user-defined string describing the segment action.
* `destination_cidr_blocks` (Optional) - List of strings containing CIDRs. You can define the IPv4 and IPv6 CIDR notation for each AWS Region. For example, `10.1.0.0/16` or `2001:db8::/56`. This is an array of CIDR notation strings.
* `destinations` (Optional) - A list of strings. Valid values include `["blackhole"]` or a list of attachment ids.
* `mode` (Optional) - A string. This mode places the attachment and return routes in each of the `share_with` segments. Valid values include: `attachment-route`.
* `segment` (Optional) - The name of the segment.
* `share_with` (Optional) - A list of strings to share with. Must be a substring is all segments. Valid values include: `["*"]` or `["<segment-names>"]`.
* `share_with_except` (Optional) - A set subtraction of segments to not share with.

## Attributes Reference

The following attribute is exported:

* `json` - Standard JSON policy document rendered based on the arguments above.
