#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

# shellcheck disable=2157 # This isn't a constant string, it's a TeamCity variable substitution
if [[ -n "%ACCTEST_ROLE_ARN%" ]]; then
    conf=$(pwd)/aws.conf

    function cleanup {
        rm "${conf}"
    }
    trap cleanup EXIT

    touch "${conf}"
    chmod 600 "${conf}"
    cat <<EOF >"${conf}"
[profile perftest]
role_arn       = %ACCTEST_ROLE_ARN%
source_profile = source

[profile source]
aws_access_key_id     = %AWS_ACCESS_KEY_ID%
aws_secret_access_key = %AWS_SECRET_ACCESS_KEY%
EOF

    unset AWS_ACCESS_KEY_ID
    unset AWS_SECRET_ACCESS_KEY

    export AWS_CONFIG_FILE="${conf}"
    export AWS_PROFILE=perftest
fi

echo "make: Resource Identity Smoke Tests"

# aws_batch_job_queue: Framework Regional ARN
SMOKE_IDENTITY_TESTS_BATCH=(
    TestAccBatchJobQueue_Identity_
)

# aws_cloudfront_key_value_store: Framework Global Single-Parameter
# aws_cloudfrontkeyvaluestore_key: Framework Global Multiple-Parameter
SMOKE_IDENTITY_TESTS_CLOUDFRONT=(
    TestAccCloudFrontKeyValueStore_Identity_
    TestAccCloudFrontKeyValueStoreKey_Identity_
)

# aws_globalaccelerator_cross_account_attachment: Framework Global ARN
SMOKE_IDENTITY_TESTS_GLOBALACCELERATOR=(
    TestAccGlobalAcceleratorCrossAccountAttachment_Identity_
)

# aws_iam_policy: SDKv2 Global ARN
# aws_iam_policy_attachment: SDKv2 Global ARN (with rename)
# aws_iam_role: SDKv2 Global Single-Parameter
# aws_iam_role_policy: SDKv2 Global Multiple-Parameter
SMOKE_IDENTITY_TESTS_IAM=(
    TestAccIAMPolicy_Identity_
    TestAccIAMPolicyAttachment_Identity_
    TestAccIAMRole_Identity_
    TestAccIAMRolePolicy_Identity_
)

# aws_lambda_function_scaling_config: Framework Regional Multiple-Parameter
SMOKE_IDENTITY_TESTS_LAMBDA=(
    TestAccLambdaFunctionScalingConfig_Identity_
)

# aws_cloudwatch_log_resource_policy: SDKv2 Regional Multiple-Parameter (with optional)
# aws_cloudwatch_log_transformer: Framework Regional ARN (with rename)
# aws_cloudwatch_log_storage_tier_policy: Framework Regional Singleton
SMOKE_IDENTITY_TESTS_LOGS=(
    TestAccLogsResourcePolicy_Identity_
    TestAccLogsTransformer_Identity_
    TestAccLogs_serial/StorageTierPolicy/Identity
)

# aws_osis_pipeline: Framework Regional Single-Parameter (with rename)
SMOKE_IDENTITY_TESTS_OSIS=(
    TestAccOpenSearchIngestionPipeline_Identity_
)

# aws_rds_certificate: SDKv2 Regional Singleton
SMOKE_IDENTITY_TESTS_RDS=(
    TestAccRDSCertificate_serial/Identity
)

# aws_redshift_namespace_registration: Framework Regional Multiple-Parameter (with optional)
SMOKE_IDENTITY_TESTS_REDSHIFT=(
    TestAccRedshiftNamespaceRegistration_Identity_
)

# aws_route53_record: SDKv2 Global Multiple-Parameter (with rename), (with optional), Mutable
SMOKE_IDENTITY_TESTS_ROUTE53=(
    TestAccRoute53Record_Identity_
)

# aws_s3_bucket: SDKv2 Regional Single-Parameter
# aws_s3_bucket_acl: SDKv2 Identity Schema Upgrader
# aws_s3_directory_bucket: Framework Regional Single-Parameter
# aws_s3_object: SDKv2 Regional Multiple-Parameter
SMOKE_IDENTITY_TESTS_S3=(
    TestAccS3Bucket_Identity_
    TestAccS3BucketACL_Identity_
    TestAccS3DirectoryBucket_Identity_
    TestAccS3Object_Identity_
)

