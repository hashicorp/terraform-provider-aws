// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package common

import "slices"

// coreServices is the list of AWS services included when the provider is built
// with the 'core' tag
//
// It includes primary services and their explicitly required dependencies.
var coreServices = []string{
	"account",
	"cloudwatch",
	"dynamodb",
	"ec2",
	"ecs",
	"eks",
	"elasticache",
	"iam",
	"kms",
	"lambda",
	"logs",
	"organizations",
	"rds",
	"s3",
	"s3control",
	"sns",
	"sqs",
	"sts",
}

// IsCoreService returns true if the given service package name is considered a
// core service
func IsCoreService(name string) bool {
	return slices.Contains(coreServices, name)
}
