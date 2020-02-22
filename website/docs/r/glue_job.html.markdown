---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_job"
description: |-
  Provides an Glue Job resource.
---

# Resource: aws_glue_job

Provides a Glue Job resource.

-> Glue functionality, such as monitoring and logging of jobs, is typically managed with the `default_arguments` argument. See the [Special Parameters Used by AWS Glue](https://docs.aws.amazon.com/glue/latest/dg/aws-glue-programming-etl-glue-arguments.html) topic in the Glue developer guide for additional information.

## Example Usage

### Python Job

```hcl
resource "aws_glue_job" "example" {
  name     = "example"
  role_arn = "${aws_iam_role.example.arn}"

  command {
    script_location = "s3://${aws_s3_bucket.example.bucket}/example.py"
  }
}
```

### Scala Job

```hcl
resource "aws_glue_job" "example" {
  name     = "example"
  role_arn = "${aws_iam_role.example.arn}"

  command {
    script_location = "s3://${aws_s3_bucket.example.bucket}/example.scala"
  }

  default_arguments = {
    "--job-language" = "scala"
  }
}
```

### Enabling CloudWatch Logs and Metrics

```hcl
resource "aws_cloudwatch_log_group" "example" {
  name              = "example"
  retention_in_days = 14
}

resource "aws_glue_job" "example" {
  # ... other configuration ...

  default_arguments = {
    # ... potentially other arguments ...
    "--continuous-log-logGroup"          = "${aws_cloudwatch_log_group.example.name}"
    "--enable-continuous-cloudwatch-log" = "true"
    "--enable-continuous-log-filter"     = "true"
    "--enable-metrics"                   = ""
  }
}
```

## Argument Reference

The following arguments are supported:

~> **NOTE:** The `allocated_capacity` attribute has been deprecated and might
be removed in future releases, please use `max_capacity` instead.

* `allocated_capacity` – **DEPRECATED** (Optional) The number of AWS Glue data processing units (DPUs) to allocate to this Job. At least 2 DPUs need to be allocated; the default is 10. A DPU is a relative measure of processing power that consists of 4 vCPUs of compute capacity and 16 GB of memory.
* `command` – (Required) The command of the job. Defined below.
* `connections` – (Optional) The list of connections used for this job.
* `default_arguments` – (Optional) The map of default arguments for this job. Defined below.
* `description` – (Optional) Description of the job.
* `execution_property` – (Optional) Execution property of the job. Defined below.
* `glue_version` - (Optional) The version of glue to use, for example "1.0". For information about available versions, see the [AWS Glue Release Notes](https://docs.aws.amazon.com/glue/latest/dg/release-notes.html).
* `max_capacity` – (Optional) The maximum number of AWS Glue data processing units (DPUs) that can be allocated when this job runs. `Required` when `pythonshell` is set, accept either `0.0625` or `1.0`.
* `max_retries` – (Optional) The maximum number of times to retry this job if it fails.
* `name` – (Required) The name you assign to this job. It must be unique in your account.
* `notification_property` - (Optional) Notification property of the job. Defined below.
* `role_arn` – (Required) The ARN of the IAM role associated with this job.
* `tags` - (Optional) Key-value mapping of resource tags
* `timeout` – (Optional) The job timeout in minutes. The default is 2880 minutes (48 hours).
* `security_configuration` - (Optional) The name of the Security Configuration to be associated with the job.
* `worker_type` - (Optional) The type of predefined worker that is allocated when a job runs. Accepts a value of Standard, G.1X, or G.2X.
* `number_of_workers` - (Optional) The number of workers of a defined workerType that are allocated when a job runs.

### command Argument Reference

* `name` - (Optional) The name of the job command. Defaults to `glueetl`. Use `pythonshell` for Python Shell Job Type, `max_capacity` needs to be set if `pythonshell` is chosen.
* `script_location` - (Required) Specifies the S3 path to a script that executes a job.
* `python_version` - (Optional) The Python version being used to execute a Python shell job. Allowed values are 2 or 3.

### default_arguments Argument Reference

You can specify arguments here that your own job-execution script consumes, as well as arguments that AWS Glue itself consumes.

#### Reserved Key-Pairs
* `"--job-language"` - (Optional) The script programming language. This must be either `scala` or `python`. If this parameter is not present, the default is python.
* `"--class"` - (Optional) The Scala class that serves as the entry point for your Scala script. This only applies if your `--job-language` is set to `scala`.
* `"--extra-py-files"` - (Optional) The Amazon S3 paths to additional Python modules that AWS Glue adds to the Python path before executing your script. Multiple values must be complete paths separated by a comma (,). Only individual files are supported, not a directory path. Currently, only pure Python modules work. Extension modules written in C or other languages are not supported.
* `"--extra-jars"` - (Optional) The Amazon S3 paths to additional Java .jar files that AWS Glue adds to the Java classpath before executing your script. Multiple values must be complete paths separated by a comma (,).
* `"--extra-files"` - (Optional) The Amazon S3 paths to additional files such as configuration files that AWS Glue copies to the working directory of your script before executing it. Multiple values must be complete paths separated by a comma (,). Only individual files are supported, not a directory path.
* `"--job-bookmark-option"` - (Optional) Controls the behavior of a job bookmark. The following option values can be set.
  * `"job-bookmark-enable"` - Keep track of previously processed data. When a job runs, process new data since the last checkpoint.
  * `"job-bookmark-disable"` - Always process the entire dataset. You are responsible for managing the output from previous job runs.
  * `"job-bookmark-pause"` - Process incremental data since the last successful run or the data in the range identified by the following suboptions, without updating the state of last bookmark. You are responsible for managing the output from previous job runs. The following are the two suboptions:
    * `"job-bookmark-from <from-value>"`- (Optional) The run ID that represents all the input that was processed until the last successful run before and including the specified run ID. The corresponding input is ignored.
    * `"job-bookmark-to <to-value>"` - (Optional) The run ID that represents all the input that was processed until the last successful run before and including the specified run ID. The corresponding input excluding the input identified by the `<from-value>` is processed by the job. Any input later than this input is also excluded for processing.
    
    The job bookmark state is not updated when this option set is specified.
    
    The suboptions are optional. However, when used, both suboptions must be provided.
* `"--TempDir"` - (Optional) Specifies an Amazon S3 path to a bucket that can be used as a temporary directory for the Job.
* `"--enable-metrics"` - (Optional) Enables the collection of metrics for job profiling for this job run. These metrics are available on the AWS Glue console and the Amazon CloudWatch console. To enable metrics, only specify the key; no value is needed.
* `"--enable-glue-datacatalog"` - (Optional) Enables you to use the AWS Glue Data Catalog as an Apache Spark Hive metastore.
* `"--enable-continuous-cloudwatch-log"` - (Optional) Enables real-time, continuous logging for AWS Glue jobs. You can view real-time Apache Spark job logs in CloudWatch.
* `"--enable-continuous-log-filter"` - (Optional) Specifies a standard filter (true) or no filter (false) when you create or edit a job enabled for continuous logging. Choosing the standard filter prunes out non-useful Apache Spark driver/executor and Apache Hadoop YARN heartbeat log messages. Choosing no filter gives you all the log messages.
* `"--continuous-log-logGroup"` - (Optional) Specifies a custom CloudWatch log group name for a job enabled for continuous logging.
* `"--continuous-log-logStreamPrefix"` - (Optional) Specifies a custom CloudWatch log stream prefix for a job enabled for continuous logging.
* `"--continuous-log-conversionPattern"` - (Optional) Specifies a custom conversion log pattern for a job enabled for continuous logging. The conversion pattern only applies to driver logs and executor logs. It does not affect the AWS Glue progress bar.

For information about the key-value pairs that AWS Glue consumes to set up your job, see the [Special Parameters Used by AWS Glue](http://docs.aws.amazon.com/glue/latest/dg/aws-glue-programming-python-glue-arguments.html) topic in the developer guide.

#### Passing Job Arguments
Any non-reserved key can be used to pass your own Job arguments, e.g.
* `"--myFirstArgKey" = "myFirstArgValue"`
* `"--mySecondArgKey" = "mySecondArgValue"`

For information about how to specify and consume your own Job arguments, see the [Calling AWS Glue APIs in Python](http://docs.aws.amazon.com/glue/latest/dg/aws-glue-programming-python-calling.html) topic in the developer guide.

### execution_property Argument Reference

* `max_concurrent_runs` - (Optional) The maximum number of concurrent runs allowed for a job. The default is 1.

### notification_property Argument Reference

* `notify_delay_after` - (Optional) After a job run starts, the number of minutes to wait before sending a job run delay notification.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of Glue Job
* `id` - Job name

## Import

Glue Jobs can be imported using `name`, e.g.

```
$ terraform import aws_glue_job.MyJob MyJob
```
