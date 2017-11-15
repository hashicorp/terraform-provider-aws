package aws

import (
	"testing"
)

func TestArn_iamRootUser(t *testing.T) {
	arn := iamArnString("aws", "1234567890", "root")
	expectedArn := "arn:aws:iam::1234567890:root"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_cloudFrontOriginAccessIdentity(t *testing.T) {
	arn := cloudFrontOriginAccessIdentityArnString("aws", "id1")
	expectedArn := "arn:aws:iam::cloudfront:user/CloudFront Origin Access Identity id1"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_dmsSubnetGroup(t *testing.T) {
	arn := dmsSubnetGroupArnString("aws", "us-east-1", "123456789012", "SG1")
	expectedArn := "arn:aws:dms:us-east-1:123456789012:subgrp:SG1"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_securityGroup(t *testing.T) {
	arn := securityGroupArnString("aws", "us-west-2", "1234567890", "SG1")
	expectedArn := "arn:aws:ec2:us-west-2:1234567890:security-group/SG1"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_elastiCacheCluster(t *testing.T) {
	arn := elastiCacheClusterArnString("aws", "us-east-2", "123456789012", "myCluster")
	expectedArn := "arn:aws:elasticache:us-east-2:123456789012:cluster:myCluster"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_rdsCluster(t *testing.T) {
	arn := rdsClusterArnString("aws", "us-east-1", "123456789012", "my-cluster1")
	expectedArn := "arn:aws:rds:us-east-1:123456789012:cluster:my-cluster1"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_rdsClusterParameterGroup(t *testing.T) {
	arn := rdsClusterParameterGroupArnString("aws", "us-east-1", "123456789012", "aurora-pg3")
	expectedArn := "arn:aws:rds:us-east-1:123456789012:cluster-pg:aurora-pg3"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_rdsDbInstance(t *testing.T) {
	arn := rdsDbInstanceArnString("aws", "us-east-1", "123456789012", "mysql-db-instance1")
	expectedArn := "arn:aws:rds:us-east-1:123456789012:db:mysql-db-instance1"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_rdsEventSubscription(t *testing.T) {
	arn := rdsEventSubscriptionArnString("aws", "us-east-1", "123456789012", "monitor-events2")
	expectedArn := "arn:aws:rds:us-east-1:123456789012:es:monitor-events2"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_rdsOptionGroup(t *testing.T) {
	arn := rdsOptionGroupArnString("aws", "us-east-1", "123456789012", "mysql-option-group1")
	expectedArn := "arn:aws:rds:us-east-1:123456789012:og:mysql-option-group1"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_rdsParameterGroup(t *testing.T) {
	arn := rdsParameterGroupArnString("aws", "us-east-1", "123456789012", "mysql-repl-pg1")
	expectedArn := "arn:aws:rds:us-east-1:123456789012:pg:mysql-repl-pg1"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_rdsSecurityGroup(t *testing.T) {
	arn := rdsSecurityGroupArnString("aws", "us-east-1", "123456789012", "dev-secgrp2")
	expectedArn := "arn:aws:rds:us-east-1:123456789012:secgrp:dev-secgrp2"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_rdsSubnetGroup(t *testing.T) {
	arn := rdsSubnetGroupArnString("aws", "us-east-1", "123456789012", "prod-subgrp1")
	expectedArn := "arn:aws:rds:us-east-1:123456789012:subgrp:prod-subgrp1"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_redshiftCluster(t *testing.T) {
	arn := redshiftClusterArnString("aws", "us-east-1", "123456789012", "my-cluster")
	expectedArn := "arn:aws:redshift:us-east-1:123456789012:cluster:my-cluster"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_redshiftSubnetGroup(t *testing.T) {
	arn := redshiftSubnetGroupArnString("aws", "us-east-1", "123456789012", "my-subnet-10")
	expectedArn := "arn:aws:redshift:us-east-1:123456789012:subnetgroup:my-subnet-10"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_s3(t *testing.T) {
	arn := s3ArnString("aws", "bucket1")
	expectedArn := "arn:aws:s3:::bucket1"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_sesDomainIdentity(t *testing.T) {
	arn := sesArnString("aws", "us-east-1", "123456789012", "identity/ID")
	expectedArn := "arn:aws:ses:us-east-1:123456789012:identity/ID"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_ssmDocument(t *testing.T) {
	arn := ssmDocumentArnString("aws", "us-east-1", "123456789012", "highAvailabilityServerSetup")
	expectedArn := "arn:aws:ssm:us-east-1:123456789012:document/highAvailabilityServerSetup"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_apiGatewayExecution(t *testing.T) {
	arn := apiGatewayExecutionArnString("aws", "us-east-1", "123456789012", "qsxrty/test/GET/mydemoresource/*")
	expectedArn := "arn:aws:execute-api:us-east-1:123456789012:qsxrty/test/GET/mydemoresource/*"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}

func TestArn_apiGatewayLambdaInvoke(t *testing.T) {
	arn := apiGatewayLambdaInvokeArnString("aws", "us-west-2", "arn:aws:lambda:us-west-2:999999999999:function:TEST")
	expectedArn := "arn:aws:apigateway:us-west-2:lambda:path/2015-03-31/functions/arn:aws:lambda:us-west-2:999999999999:function:TEST/invocations"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}
