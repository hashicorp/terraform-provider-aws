// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

const (
	// Please make sure GovCloud and commercial support these since they vary
	postgresPreferredInstanceClasses        = `"db.t3.micro", "db.t3.small", "db.t2.small", "db.t2.medium"`
	mySQLPreferredInstanceClasses           = `"db.t3.micro", "db.t3.small", "db.t2.small", "db.t2.medium"`
	mariaDBPreferredInstanceClasses         = `"db.t3.micro", "db.t3.small", "db.t2.small", "db.t2.medium"`
	oraclePreferredInstanceClasses          = `"db.t3.medium", "db.t2.medium", "db.t3.large", "db.t2.large"` // Oracle requires at least a medium instance as a replica source
	sqlServerCustomPreferredInstanceClasses = `"db.m5.large", "db.m5.xlarge"`
	sqlServerPreferredInstanceClasses       = `"db.t2.small", "db.t3.small"`
	sqlServerSEPreferredInstanceClasses     = `"db.m5.large", "db.m4.large", "db.r4.large"`
	oracleSE2PreferredInstanceClasses       = `"db.m5.large", "db.m4.large", "db.r4.large"`
	outpostPreferredInstanceClasses         = `"db.m5.large", "db.r5.large"` // https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-on-outposts.db-instance-classes.html
)
