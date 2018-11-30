---
layout: "aws"
page_title: "AWS: aws_glue_script"
sidebar_current: "docs-aws-datasource-glue-script"
description: |-
  Generate Glue script from Directed Acyclic Graph
---

# Data Source: aws_glue_script

Use this data source to generate a Glue script from a Directed Acyclic Graph (DAG).

## Example Usage

### Generate Python Script

```hcl
data "aws_glue_script" "example" {
  language = "PYTHON"

  dag_edge = []

  # ...

  dag_node = []

  # ...
}

output "python_script" {
  value = "${data.aws_glue_script.example.python_script}"
}
```

### Generate Scala Code

```hcl
data "aws_glue_script" "example" {
  language = "SCALA"

  dag_edge = []

  # ...

  dag_node = []

  # ...
}

output "scala_code" {
  value = "${data.aws_glue_script.example.scala_code}"
}
```

## Argument Reference

* `dag_edge` - (Required) A list of the edges in the DAG. Defined below.
* `dag_node` - (Required) A list of the nodes in the DAG. Defined below.
* `language` - (Optional) The programming language of the resulting code from the DAG. Defaults to `PYTHON`. Valid values are `PYTHON` and `SCALA`.

### dag_edge Argument Reference

* `source` - (Required) The ID of the node at which the edge starts.
* `target` - (Required) The ID of the node at which the edge ends.
* `target_parameter` - (Optional) The target of the edge.

### dag_node Argument Reference

* `args` - (Required) Nested configuration an argument or property of a node. Defined below.
* `id` - (Required) A node identifier that is unique within the node's graph.
* `node_type` - (Required) The type of node this is.
* `line_number` - (Optional) The line number of the node.

#### args Argument Reference

* `name` - (Required) The name of the argument or property.
* `value` - (Required) The value of the argument or property.
* `param` - (Optional) Boolean if the value is used as a parameter. Defaults to `false`.

## Attributes Reference

* `python_script` - The Python script generated from the DAG when the `language` argument is set to `PYTHON`.
* `scala_code` - The Scala code generated from the DAG when the `language` argument is set to `SCALA`.
