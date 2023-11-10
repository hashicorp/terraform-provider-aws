---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_ontap_storage_virtual_machines"
description: |-
  This resource can be useful for getting back a set of FSx ONTAP Storage Virtual Machine (SVM) IDs.
---

# Data Source: aws_fsx_ontap_storage_virtual_machines

This resource can be useful for getting back a set of FSx ONTAP Storage Virtual Machine (SVM) IDs.

## Example Usage

The following shows outputting all SVM IDs for a given FSx ONTAP File System.

```
data "aws_fsx_ontap_storage_virtual_machines" "example" {
  filter {
    name   = "file-system-id"
    values = ["fs-12345678"]
  }
}
```

## Argument Reference

* `filter` - (Optional) Configuration block. Detailed below.

### filter

This block allows for complex filters.

The following arguments are required:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/fsx/latest/APIReference/API_StorageVirtualMachineFilter.html).
* `values` - (Required) Set of values that are accepted for the given field. An SVM will be selected if any one of the given values matches.

## Attributes Reference

* `ids` - List of all SVM IDs found.
