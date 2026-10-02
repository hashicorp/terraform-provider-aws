# AWS Resilience Hub V2 (Next Generation Resilience Hub)
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

This example provisions a complete Next Generation AWS Resilience Hub (`resiliencehubv2`)
setup the way a customer would compose it end to end:

* `aws_resiliencehubv2_policy` — a reusable resilience policy (availability SLO + multi-AZ DR targets)
* `aws_resiliencehubv2_system` — a top-level system grouping
* `aws_resiliencehubv2_service` — a service assessed against the policy
* `aws_resiliencehubv2_user_journey` — a critical end-user journey within the system
* `aws_resiliencehubv2_service_function` — a technical workflow subset of the service
* `aws_resiliencehubv2_input_source` — resource discovery from a CloudFormation stack

## Running this example

```console
terraform init
terraform plan
terraform apply
```

The `permission_model` on the service references the `AWSResilienceHubAssessmentRole`
IAM role, which must exist in the account before `apply`.
