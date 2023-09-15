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

if [ ! -f "iamsanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/iam/... \
        -v -parallel 10 -run='TestAccIAMRole_basic|TestAccIAMRole_namePrefix|TestAccIAMRole_disappears|TestAccIAMRole_InlinePolicy_basic|TestAccIAMPolicyDocumentDataSource_basic|TestAccIAMPolicyDocumentDataSource_sourceConflicting|TestAccIAMPolicyDocumentDataSource_sourceJSONValidJSON|TestAccIAMRolePolicyAttachment_basic|TestAccIAMRolePolicyAttachment_disappears|TestAccIAMRolePolicyAttachment_Disappears_role|TestAccIAMPolicy_basic|TestAccIAMPolicy_policy|TestAccIAMPolicy_tags|TestAccIAMRolePolicy_basic|TestAccIAMRolePolicy_unknownsInPolicy|TestAccIAMInstanceProfile_basic|TestAccIAMInstanceProfile_tags' -timeout 60m
    touch iamsanity.test
    exit 0
fi

if [ ! -f "logssanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/logs/... \
        -v -parallel 10 -run='TestAccLogsGroup_basic|TestAccLogsGroup_multiple' -timeout 60m
    touch logssanity.test
    exit 0
fi

if [ ! -f "ec2sanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/ec2/... \
        -v -parallel 10 -run='TestAccVPCSecurityGroup_basic|TestAccVPCSecurityGroup_ipRangesWithSameRules|TestAccVPCSecurityGroup_vpcAllEgress|TestAccVPCSecurityGroupRule_race|TestAccVPCSecurityGroupRule_protocolChange|TestAccVPCDataSource_basic|TestAccVPCSubnet_basic|TestAccVPC_tenancy|TestAccVPCRouteTableAssociation_Subnet_basic|TestAccVPCRouteTable_basic' -timeout 60m
    touch ec2sanity.test
    exit 0
fi

if [ ! -f "ecssanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/ecs/... \
        -v -parallel 10 -run='TestAccECSTaskDefinition_basic|TestAccECSService_basic' -timeout 60m
    touch ecssanity.test
    exit 0
fi

if [ ! -f "elbv2sanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/elbv2/... \
        -v -parallel 10 -run='TestAccELBV2TargetGroup_basic' -timeout 60m
    touch elbv2sanity.test
    exit 0
fi

if [ ! -f "kmssanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/kms/... \
        -v -parallel 10 -run='TestAccKMSKey_basic' -timeout 60m
    touch kmssanity.test
    exit 0
fi

if [ ! -f "lambdasanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/lambda/... \
        -v -parallel 10 -run='TestAccLambdaFunction_basic|TestAccLambdaPermission_basic' -timeout 60m
    touch lambdasanity.test
    exit 0
fi

if [ ! -f "metasanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/meta/... \
        -v -parallel 10 -run='TestAccMetaRegionDataSource_basic|TestAccMetaRegionDataSource_endpoint|TestAccMetaPartitionDataSource_basic' -timeout 60m
    touch metasanity.test
    exit 0
fi

if [ ! -f "route53sanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/route53/... \
        -v -parallel 10 -run='TestAccRoute53Record_basic|TestAccRoute53Record_Latency_basic|TestAccRoute53ZoneDataSource_name' -timeout 60m
    touch route53sanity.test
    exit 0
fi

if [ ! -f "s3sanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/s3/... \
        -v -parallel 10 -run='TestAccS3Bucket_Basic_basic|TestAccS3Bucket_Security_corsUpdate|TestAccS3BucketPublicAccessBlock_basic|TestAccS3BucketPolicy_basic|TestAccS3BucketACL_updateACL' -timeout 60m
    touch s3sanity.test
    exit 0
fi

if [ ! -f "secretsmanagersanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/secretsmanager/... \
        -v -parallel 10 -run='TestAccSecretsManagerSecret_basic' -timeout 60m
    touch secretsmanagersanity.test
    exit 0
fi

if [ ! -f "stssanity.test" ]; then
    TF_ACC=1 go test \
        ./internal/service/sts/... \
        -v -parallel 10 -run='TestAccSTSCallerIdentityDataSource_basic' -timeout 60m
    touch stssanity.test
    exit 0
fi

oneliner=$( sort -R oneliners.txt | head -1 )
echo "##teamcity[notification notifier='slack' message='**Sanity Tests Passed!**\n${oneliner}' sendTo='CN0G9S7M4' connectionId='PROJECT_EXT_8']\n"
