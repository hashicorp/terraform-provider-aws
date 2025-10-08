#!/usr/bin/env bash

set -euo pipefail

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
    local service=$1
    local tests=$2

    local results=`TF_ACC=1 go test ./internal/service/"${service}"/... -v -parallel 4 -run="${tests}" -timeout 60m 2>&1`
    local exit_code=$?

    echo "${results}"

    if [[ "${results}" == *"text file busy"* ]]; then
        echo "FAILED attempt to run tests"
        echo "Trying again..."
        sleep 5
        tester "${service}" "${tests}"
    fi

    if [ "${exit_code}" -ne 0 ]; then
        exit "${exit_code}"
    fi
}

if [ ! -f "iamsanity.test" ]; then
    tester "iam" 'TestAccIAMRole_basic|TestAccIAMRole_namePrefix|TestAccIAMRole_disappears|TestAccIAMRole_InlinePolicy_basic|TestAccIAMPolicyDocumentDataSource_basic|TestAccIAMPolicyDocumentDataSource_sourceConflicting|TestAccIAMPolicyDocumentDataSource_sourceJSONValidJSON|TestAccIAMRolePolicyAttachment_basic|TestAccIAMRolePolicyAttachment_disappears|TestAccIAMRolePolicyAttachment_Disappears_role|TestAccIAMPolicy_basic|TestAccIAMPolicy_policy|TestAccIAMPolicy_tags|TestAccIAMRolePolicy_basic|TestAccIAMRolePolicy_unknownsInPolicy|TestAccIAMInstanceProfile_basic|TestAccIAMInstanceProfile_tags'
    touch iamsanity.test
    exit 0
fi

if [ ! -f "logssanity.test" ]; then
    tester "logs" 'TestAccLogsGroup_basic|TestAccLogsGroup_multiple'
    touch logssanity.test
    exit 0
fi

if [ ! -f "ec2sanity.test" ]; then
    tester "ec2" 'TestAccVPCSecurityGroup_basic|TestAccVPCSecurityGroup_egressMode|TestAccVPCSecurityGroup_vpcAllEgress|TestAccVPCSecurityGroupRule_race|TestAccVPCSecurityGroupRule_protocolChange|TestAccVPCDataSource_basic|TestAccVPCSubnet_basic|TestAccVPC_tenancy|TestAccVPCRouteTableAssociation_Subnet_basic|TestAccVPCRouteTable_basic'
    touch ec2sanity.test
    exit 0
fi

if [ ! -f "ecssanity.test" ]; then
    tester "ecs" 'TestAccECSTaskDefinition_basic|TestAccECSService_basic'
    touch ecssanity.test
    exit 0
fi

if [ ! -f "elbv2sanity.test" ]; then
    tester "elbv2" 'TestAccELBV2TargetGroup_basic'
    touch elbv2sanity.test
    exit 0
fi

if [ ! -f "kmssanity.test" ]; then
    tester "kms" 'TestAccKMSKey_basic'
    touch kmssanity.test
    exit 0
fi

if [ ! -f "lambdasanity.test" ]; then
    tester "lambda" 'TestAccLambdaFunction_basic|TestAccLambdaPermission_basic'
    touch lambdasanity.test
    exit 0
fi

if [ ! -f "metasanity.test" ]; then
    tester "meta" 'TestAccMetaRegionDataSource_basic|TestAccMetaRegionDataSource_endpoint|TestAccMetaPartitionDataSource_basic'
    touch metasanity.test
    exit 0
fi

if [ ! -f "route53sanity.test" ]; then
    tester "route53" 'TestAccRoute53Record_basic|TestAccRoute53Record_Latency_basic|TestAccRoute53ZoneDataSource_name'
    touch route53sanity.test
    exit 0
fi

if [ ! -f "s3sanity.test" ]; then
    tester "s3" 'TestAccS3Bucket_Basic_basic|TestAccS3Bucket_Security_corsUpdate|TestAccS3BucketPublicAccessBlock_basic|TestAccS3BucketPolicy_basic|TestAccS3BucketACL_updateACL'
    touch s3sanity.test
    exit 0
fi

if [ ! -f "secretsmanagersanity.test" ]; then
    tester "secretsmanager" 'TestAccSecretsManagerSecret_basic'
    touch secretsmanagersanity.test
    exit 0
fi

if [ ! -f "stssanity.test" ]; then
    tester "sts" 'TestAccSTSCallerIdentityDataSource_basic'
    touch stssanity.test
    exit 0
fi

echo "##teamcity[notification notifier='slack' message='*Sanity Tests Passed!*:white_check_mark:' sendTo='CN0G9S7M4' connectionId='PROJECT_EXT_8']\n"
