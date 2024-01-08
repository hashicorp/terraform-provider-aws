---
subcategory: "Systems Manager for SAP"
layout: "aws"
page_title: "AWS: aws_ssmsap_application"
description: |-
  Terraform resource for managing an AWS Systems Manager for SAP Application.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->`
# Resource: aws_ssmsap_application

Terraform resource for managing an AWS Systems Manager for SAP Application.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssmsap_application" "example" {
		id      = "sap_hana_hdb"
		application_type    = "HANA"
		instances           = ["i-1234567890"]
		sap_instance_number = "00"
		sap_system_id       = "HDB"
	  
	  
		credentials {
		  database_name   = "SYSTEMDB"
		  credential_type = "ADMIN"
		  secret_id       = aws_secretsmanager_secret.systemdb.id
		}
	  
		credentials {
		  database_name   = "HDB"
		  credential_type = "ADMIN"
		  secret_id       = aws_secretsmanager_secret.hdb.id
		}
	  
		depends_on = [aws_ec2_tag.ssmsapmanaged]
}
	  
	  
resource "aws_ec2_tag" "ssmsapmanaged" {
		resource_id = "i-1234567890"
		key         = "SSMForSAPManaged"
		value       = "true"
}

	  
resource "aws_secretsmanager_secret" "systemdb" {
}
	  
resource "aws_secretsmanager_secret_version" "systemdb" {
	  secret_string = "{\"password\":\"Password123@\", \"username\":\"SYSTEM\"}"
		secret_id     = aws_secretsmanager_secret.systemdb.id
}

resource "aws_secretsmanager_secret" "hdb" {
}

resource "aws_secretsmanager_secret_version" "hdb" {
	  secret_string = "{\"password\":\"Password@123\", \"username\":\"SYSTEM\"}"
		secret_id     = aws_secretsmanager_secret.hdb.id
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) ID of the application.

* `application_type` - (Required) Type of the application. 

* `sap_instance_number` - (Required) SAP instance number of the application

* `sap_system_id` - (Required) System ID of the application

* `instances` - (Required) Set of Amazon EC2 instances on which the SAP application is running. Exactly one value expected.

The following arguments are optional:

* `credentials` - (Optional) Configuration Block of database credentials. Multiple values possible. 

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Application.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Systems Manager for SAP Application using the `application_id`. For example:

```terraform
import {
  to = aws_ssmsap_application.example
  id = "sap_hana_hdb"
}
```

Using `terraform import`, import Systems Manager for SAP Application using the `application_id`. For example:

```console
% terraform import aws_ssmsap_application.example sap_hana_hdb
```
