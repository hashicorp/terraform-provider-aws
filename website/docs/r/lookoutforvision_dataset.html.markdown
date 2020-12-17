---
subcategory: "Lookout for Vision"
layout: "aws"
page_title: "AWS: aws_lookoutforvision_dataset"
description: |-
  Adds a dataset to an Amazon Lookout for Vision project.
---

# Resource: aws_lookoutforvision_dataset

Adds a dataset to an Amazon Lookout for Vision project.

## Example Usage

Basic usage:

```hcl
resource "aws_lookoutforvision_project" "demo" {
  name = "demo"
}

// Creates a train dataset with samples
resource "aws_lookoutforvision_dataset" "train" {
  project      = aws_lookout_for_vision_project.demo.name
  dataset_type = "train"
  source {
    bucket = "my-bucket"
    key    = "path/to/manifest"
  }
}

// Creates an empty test dataset
resource "aws_lookoutforvision_dataset" "test" {
  project      = aws_lookout_for_vision_project.demo.name
  dataset_type = "test"
}
```

## Argument Reference

The following arguments are supported:

* `project` - (Required) The name of the project in which you want to create a dataset
* `dataset_type` - (Required) The type of the dataset. Specify `train` for a training dataset. Specify `test` for a test dataset
* `source` - (Optional) The location of the manifest file that Amazon Lookout for Vision uses to create the dataset

### Source Configuration

* `bucket` - (Required) The Amazon S3 bucket that contains the manifest
* `key` - (Required) The name and location of the manifest file within the bucket~
* `version_id` - (Optional) The version ID of the bucket
