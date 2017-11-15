package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/arn"
)

func arnString(partition, region, service, accountId, resource string) string {
	return arn.ARN{
		Partition: partition,
		Region:    region,
		// See http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#genref-aws-service-namespaces
		Service:   service,
		AccountID: accountId,
		Resource:  resource,
	}.String()
}

// See http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-iam
func iamArnString(partition, accountId, resource string) string {
	return arnString(
		partition,
		"",
		"iam",
		accountId,
		resource)
}

func cloudFrontOriginAccessIdentityArnString(partition, identity string) string {
	return iamArnString(
		partition,
		"cloudfront",
		"user/CloudFront Origin Access Identity "+identity)
}

// See http://docs.aws.amazon.com/dms/latest/userguide/CHAP_Introduction.ARN.html
func dmsArnString(partition, region, accountId, resourceType, resourceName string) string {
	return arnString(
		partition,
		region,
		"dms",
		accountId,
		resourceType+":"+resourceName)
}

func dmsSubnetGroupArnString(partition, region, accountId, sngId string) string {
	return dmsArnString(
		partition,
		region,
		accountId,
		"subgrp",
		sngId)
}

// See http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-ec2
func ec2ArnString(partition, region, accountId, resource string) string {
	return arnString(
		partition,
		region,
		"ec2",
		accountId,
		resource)
}

func securityGroupArnString(partition, region, accountId, sgId string) string {
	return ec2ArnString(
		partition,
		region,
		accountId,
		"security-group/"+sgId)
}

// See http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-elasticache
func elastiCacheArnString(partition, region, accountId, resourceType, resourceName string) string {
	return arnString(
		partition,
		region,
		"elasticache",
		accountId,
		resourceType+":"+resourceName)
}

func elastiCacheClusterArnString(partition, region, accountId, clusterId string) string {
	return elastiCacheArnString(
		partition,
		region,
		accountId,
		"cluster",
		clusterId)
}

// See http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-rds
func rdsArnString(partition, region, accountId, resourceType, resourceName string) string {
	return arnString(
		partition,
		region,
		"rds",
		accountId,
		resourceType+":"+resourceName)
}

func rdsClusterArnString(partition, region, accountId, clusterId string) string {
	return rdsArnString(
		partition,
		region,
		accountId,
		"cluster",
		clusterId)
}

func rdsClusterParameterGroupArnString(partition, region, accountId, pgId string) string {
	return rdsArnString(
		partition,
		region,
		accountId,
		"cluster-pg",
		pgId)
}

func rdsDbInstanceArnString(partition, region, accountId, dbId string) string {
	return rdsArnString(
		partition,
		region,
		accountId,
		"db",
		dbId)
}

func rdsEventSubscriptionArnString(partition, region, accountId, esId string) string {
	return rdsArnString(
		partition,
		region,
		accountId,
		"es",
		esId)
}

func rdsOptionGroupArnString(partition, region, accountId, ogId string) string {
	return rdsArnString(
		partition,
		region,
		accountId,
		"og",
		ogId)
}

func rdsParameterGroupArnString(partition, region, accountId, pgId string) string {
	return rdsArnString(
		partition,
		region,
		accountId,
		"pg",
		pgId)
}

func rdsSecurityGroupArnString(partition, region, accountId, sgId string) string {
	return rdsArnString(
		partition,
		region,
		accountId,
		"secgrp",
		sgId)
}

func rdsSubnetGroupArnString(partition, region, accountId, sngId string) string {
	return rdsArnString(
		partition,
		region,
		accountId,
		"subgrp",
		sngId)
}

func buildRdsArnString(id, partition, accountId, region string, f func(string, string, string, string) string) (string, error) {
	if partition == "" {
		return "", fmt.Errorf("Unable to construct RDS ARN because of missing AWS partition")
	}
	if accountId == "" {
		return "", fmt.Errorf("Unable to construct RDS ARN because of missing AWS Account ID")
	}
	return f(partition, region, accountId, id), nil
}

// See http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-redshift
func redshiftArnString(partition, region, accountId, resourceType, resourceName string) string {
	return arnString(
		partition,
		region,
		"redshift",
		accountId,
		resourceType+":"+resourceName)
}

func redshiftClusterArnString(partition, region, accountId, clusterId string) string {
	return redshiftArnString(
		partition,
		region,
		accountId,
		"cluster",
		clusterId)
}

func redshiftSubnetGroupArnString(partition, region, accountId, sngId string) string {
	return redshiftArnString(
		partition,
		region,
		accountId,
		"subnetgroup",
		sngId)
}

func buildRedshiftArnString(id, partition, accountId, region string, f func(string, string, string, string) string) (string, error) {
	if partition == "" {
		return "", fmt.Errorf("Unable to construct Redshift ARN because of missing AWS partition")
	}
	if accountId == "" {
		return "", fmt.Errorf("Unable to construct Redshift ARN because of missing AWS Account ID")
	}
	return f(partition, region, accountId, id), nil
}

// See http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-s3
func s3ArnString(partition, resource string) string {
	return arnString(
		partition,
		"",
		"s3",
		"",
		resource)
}

func sesArnString(partition, region, accountId, resource string) string {
	return arnString(
		partition,
		region,
		"ses",
		accountId,
		resource)
}

// See http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-ssm
func ssmArnString(partition, region, accountId, resourceType, resourceName string) string {
	return arnString(
		partition,
		region,
		"ssm",
		accountId,
		resourceType+"/"+resourceName)
}

func ssmDocumentArnString(partition, region, accountId, documentId string) string {
	return ssmArnString(
		partition,
		region,
		accountId,
		"document",
		documentId)
}

// See http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-apigateway
func apiGatewayExecutionArnString(partition, region, accountId, id string) string {
	return arnString(
		partition,
		region,
		"execute-api",
		accountId,
		id)
}

func apiGatewayLambdaInvokeArnString(partition, region, lambdaArn string) string {
	apiVersion := "2015-03-31"
	return arnString(
		partition,
		region,
		"apigateway",
		"lambda",
		fmt.Sprintf("path/%s/functions/%s/invocations", apiVersion, lambdaArn))
}
