# EventBridge Event sent to SNS Topic

This example sets up an EventBridge Rule with a Target and SNS Topic
to send any CloudTrail API operation into that SNS topic. This allows you
to add SNS subscriptions which may notify you about suspicious activity.

See more details about [EventBridge](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-what-is.html)
in the official AWS docs.

## How to run the example

```
terraform apply -var=aws_region=us-west-2
```
