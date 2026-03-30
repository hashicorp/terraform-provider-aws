// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import "testing"

func TestIsServicePrincipal(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		principal string
		valid     bool
	}{
		{"ec2.amazonaws.com", true},
		{"s3.amazonaws.com", true},
		{"elasticmapreduce.amazonaws.com", true},
		{"lambda.amazonaws.com", true},
		{"ecs-tasks.amazonaws.com", true},
		{"states.amazonaws.com", true},
		{"logs.amazonaws.com", true},
		{"delivery.logs.amazonaws.com", true},
		{"cognito-idp.amazonaws.com", true},
		{"config.amazonaws.com", true},
		{"ssm.amazonaws.com", true},
		{"sns.amazonaws.com", true},
		{"sqs.amazonaws.com", true},
		{"apigateway.amazonaws.com", true},
		{"cloudformation.amazonaws.com", true},
		{"codebuild.amazonaws.com", true},
		{"codepipeline.amazonaws.com", true},
		{"dynamodb.amazonaws.com", true},
		{"elasticloadbalancing.amazonaws.com", true},
		{"events.amazonaws.com", true},
		{"firehose.amazonaws.com", true},
		{"kinesis.amazonaws.com", true},
		{"kms.amazonaws.com", true},
		{"rds.amazonaws.com", true},
		{"redshift.amazonaws.com", true},
		{"route53.amazonaws.com", true},
		{"secretsmanager.amazonaws.com", true},
		{"sts.amazonaws.com", true},
		{"ec2.amazon.com", true},
		{"", false},
		{"invalid", false},
		{"amazonaws.com", false},
		{"ec2", false},
		{".amazonaws.com", false},
		{"ec2.amazonaws", false},
		{"ec2.amazonaws.org", false},
		{"ec2.example.com", false},
		{"arn:aws:iam::123456789012:role/test", false},
		{"123456789012", false},
	} {
		ok := IsServicePrincipal(tc.principal)
		if got, want := ok, tc.valid; got != want {
			t.Errorf("IsServicePrincipal(%q) = %v, want %v", tc.principal, got, want)
		}
	}
}
