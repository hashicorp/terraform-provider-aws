<!---
See what makes a good Pull Request at: https://hashicorp.github.io/terraform-provider-aws/raising-a-pull-request/
--->

<!-- heimdall_github_prtemplate:grc-pci_dss-2024-01-05 -->

## Rollback Plan

If a change needs to be reverted, we will publish an updated version of the library.

## Changes to Security Controls

Are there any changes to security controls (access controls, encryption, logging) in this pull request? If so, explain.

No changes to security controls.

### Description

Make `agent_arns` field optional in `aws_datasync_location_object_storage` resource.

Currently, the `agent_arns` field is marked as required in Terraform, which prevents users from creating DataSync Object Storage locations without specifying agent ARNs. However, the AWS Console allows creating these locations without agents, and when importing such console-created resources into Terraform, the `agent_arns` field appears as an empty array `[]`.

This change allows users to create DataSync Object Storage locations without specifying agent ARNs, which is useful for scenarios where agents are not required or will be configured separately. This aligns the Terraform resource behavior with the AWS Console and API capabilities.

**Changes made:**
- Changed `agent_arns` field from `Required: true` to `Optional: true`
- Added conditional logic to set `AgentArns` only when the field is specified
- Updated both `Create` and `Update` operations to handle empty `agent_arns`
- Added comprehensive test coverage for both empty array and omitted field scenarios

### Relations

Closes #0000

### References

AWS DataSync Documentation: https://docs.aws.amazon.com/datasync/latest/userguide/create-locations.html

### Output from Acceptance Testing

```console
% TF_ACC=1 go test -v -timeout=60m -run "TestAccDataSyncLocationObjectStorage_emptyAgentArns" ./internal/service/datasync/

=== RUN   TestAccDataSyncLocationObjectStorage_emptyAgentArns
=== PAUSE TestAccDataSyncLocationObjectStorage_emptyAgentArns
=== CONT  TestAccDataSyncLocationObjectStorage_emptyAgentArns
--- PASS: TestAccDataSyncLocationObjectStorage_emptyAgentArns (11.74s)
PASS
ok      github.com/hashicorp/terraform-provider-aws/internal/service/datasync   21.194s
```

```console
% TF_ACC=1 go test -v -timeout=60m -run "TestAccDataSyncLocationObjectStorage_noAgentArns" ./internal/service/datasync/

=== RUN   TestAccDataSyncLocationObjectStorage_noAgentArns
=== PAUSE TestAccDataSyncLocationObjectStorage_noAgentArns
=== CONT  TestAccDataSyncLocationObjectStorage_noAgentArns
--- PASS: TestAccDataSyncLocationObjectStorage_noAgentArns (9.99s)
PASS
ok      github.com/hashicorp/terraform-provider-aws/internal/service/datasync   19.420s
```

```console
% go test -v -run "TestDecodeObjectStorageURI" ./internal/service/datasync/

=== RUN   TestDecodeObjectStorageURI
=== PAUSE TestDecodeObjectStorageURI
=== CONT  TestDecodeObjectStorageURI
=== RUN   TestDecodeObjectStorageURI/empty_URI
=== RUN   TestDecodeObjectStorageURI/S3_bucket_URI_top_level
=== RUN   TestDecodeObjectStorageURI/Object_storage_top_level
=== RUN   TestDecodeObjectStorageURI/Object_storage_one_level
=== RUN   TestDecodeObjectStorageURI/Object_storage_two_levels
--- PASS: TestDecodeObjectStorageURI (0.00s)
    --- PASS: TestDecodeObjectStorageURI/empty_URI (0.00s)
    --- PASS: TestDecodeObjectStorageURI/Object_storage_two_levels (0.00s)
    --- PASS: TestDecodeObjectStorageURI/S3_bucket_URI_top_level (0.00s)
    --- PASS: TestDecodeObjectStorageURI/Object_storage_one_level (0.00s)
    --- PASS: TestDecodeObjectStorageURI/Object_storage_top_level (0.00s)
PASS
ok      github.com/hashicorp/terraform-provider-aws/internal/service/datasync   8.806s
``` 