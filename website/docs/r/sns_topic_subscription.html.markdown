---
subcategory: "SNS"
layout: "aws"
page_title: "AWS: aws_sns_topic_subscription"
description: |-
  Provides a resource for subscribing to SNS topics.
---

# Resource: aws_sns_topic_subscription

  Provides a resource for subscribing to SNS topics. Requires that an SNS topic exist for the subscription to attach to.
This resource allows you to automatically place messages sent to SNS topics in SQS queues, send them as HTTP(S) POST requests
to a given endpoint, send SMS messages, or notify devices / applications. The most likely use case for Terraform users will
probably be SQS queues.

## Example Usage

You can directly supply a topic and ARN by hand in the `topic_arn` property along with the queue ARN:

```hcl
resource "aws_sns_topic_subscription" "user_updates_sqs_target" {
  topic_arn = "arn:aws:sns:us-west-2:432981146916:user-updates-topic"
  protocol  = "sqs"
  endpoint  = "arn:aws:sqs:us-west-2:432981146916:terraform-queue-too"
}
```

Alternatively you can use the ARN properties of a managed SNS topic and SQS queue:

```hcl
resource "aws_sns_topic" "user_updates" {
  name = "user-updates-topic"
}

resource "aws_sqs_queue" "user_updates_queue" {
  name = "user-updates-queue"
}

resource "aws_sns_topic_subscription" "user_updates_sqs_target" {
  topic_arn = aws_sns_topic.user_updates.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.user_updates_queue.arn
}
```

You can subscribe SNS topics to SQS queues in different Amazon accounts and regions:

-> NOTE:
Terraform must be run on each account individually.
SQS in account `222222222222` and region `us-east-1`
SNS topic and Subscription in account `111111111111` and region `us-west-1`

### SQS Queue (Account Id: 222222222222 /  Region: us-east-1)

```hcl
resource "aws_sqs_queue" "this" {
  name = "example-sqs-queue"
}

resource "aws_sqs_queue_policy" "this" {
  queue_url = aws_sqs_queue.this.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "example-sns-topic",
  "Statement": [
    {
      "Sid": "sid1",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "SQS:SendMessage",
      "Resource": "arn:aws:sqs:us-east-1:222222222222:example-sqs-queue",
      "Condition": {
        "ArnLike": {
          "aws:SourceArn": "arn:aws:sns:us-west-1:111111111111:*"
        }
      }
    },
    {
      "Sid": "Allow-other-account-to-subscribe-to-topic",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::111111111111:root"
      },
      "Action": "sqs:*",
      "Resource": "arn:aws:sns:us-west-1:111111111111:*"
    }
  ]
}
POLICY
}
```

### SNS Topic and Subscription (Account Id: 111111111111 / Region: us-west-1)
```hcl
resource "aws_sns_topic" "this" {
  name = "example-sns-topic"
}

resource "aws_sns_topic_subscription" "this" {
  topic_arn = aws_sns_topic.this.arn
  protocol  = "sqs"
  endpoint  = "arn:aws:sqs:us-east-1:222222222222:example-sqs-queue"
  confirmation_timeout_in_minutes = "5"
  depends_on = [ aws_sns_topic.this ]
}
```

## Argument Reference

The following arguments are supported:

* `topic_arn` - (Required) The ARN of the SNS topic to subscribe to
* `protocol` - (Required) The protocol to use, see below. Refer to the [SNS API docs](https://docs.aws.amazon.com/sns/latest/api/API_Subscribe.html) for more details.
* `endpoint` - (Required) The endpoint to send data to, the contents will vary with the protocol. (see below for more information)
* `endpoint_auto_confirms` - (Deprecated) The endpoint auto confirms exists for historical compatibility and should not be used.
* `confirmation_timeout_in_minutes` - (Optional) Integer indicating number of minutes to wait in retying mode for fetching subscription arn before marking it as failure. You must receive the confirmation message to accept the subscription. (default is 1 minute). Refer to the [SNS docs](https://docs.aws.amazon.com/sns/latest/dg/sns-send-message-to-sqs-cross-account.html) for more details.
* `raw_message_delivery` - (Optional) Boolean indicating whether or not to enable raw message delivery (the original message is directly passed, not wrapped in JSON with the original message in the message property) (default is false).
* `filter_policy` - (Optional) JSON String with the filter policy that will be used in the subscription to filter messages seen by the target resource. Refer to the [SNS docs](https://docs.aws.amazon.com/sns/latest/dg/message-filtering.html) for more details.
* `delivery_policy` - (Optional) JSON String with the delivery policy (retries, backoff, etc.) that will be used in the subscription - this only applies to HTTP/S subscriptions. Refer to the [SNS docs](https://docs.aws.amazon.com/sns/latest/dg/DeliveryPolicies.html) for more details.

### Protocols supported

Supported SNS protocols include:

* `lambda` -- delivery of JSON-encoded message to a lambda function
* `sqs` -- delivery of JSON-encoded message to an Amazon SQS queue
* `application` -- delivery of JSON-encoded message to an EndpointArn for a mobile app and device
* `sms` -- delivery text message
* `http` -- delivery of JSON-encoded messages via HTTP.
* `https` -- delivery of JSON-encoded messages via HTTPS.
* `email` -- delivery of message via SMTP
* `email-json` -- delivery of JSON-encoded message via SMTP

-> NOTE:
You should receive a confirmation message at the configured endpoint and validate the subscription.

### Specifying endpoints

Endpoints have different format requirements according to the protocol that is chosen.

* SQS endpoints come in the form of the SQS queue's ARN (not the URL of the queue) e.g: `arn:aws:sqs:us-west-2:432981146916:terraform-queue-too`
* Application endpoints are also the endpoint ARN for the mobile app and device.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the subscription
* `topic_arn` - The ARN of the topic the subscription belongs to
* `protocol` - The protocol being used
* `endpoint` - The full endpoint to send data to (SQS ARN, HTTP(S) URL, Application ARN, SMS number, etc.)
* `arn` - The ARN of the subscription stored as a more user-friendly property

## Import

SNS Topic Subscriptions can be imported using the `subscription arn`, e.g.

```
$ terraform import aws_sns_topic_subscription.user_updates_sqs_target arn:aws:sns:us-west-2:0123456789012:my-topic:8a21d249-4329-4871-acc6-7be709c6ea7f
```

