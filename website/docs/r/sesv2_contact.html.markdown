---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_contact"
description: |-
  Terraform resource for managing an AWS SESv2 (Simple Email V2) Contact.
---

# Resource: aws_sesv2_contact

Terraform resource for managing an AWS SESv2 (Simple Email V2) Contact.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_contact" "example" {
  contact_list_name = "example"
  email_address     = "example"
}
```

### Extended Usage

```terraform
resource "aws_sesv2_contact" "example" {
  contact_list_name = "example"
  email_address     = "example@example.com"
  unsubscribe_all   = false

  topic_preferences {
    topic_name          = "example-topic-1"
    subscription_status = "OPT_IN"
  }

  topic_preferences {
    topic_name          = "example-topic-2"
    subscription_status = "OPT_IN"
  }
}
```

## Argument Reference

The following arguments are required:

* `contact_list_name` - (Required) The name of the contact list to which the contact should be added.
* `email_address` - (Required) The contact’s email address.

The following arguments are optional:

* `topic_preferences` - (Optional) The contact’s preferences for being opted-in to or opted-out of topics.
* `unsubscribe_all` - (Optional) A boolean value status noting if the contact is unsubscribed from all contact list topics. Default is `false`

### topic_preferences

The following arguments are required:

* `subscription_status` - (Required) The contact’s subscription status to a topic which is either `OPT_IN` or `OPT_OUT`.
* `topic_name` - (Required) The name of the topic.
