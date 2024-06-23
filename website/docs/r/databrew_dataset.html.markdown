---
subcategory: "DataBrew"
layout: "aws"
page_title: "AWS: aws_databrew_dataset"
description: |-
  Terraform resource for managing an AWS DataBrew Dataset.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->`
# Resource: aws_databrew_project

Terraform resource for managing an AWS DataBrew Dataset.

## Example Usage

### Basic Usage

```terraform
resource "aws_databrew_dataset" "example" {
    name = "test"
    input {
        s3_input_definition {
            bucket = "bucket-name"
        }
    }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the dataset to be created.

* `input` - (Required) Represents information on how DataBrew can find data, in either the AWS Glue Data Catalog or Amazon S3.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataBrew Project using the `example_id_arg`. For example:

```terraform
import {
  to = aws_databrew_dataset.example
  id = "project-dataset-name"
}
```

Using `terraform import`, import DataBrew Dataset using the `example_id_arg`. For example:

```console
% terraform import aws_databrew_dataset.example dataset-name
```
