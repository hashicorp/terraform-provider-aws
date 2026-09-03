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

unset TF_LOG

function tester {
    local pkg=$1
    local tests=$2

    local tmp
    tmp=$(mktemp)

    # When `-json` flag is set, some error conditions show no output if output is captured and then `echo`ed.
    # `tee` results to a temp file so that the "text file busy" error case can be handled correctly.
    #
    # Disable errexit around the pipeline so that a non-zero exit from `go test` does not
    # abort the script before PIPESTATUS can be read and handled explicitly below.
    set +e
    TF_ACC=1 go test ./"${pkg}"/... -v -json -parallel 4 -run="${tests}" -timeout 60m -count 1 -vet=off -buildvcs=false 2>&1 | tee "${tmp}"
    local exit_code=${PIPESTATUS[0]}
    set -e

    if grep -qF "text file busy" "${tmp}"; then
        rm -f "${tmp}"
        echo "FAILED attempt to run tests"
        echo "Trying again..."
        sleep 5
        tester "${pkg}" "${tests}"
        return
    fi

    rm -f "${tmp}"

    if [[ "${exit_code}" -ne 0 ]]; then
        exit "${exit_code}"
    fi
}

if [[ ! -f "iamsanity.test" ]]; then
    SMOKE_TESTS_IAM=(
        TestAccIAMRole_basic
        TestAccIAMRole_namePrefix
        TestAccIAMRole_disappears
        TestAccIAMRole_InlinePolicy_basic
        TestAccIAMPolicyDocumentDataSource_basic
        TestAccIAMPolicyDocumentDataSource_sourceConflicting
        TestAccIAMPolicyDocumentDataSource_sourcePolicyValidJSON
        TestAccIAMRolePolicyAttachment_basic
        TestAccIAMRolePolicyAttachment_disappears
        TestAccIAMRolePolicyAttachment_Disappears_role
        TestAccIAMPolicy_basic
        TestAccIAMPolicy_policy
        TestAccIAMPolicy_tags
        TestAccIAMRolePolicy_basic
        TestAccIAMRolePolicy_unknownsInPolicy
        TestAccIAMInstanceProfile_basic
        TestAccIAMInstanceProfile_tags
        TestAccIAMPolicy_List_basic
        TestAccIAMRole_Identity_basic
    )
    printf -v iam_tests '^%s$|' "${SMOKE_TESTS_IAM[@]}"
    tester "internal/service/iam" "${iam_tests%|}"
    touch iamsanity.test
    exit 0
fi

if [[ ! -f "logssanity.test" ]]; then
    SMOKE_TESTS_LOGS=(
        TestAccLogsLogGroup_basic
        TestAccLogsLogGroup_multiple
    )
    printf -v logs_tests '^%s$|' "${SMOKE_TESTS_LOGS[@]}"
    tester "internal/service/logs" "${logs_tests%|}"
    touch logssanity.test
    exit 0
fi

if [[ ! -f "ec2sanity.test" ]]; then
    SMOKE_TESTS_EC2=(
        TestAccVPCSecurityGroup_basic
        TestAccVPCSecurityGroup_egressMode
        TestAccVPCSecurityGroup_vpcAllEgress
        TestAccVPCSecurityGroupRule_race
        TestAccVPCSecurityGroupRule_protocolChange
        TestAccVPCDataSource_basic
        TestAccVPCSubnet_basic
        TestAccVPC_tenancy
        TestAccVPCRouteTableAssociation_Subnet_basic
        TestAccVPCRouteTable_basic
    )
    printf -v ec2_tests '^%s$|' "${SMOKE_TESTS_EC2[@]}"
    tester "internal/service/ec2" "${ec2_tests%|}"
    touch ec2sanity.test
    exit 0
fi

if [[ ! -f "ecssanity.test" ]]; then
    SMOKE_TESTS_ECS=(
        TestAccECSTaskDefinition_basic
        TestAccECSService_basic
    )
    printf -v ecs_tests '^%s$|' "${SMOKE_TESTS_ECS[@]}"
    tester "internal/service/ecs" "${ecs_tests%|}"
    touch ecssanity.test
    exit 0
fi

if [[ ! -f "elbv2sanity.test" ]]; then
    SMOKE_TESTS_ELBV2=(
        TestAccELBV2TargetGroup_basic
    )
    printf -v elbv2_tests '^%s$|' "${SMOKE_TESTS_ELBV2[@]}"
    tester "internal/service/elbv2" "${elbv2_tests%|}"
    touch elbv2sanity.test
    exit 0
fi

