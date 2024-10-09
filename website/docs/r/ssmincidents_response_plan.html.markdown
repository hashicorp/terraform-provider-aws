---
subcategory: "SSM Incident Manager Incidents"
layout: "aws"
page_title: "AWS: aws_ssmincidents_response_plan"
description: |-
  Terraform resource for managing an incident response plan in AWS Systems Manager Incident Manager.
---

# Resource: aws_ssmincidents_response_plan

Provides a Terraform resource to manage response plans in AWS Systems Manager Incident Manager.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssmincidents_response_plan" "example" {
  name = "name"

  incident_template {
    title  = "title"
    impact = "3"
  }

  tags = {
    key = "value"
  }

  depends_on = [aws_ssmincidents_replication_set.example]
}

```

### Usage With All Fields

```terraform
resource "aws_ssmincidents_response_plan" "example" {
  name = "name"

  incident_template {
    title         = "title"
    impact        = "3"
    dedupe_string = "dedupe"
    incident_tags = {
      key = "value"
    }

    notification_target {
      sns_topic_arn = aws_sns_topic.example1.arn
    }

    notification_target {
      sns_topic_arn = aws_sns_topic.example2.arn
    }

    summary = "summary"
  }

  display_name = "display name"
  chat_channel = [aws_sns_topic.topic.arn]
  engagements  = ["arn:aws:ssm-contacts:us-east-2:111122223333:contact/test1"]

  action {
    ssm_automation {
      document_name    = aws_ssm_document.document1.name
      role_arn         = aws_iam_role.role1.arn
      document_version = "version1"
      target_account   = "RESPONSE_PLAN_OWNER_ACCOUNT"
      parameter {
        name   = "key"
        values = ["value1", "value2"]
      }
      parameter {
        name   = "foo"
        values = ["bar"]
      }
      dynamic_parameters = {
        someKey    = "INVOLVED_RESOURCES"
        anotherKey = "INCIDENT_RECORD_ARN"
      }
    }
  }

  integration {
    pagerduty {
      name       = "pagerdutyIntergration"
      service_id = "example"
      secret_id  = "example"
    }
  }

  tags = {
    key = "value"
  }

  depends_on = [aws_ssmincidents_replication_set.example]
}

```

## Argument Reference

~> NOTE: A response plan implicitly depends on a replication set. If you configured your replication set in Terraform,
we recommend you add it to the `depends_on` argument for the Terraform ResponsePlan Resource.

The following arguments are required:

* `name` - (Required) The name of the response plan.

The `incident_template` configuration block is required and supports the following arguments:

* `title` - (Required) The title of a generated incident.
* `impact` - (Required) The impact value of a generated incident. The following values are supported:
    * `1` - Severe Impact
    * `2` - High Impact
    * `3` - Medium Impact
    * `4` - Low Impact
    * `5` - No Impact
* `dedupe_string` - (Optional) A string used to stop Incident Manager from creating multiple incident records for the same incident.
* `incident_tags` - (Optional) The tags assigned to an incident template. When an incident starts, Incident Manager assigns the tags specified in the template to the incident.
* `summary` - (Optional) The summary of an incident.
* `notification_target` - (Optional) The Amazon Simple Notification Service (Amazon SNS) targets that this incident notifies when it is updated. The `notification_target` configuration block supports the following argument:
    * `sns_topic_arn` - (Required) The ARN of the Amazon SNS topic.

The following arguments are optional:

* `tags` - (Optional) The tags applied to the response plan.
* `display_name` - (Optional) The long format of the response plan name. This field can contain spaces.
* `chat_channel` - (Optional) The Chatbot chat channel used for collaboration during an incident.
* `engagements` - (Optional) The Amazon Resource Name (ARN) for the contacts and escalation plans that the response plan engages during an incident.
* `action` - (Optional) The actions that the response plan starts at the beginning of an incident.
    * `ssm_automation` - (Optional) The Systems Manager automation document to start as the runbook at the beginning of the incident. The following values are supported:
        * `document_name` - (Required) The automation document's name.
        * `role_arn` - (Required) The Amazon Resource Name (ARN) of the role that the automation document assumes when it runs commands.
        * `document_version` - (Optional) The version of the automation document to use at runtime.
        * `target_account` -  (Optional) The account that the automation document runs in. This can be in either the management account or an application account.
        * `parameter` - (Optional) The key-value pair parameters to use when the automation document runs. The following values are supported:
            * `name` - The name of parameter.
            * `values` - The values for the associated parameter name.
        * `dynamic_parameters` - (Optional) The key-value pair to resolve dynamic parameter values when processing a Systems Manager Automation runbook.
* `integration` - (Optional) Information about third-party services integrated into the response plan. The following values are supported:
    * `pagerduty` - (Optional) Details about the PagerDuty configuration for a response plan. The following values are supported:
        * `name` - (Required) The name of the PagerDuty configuration.
        * `service_id` - (Required) The ID of the PagerDuty service that the response plan associated with the incident at launch.
        * `secret_id` - (Required) The ID of the AWS Secrets Manager secret that stores your PagerDuty key &mdash; either a General Access REST API Key or User Token REST API Key &mdash; and other user credentials.

For more information about the constraints for each field, see [CreateResponsePlan](https://docs.aws.amazon.com/incident-manager/latest/APIReference/API_CreateResponsePlan.html) in the *AWS Systems Manager Incident Manager API Reference*.
  
## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the response plan.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Incident Manager response plan using the response plan ARN. You can find the response plan ARN in the AWS Management Console. For example:

```terraform
import {
  to = aws_ssmincidents_response_plan.responsePlanName
  id = "ARNValue"
}
```

Using `terraform import`, import an Incident Manager response plan using the response plan ARN. You can find the response plan ARN in the AWS Management Console. For example:

```console
% terraform import aws_ssmincidents_response_plan.responsePlanName ARNValue
```
