---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_classifier"
description: |-
  Provides an Glue Classifier resource.
---

# Resource: aws_glue_classifier

Provides a Glue Classifier resource.

~> **NOTE:** It is only valid to create one type of classifier (CSV, grok, JSON, or XML). Changing classifier types will recreate the classifier.

## Example Usage

### CSV Classifier

```terraform
resource "aws_glue_classifier" "example" {
  name = "example"

  csv_classifier {
    allow_single_column    = false
    contains_header        = "PRESENT"
    delimiter              = ","
    disable_value_trimming = false
    header                 = ["example1", "example2"]
    quote_symbol           = "'"
  }
}
```

### Grok Classifier

```terraform
resource "aws_glue_classifier" "example" {
  name = "example"

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}
```

### JSON Classifier

```terraform
resource "aws_glue_classifier" "example" {
  name = "example"

  json_classifier {
    json_path = "example"
  }
}
```

### XML Classifier

```terraform
resource "aws_glue_classifier" "example" {
  name = "example"

  xml_classifier {
    classification = "example"
    row_tag        = "example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `csv_classifier` - (Optional) A classifier for CSV content. Defined below.
* `grok_classifier` – (Optional) A classifier that uses grok patterns. Defined below.
* `json_classifier` – (Optional) A classifier for JSON content. Defined below.
* `name` – (Required) The name of the classifier.
* `xml_classifier` – (Optional) A classifier for XML content. Defined below.

### csv_classifier

* `allow_single_column` - (Optional) Enables the processing of files that contain only one column.
* `contains_header` - (Optional) Indicates whether the CSV file contains a header. This can be one of "ABSENT", "PRESENT", or "UNKNOWN".
* `custom_datatype_configured` - (Optional) Enables the custom datatype to be configured.
* `custom_datatypes` - (Optional) A list of supported custom datatypes. Valid values are `BINARY`, `BOOLEAN`, `DATE`, `DECIMAL`, `DOUBLE`, `FLOAT`, `INT`, `LONG`, `SHORT`, `STRING`, `TIMESTAMP`.
* `delimiter` - (Optional) The delimiter used in the CSV to separate columns.
* `disable_value_trimming` - (Optional) Specifies whether to trim column values.
* `header` - (Optional) A list of strings representing column names.
* `quote_symbol` - (Optional) A custom symbol to denote what combines content into a single column value. It must be different from the column delimiter.
* `serde` – (Optional) The SerDe for processing CSV. Valid values are `OpenCSVSerDe`, `LazySimpleSerDe`, `None`.

### grok_classifier

* `classification` - (Required) An identifier of the data format that the classifier matches, such as Twitter, JSON, Omniture logs, Amazon CloudWatch Logs, and so on.
* `custom_patterns` - (Optional) Custom grok patterns used by this classifier.
* `grok_pattern` - (Required) The grok pattern used by this classifier.

### json_classifier

* `json_path` - (Required) A `JsonPath` string defining the JSON data for the classifier to classify. AWS Glue supports a subset of `JsonPath`, as described in [Writing JsonPath Custom Classifiers](https://docs.aws.amazon.com/glue/latest/dg/custom-classifier.html#custom-classifier-json).

### xml_classifier

* `classification` - (Required) An identifier of the data format that the classifier matches.
* `row_tag` - (Required) The XML tag designating the element that contains each record in an XML document being parsed. Note that this cannot identify a self-closing element (closed by `/>`). An empty row element that contains only attributes can be parsed as long as it ends with a closing tag (for example, `<row item_a="A" item_b="B"></row>` is okay, but `<row item_a="A" item_b="B" />` is not).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Name of the classifier

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Classifiers using their name. For example:

```terraform
import {
  to = aws_glue_classifier.MyClassifier
  id = "MyClassifier"
}
```

Using `terraform import`, import Glue Classifiers using their name. For example:

```console
% terraform import aws_glue_classifier.MyClassifier MyClassifier
```
