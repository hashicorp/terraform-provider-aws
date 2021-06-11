---
subcategory: "Gamelift"
layout: "aws"
page_title: "AWS: aws_gamelift_matchmaking_rule_set"
description: |-
  Provides a Gamelift Matchmaking Rule Set resource.
---

# Resource: aws_gamelift_matchmaking_rule_set

Provides a Gamelift Matchmaking Rule Set resource.

## Example Usage

```hcl
resource "aws_gamelift_matchmaking_rule_set" "test" {
  name = "example-rule-set"

  rule_set_body = <<RULE_SET_BODY
{
  "name": "test",
  "ruleLanguageVersion": "1.0",
  "teams": [{
    "name": "alpha",
    "minPlayers": 1,
    "maxPlayers": 5
  }]
}
RULE_SET_BODY
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the matchmaking rule set.
* `rule_set_body` - (Required) A collection of matchmaking rules, formatted as a JSON string.
* `tags` - (Optional) Key-value mapping of resource tags.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Matchmaking Rule Set ARN.

## Import

Gamelift Matchmaking Rule Sets can be imported by their `name`, e.g.

```
$ terraform import aws_gamelift_matchmaking_rule_set.example example
```
