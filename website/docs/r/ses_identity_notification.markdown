---
layout: "aws"
page_title: "AWS: ses_identity_notification"
sidebar_current: "docs-aws-resource-ses-identity-notification"
description: |-
  Setting SES Identity Notification
---

# ses_identity_notification

Setting SES Identity Notification for a domain

## Example Usage

```hcl
resource "ses_identity_notification" "test" {
  topic_arn = "${aws_sns_topic.example.arn}"
  notification_type = "Bounce"
  identity = "${aws_ses_domain_identity.example.domain}"
}
```

## Argument Reference

The following arguments are supported:

* `topic_arn` - (Required) The SNS topic to push the notificaiton to
* `notification_type` - (Required) The type of notification to configure, *Bounce*, *Complaint* or *Delivery*.
* `identity` - (Required) The domain identity to configure the notification for

#