if [[ ! -f "eventssanity.test" ]]; then
    SMOKE_TESTS_EVENTS=(
        TestAccEventsPutEventsAction_basic
    )
    printf -v events_tests '^%s$|' "${SMOKE_TESTS_EVENTS[@]}"
    tester "internal/service/events" "${events_tests%|}"
    touch eventssanity.test
    exit 0
fi

if [[ ! -f "kmssanity.test" ]]; then
    SMOKE_TESTS_KMS=(
        TestAccKMSKey_basic
    )
    printf -v kms_tests '^%s$|' "${SMOKE_TESTS_KMS[@]}"
    tester "internal/service/kms" "${kms_tests%|}"
    touch kmssanity.test
    exit 0
fi

if [[ ! -f "lambdasanity.test" ]]; then
    SMOKE_TESTS_LAMBDA=(
        TestAccLambdaFunction_basic
        TestAccLambdaPermission_basic
        TestAccLambdaCapacityProvider_List_basic
    )
    printf -v lambda_tests '^%s$|' "${SMOKE_TESTS_LAMBDA[@]}"
    tester "internal/service/lambda" "${lambda_tests%|}"
    touch lambdasanity.test
    exit 0
fi

if [[ ! -f "metasanity.test" ]]; then
    SMOKE_TESTS_META=(
        TestAccMetaRegionDataSource_basic
        TestAccMetaRegionDataSource_endpoint
        TestAccMetaPartitionDataSource_basic
    )
    printf -v meta_tests '^%s$|' "${SMOKE_TESTS_META[@]}"
    tester "internal/service/meta" "${meta_tests%|}"
    touch metasanity.test
    exit 0
fi

if [[ ! -f "route53sanity.test" ]]; then
    SMOKE_TESTS_ROUTE53=(
        TestAccRoute53Record_basic_FullName
        TestAccRoute53Record_basic_ShortName
        TestAccRoute53Record_Latency_basic
        TestAccRoute53ZoneDataSource_name
    )
    printf -v route53_tests '^%s$|' "${SMOKE_TESTS_ROUTE53[@]}"
    tester "internal/service/route53" "${route53_tests%|}"
    touch route53sanity.test
    exit 0
fi

if [[ ! -f "s3sanity.test" ]]; then
    SMOKE_TESTS_S3=(
        TestAccS3Bucket_Basic_basic
        TestAccS3Bucket_Security_corsUpdate
        TestAccS3BucketPublicAccessBlock_basic
        TestAccS3BucketPolicy_basic
        TestAccS3BucketACL_updateACL
        TestAccS3Object_basic
    )
    printf -v s3_tests '^%s$|' "${SMOKE_TESTS_S3[@]}"
    tester "internal/service/s3" "${s3_tests%|}"
    touch s3sanity.test
    exit 0
fi

if [[ ! -f "ssmsanity.test" ]]; then
    SMOKE_TESTS_SSM=(
        TestAccSSMParameterEphemeral_basic
    )
    printf -v ssm_tests '^%s$|' "${SMOKE_TESTS_SSM[@]}"
    tester "internal/service/ssm" "${ssm_tests%|}"
    touch ssmsanity.test
    exit 0
fi

if [[ ! -f "secretsmanagersanity.test" ]]; then
    SMOKE_TESTS_SECRETSMANAGER=(
        TestAccSecretsManagerSecret_basic
    )
    printf -v secretsmanager_tests '^%s$|' "${SMOKE_TESTS_SECRETSMANAGER[@]}"
    tester "internal/service/secretsmanager" "${secretsmanager_tests%|}"
    touch secretsmanagersanity.test
    exit 0
fi

if [[ ! -f "stssanity.test" ]]; then
    SMOKE_TESTS_STS=(
        TestAccSTSCallerIdentityDataSource_basic
    )
    printf -v sts_tests '^%s$|' "${SMOKE_TESTS_STS[@]}"
    tester "internal/service/sts" "${sts_tests%|}"
    touch stssanity.test
    exit 0
fi

if [[ ! -f "functionsanity.test" ]]; then
    SMOKE_TESTS_FUNCTION=(
        TestARNParseFunction_known
    )
    printf -v function_tests '^%s$|' "${SMOKE_TESTS_FUNCTION[@]}"
    tester "internal/function" "${function_tests%|}"
    touch functionsanity.test
    exit 0
fi

echo "##teamcity[notification notifier='slack' message='*Sanity Tests Passed!*:white_check_mark:' sendTo='CN0G9S7M4' connectionId='PROJECT_EXT_8']"