# aws_s3_account_public_access_block: SDKv2 Global Singleton
SMOKE_IDENTITY_TESTS_S3CONTROL=(
    TestAccS3ControlAccountPublicAccessBlock_serial/PublicAccessBlock/Identity
)

# aws_secretsmanager_secret_policy: SDKv2 Regional ARN (with rename)
SMOKE_IDENTITY_TESTS_SECRETSMANAGER=(
    TestAccSecretsManagerSecretPolicy_Identity_
)

# aws_shield_application_layer_automatic_response: Framework Global ARN (with rename)
SMOKE_IDENTITY_TESTS_SHIELD=(
    TestAccShieldApplicationLayerAutomaticResponse_Identity_
)

# aws_sns_topic: SDKv2 Regional ARN
SMOKE_IDENTITY_TESTS_SNS=(
    TestAccSNSTopic_Identity_
)

# aws_sqs_queue: SDKv2 Custom Inherent Regional
SMOKE_IDENTITY_TESTS_SQS=(
    TestAccSQSQueue_Identity_
)

# aws_ssoadmin_application: Framework Global ARN format for regional resource
SMOKE_IDENTITY_TESTS_SSOADMIN=(
    TestAccSSOAdminApplication_Identity_
)

# aws_uxc_account_customizations: Framework Global Singleton
SMOKE_IDENTITY_TESTS_UXC=(
    TestAccUXC_serial/AccountCustomizations/Identity
)

SMOKE_IDENTITY_TESTS=(
    "${SMOKE_IDENTITY_TESTS_BATCH[@]}"
    "${SMOKE_IDENTITY_TESTS_CLOUDFRONT[@]}"
    "${SMOKE_IDENTITY_TESTS_GLOBALACCELERATOR[@]}"
    "${SMOKE_IDENTITY_TESTS_IAM[@]}"
    "${SMOKE_IDENTITY_TESTS_LAMBDA[@]}"
    "${SMOKE_IDENTITY_TESTS_LOGS[@]}"
    "${SMOKE_IDENTITY_TESTS_OSIS[@]}"
    "${SMOKE_IDENTITY_TESTS_RDS[@]}"
    "${SMOKE_IDENTITY_TESTS_REDSHIFT[@]}"
    "${SMOKE_IDENTITY_TESTS_ROUTE53[@]}"
    "${SMOKE_IDENTITY_TESTS_S3[@]}"
    "${SMOKE_IDENTITY_TESTS_S3CONTROL[@]}"
    "${SMOKE_IDENTITY_TESTS_SECRETSMANAGER[@]}"
    "${SMOKE_IDENTITY_TESTS_SHIELD[@]}"
    "${SMOKE_IDENTITY_TESTS_SNS[@]}"
    "${SMOKE_IDENTITY_TESTS_SQS[@]}"
    "${SMOKE_IDENTITY_TESTS_SSOADMIN[@]}"
    "${SMOKE_IDENTITY_TESTS_UXC[@]}"
)

printf -v run_tests '%s|' "${SMOKE_IDENTITY_TESTS[@]}"

TF_ACC=1 go test \
    ./internal/service/batch/... \
    ./internal/service/cloudfront/... \
    ./internal/service/cloudfrontkeyvaluestore/... \
    ./internal/service/globalaccelerator/... \
    ./internal/service/iam/... \
    ./internal/service/lambda/... \
    ./internal/service/logs/... \
    ./internal/service/osis/... \
    ./internal/service/rds/... \
    ./internal/service/redshift/... \
    ./internal/service/route53/... \
    ./internal/service/s3/... \
    ./internal/service/s3control/... \
    ./internal/service/secretsmanager/... \
    ./internal/service/shield/... \
    ./internal/service/sns/... \
    ./internal/service/sqs/... \
    ./internal/service/ssoadmin/... \
    ./internal/service/uxc/... \
    -count 1 -p $(( "%teamcity.agent.hardware.cpuCount%" / 2 )) -timeout 60m -vet=off -buildvcs=false \
    -run="${run_tests%|}"
