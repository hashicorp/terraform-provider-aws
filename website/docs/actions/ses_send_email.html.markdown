---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_send_email"
description: |-
  Sends an email using Amazon SES.
---

# Action: aws_ses_send_email

~> **Note:** `aws_ses_send_email` is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Sends an email using Amazon SES. This action allows for imperative email sending with full control over recipients, content, and formatting.

For information about Amazon SES, see the [Amazon SES Developer Guide](https://docs.aws.amazon.com/ses/latest/dg/). For specific information about sending emails, see the [SendEmail](https://docs.aws.amazon.com/ses/latest/APIReference/API_SendEmail.html) page in the Amazon SES API Reference.

~> **Note:** All email addresses used must be verified in Amazon SES or belong to a verified domain. Due to the difficulty in testing, your help is important in discovering and reporting issues.

## Example Usage

### Basic Usage

```terraform
resource "aws_ses_email_identity" "example" {
  email = "sender@example.com"
}

action "aws_ses_send_email" "example" {
  config {
    source       = aws_ses_email_identity.example.email
    subject      = "Test Email"
    text_body    = "This is a test email sent from Terraform."
    to_addresses = ["recipient@example.com"]
  }
}

resource "terraform_data" "example" {
  input = "send-notification"

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_ses_send_email.example]
    }
  }
}
```

### HTML Email with Multiple Recipients

```terraform
action "aws_ses_send_email" "newsletter" {
  config {
    source             = aws_ses_email_identity.marketing.email
    subject            = "Monthly Newsletter - ${formatdate("MMMM YYYY", timestamp())}"
    html_body          = "<h1>Welcome!</h1><p>This is our <strong>monthly newsletter</strong>.</p>"
    to_addresses       = var.subscriber_emails
    cc_addresses       = ["manager@example.com"]
    reply_to_addresses = ["support@example.com"]
    return_path        = "bounces@example.com"
  }
}
```

### Deployment Notification

```terraform
action "aws_ses_send_email" "deploy_notification" {
  config {
    source       = "deployments@example.com"
    subject      = "Deployment Complete: ${var.environment}"
    text_body    = "Application ${var.app_name} has been successfully deployed to ${var.environment}."
    to_addresses = var.team_emails
  }
}

resource "terraform_data" "deployment" {
  input = var.deployment_id

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_ses_send_email.deploy_notification]
    }
  }

  depends_on = [aws_instance.app]
}
```

### Alert Email with Dynamic Content

```terraform
locals {
  alert_body = templatefile("${path.module}/templates/alert.txt", {
    service     = var.service_name
    environment = var.environment
    timestamp   = timestamp()
    details     = var.alert_details
  })
}

action "aws_ses_send_email" "alert" {
  config {
    source       = "alerts@example.com"
    subject      = "ALERT: ${var.service_name} Issue Detected"
    text_body    = local.alert_body
    to_addresses = var.oncall_emails
    cc_addresses = var.manager_emails
  }
}
```

### Multi-format Email

```terraform
action "aws_ses_send_email" "welcome" {
  config {
    source    = aws_ses_email_identity.noreply.email
    subject   = "Welcome to ${var.company_name}!"
    text_body = "Welcome! Thank you for joining us. Visit our website for more information."
    html_body = templatefile("${path.module}/templates/welcome.html", {
      user_name    = var.user_name
      company_name = var.company_name
      website_url  = var.website_url
    })
    to_addresses = [var.user_email]
  }
}
```

### Conditional Email Sending

```terraform
action "aws_ses_send_email" "conditional" {
  config {
    source       = "notifications@example.com"
    subject      = var.environment == "production" ? "Production Alert" : "Test Alert"
    text_body    = "This is a ${var.environment} environment notification."
    to_addresses = var.environment == "production" ? var.prod_emails : var.dev_emails
  }
}
```

### Batch Processing Notification

```terraform
action "aws_ses_send_email" "batch_complete" {
  config {
    source       = "batch-jobs@example.com"
    subject      = "Batch Processing Complete - ${var.job_name}"
    html_body    = <<-HTML
      <h2>Batch Job Results</h2>
      <p><strong>Job:</strong> ${var.job_name}</p>
      <p><strong>Records Processed:</strong> ${var.records_processed}</p>
      <p><strong>Duration:</strong> ${var.processing_duration}</p>
      <p><strong>Status:</strong> ${var.job_status}</p>
    HTML
    to_addresses = var.admin_emails
  }
}
```

## Argument Reference

This action supports the following arguments:

* `bcc_addresses` - (Optional) List of email addresses for the BCC: field of the message. Recipients in this list will receive the email but their addresses will not be visible to other recipients.
* `cc_addresses` - (Optional) List of email addresses for the CC: field of the message. Recipients in this list will receive the email and their addresses will be visible to all recipients.
* `html_body` - (Optional) Message body in HTML format. Either `text_body` or `html_body` (or both) must be specified. HTML content allows for rich formatting including links, images, and styling.
* `reply_to_addresses` - (Optional) List of reply-to email addresses for the message. If the recipient replies to the message, each reply-to address will receive the reply. If not specified, replies will go to the source address.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `return_path` - (Optional) Email address that bounces and complaints will be forwarded to when feedback forwarding is enabled. This is useful for handling delivery failures and spam complaints.
* `source` - (Required) Email address that is sending the email. This address must be either individually verified with Amazon SES, or from a domain that has been verified with Amazon SES.
* `subject` - (Required) Subject of the message: A short summary of the content, which will appear in the recipient's inbox.
* `text_body` - (Optional) Message body in text format. Either `text_body` or `html_body` (or both) must be specified. Text format ensures compatibility with all email clients.
* `to_addresses` - (Optional) List of email addresses for the To: field of the message. These are the primary recipients of the email.
