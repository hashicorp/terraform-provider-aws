---
subcategory: "Compute Optimizer"
layout: "aws"
page_title: "AWS: aws_computeoptimizer_recommendation_preferences"
description: |-
  Terraform resource for managing an AWS Compute Optimizer Recommendation Preferences.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->`
# Resource: aws_computeoptimizer_recommendation_preferences

Terraform resource for managing an AWS Compute Optimizer Recommendation Preferences.

## Example Usage

### Basic Usage

```terraform
resource "aws_computeoptimizer_recommendation_preferences" "example" {
  enhanced_infrastructure_metrics = "Active" 
	external_metrics_preference = "DataDog"
 	inferred_workload_types          = "Active"
	lookback_period = "DAYS_14"
	preferred_resources = [
		""
	]
	resource_type                    = "Ec2Instance"
	savings_estimation_mode = "AfterDiscounts"

	utilization_prefrences = "P99_5"
	scope = {
	  name  = "ResourceARN"
	  value = aws_autoscaling_group.web.arn
	}
}
```

## Argument Reference

The following arguments are required:

* `resource_type` - (Required) Target resource type.

The following arguments are optional:

* `enhanced_infrastructure_metrics` - (Optional) Enhanced infrastructure metrics.

* `external_metrics_preference` - (Optional) External metrics provider.

* `inferred_workload_types` - (Optional) Inferred workload status.

* `look_back_period` - (Optional) Number of days analyzed.

* `preferred_resources` - (Optional) Resource types considered.

* `savings_estimation_mode` - (Optional) Savings estimation status.

* `scope` - (Optional) Recommendation preferences.

* `utilization_preferences` - (Optional) CPU utilization thresholds.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS Account the recommendation is created in.
# TODO - Verify attribute references.
* `recommendation_preference_names` - Recommendation preference created.
* `resource_type` - Resource types monitored.


## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Compute Optimizer Recommendation Preferences using the `resource_type`, and `recommendation_preference_names` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_computeoptimizer_recommendation_preferences.example
  id = "Ec2Instance:UtilizationPreferences"
}
```

Using `terraform import`, import Compute Optimizer Recommendation Preferences using the `example_id_arg`. For example:

```console
% terraform import aws_computeoptimizer_recommendation_preferences.example Ec2Instance:UtilizationPreferences
```