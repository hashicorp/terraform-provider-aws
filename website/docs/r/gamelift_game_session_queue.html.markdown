---
subcategory: "GameLift"
layout: "aws"
page_title: "AWS: aws_gamelift_game_session_queue"
description: |-
  Provides a GameLift Game Session Queue resource.
---

# Resource: aws_gamelift_game_session_queue

Provides an GameLift Game Session Queue resource.

## Example Usage

```terraform
resource "aws_gamelift_game_session_queue" "test" {
  name = "example-session-queue"

  destinations = [
    aws_gamelift_fleet.us_west_2_fleet.arn,
    aws_gamelift_fleet.eu_central_1_fleet.arn,
  ]

  notification_target = aws_sns_topic.game_session_queue_notifications.arn

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 100
    policy_duration_seconds                        = 5
  }

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 200
  }

  timeout_in_seconds = 60
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name of the session queue.
* `timeout_in_seconds` - (Required) Maximum time a game session request can remain in the queue.
* `custom_event_data` - (Optional) Information to be added to all events that are related to this game session queue.
* `destinations` - (Optional) List of fleet/alias ARNs used by session queue for placing game sessions.
* `notification_target` - (Optional) An SNS topic ARN that is set up to receive game session placement notifications.
* `player_latency_policy` - (Optional) One or more policies used to choose fleet based on player latency. See below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Nested Fields

#### `player_latency_policy`

* `maximum_individual_player_latency_milliseconds` - (Required) Maximum latency value that is allowed for any player.
* `policy_duration_seconds` - (Optional) Length of time that the policy is enforced while placing a new game session. Absence of value for this attribute means that the policy is enforced until the queue times out.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Game Session Queue ARN.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import GameLift Game Session Queues using their `name`. For example:

```terraform
import {
  to = aws_gamelift_game_session_queue.example
  id = "example"
}
```

Using `terraform import`, import GameLift Game Session Queues using their `name`. For example:

```console
% terraform import aws_gamelift_game_session_queue.example example
```
