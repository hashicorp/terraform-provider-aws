// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import "strings"

const (
	// Please make sure GovCloud and commercial support these since they vary
	db2PreferredInstanceClasses             = `"db.t3.small", "db.r6i.large", "db.m6i.large"`
	mariaDBPreferredInstanceClasses         = `"db.t3.micro", "db.t3.small", "db.t2.small", "db.t2.medium"`
	mySQLPreferredInstanceClasses           = `"db.t3.micro", "db.t3.small", "db.t2.small", "db.t2.medium"`
	oraclePreferredInstanceClasses          = `"db.t3.medium", "db.t2.medium", "db.t3.large", "db.t2.large"` // Oracle requires at least a medium instance as a replica source
	oracleSE2PreferredInstanceClasses       = `"db.m5.large", "db.m4.large", "db.r4.large"`
	outpostPreferredInstanceClasses         = `"db.m5.large", "db.r5.large"` // https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-on-outposts.db-instance-classes.html
	postgresPreferredInstanceClasses        = `"db.t3.micro", "db.t3.small", "db.t2.small", "db.t2.medium"`
	sqlServerCustomPreferredInstanceClasses = `"db.m5.large", "db.m5.xlarge"`
	sqlServerPreferredInstanceClasses       = `"db.t2.small", "db.t3.small"`
	sqlServerSEPreferredInstanceClasses     = `"db.m5.large", "db.m4.large", "db.r4.large"`
)

var (
	// Prices for mysql as of 2024-02-02 in us-west-2 which are representative of
	// relative prices in other regions. Prices are per hour.
	instanceClassesSlice = []string{
		/* 0.016 */ `"db.t4g.micro"`,
		/* 0.017 */ `"db.t3.micro"`,
		/* 0.032 */ `"db.t4g.small"`,
		/* 0.034 */ `"db.t3.small"`,
		/* 0.065 */ `"db.t4g.medium"`,
		/* 0.068 */ `"db.t3.medium"`,
		/* 0.129 */ `"db.t4g.large"`,
		/* 0.136 */ `"db.t3.large"`,
		/* 0.152 */ `"db.m6g.large"`,
		/* 0.168 */ `"db.m7g.large"`,
		/* 0.171 */ `"db.m5.large"`,
		/* 0.171 */ `"db.m6i.large"`,
		/* 0.187 */ `"db.m6gd.large"`,
		/* 0.210 */ `"db.m5d.large"`,
		/* 0.215 */ `"db.r6g.large"`,
		/* 0.224 */ `"db.m6id.large"`,
		/* 0.239 */ `"db.r7g.large"`,
		/* 0.240 */ `"db.r5.large"`,
		/* 0.240 */ `"db.r6i.large"`,
		/* 0.257 */ `"db.r6gd.large"`,
		/* 0.258 */ `"db.m6in.large"`,
		/* 0.258 */ `"db.t4g.xlarge"`,
		/* 0.272 */ `"db.t3.xlarge"`,
		/* 0.286 */ `"db.r5d.large"`,
		/* 0.295 */ `"db.m6idn.large"`,
		/* 0.296 */ `"db.r5b.large"`,
		/* 0.298 */ `"db.r6id.large"`,
		/* 0.304 */ `"db.m6g.xlarge"`,
		/* 0.326 */ `"db.x2g.large"`,
		/* 0.337 */ `"db.m7g.xlarge"`,
		/* 0.342 */ `"db.m5.xlarge"`,
		/* 0.342 */ `"db.m6i.xlarge"`,
		/* 0.346 */ `"db.r6in.large"`,
		/* 0.373 */ `"db.m6gd.xlarge"`,
		/* 0.388 */ `"db.r6idn.large"`,
		/* 0.419 */ `"db.m5d.xlarge"`,
		/* 0.430 */ `"db.r6g.xlarge"`,
		/* 0.448 */ `"db.m6id.xlarge"`,
		/* 0.478 */ `"db.r7g.xlarge"`,
		/* 0.480 */ `"db.r5.xlarge"`,
		/* 0.480 */ `"db.r6i.xlarge"`,
		/* 0.514 */ `"db.r6gd.xlarge"`,
		/* 0.516 */ `"db.m6in.xlarge"`,
		/* 0.517 */ `"db.t4g.2xlarge"`,
		/* 0.544 */ `"db.t3.2xlarge"`,
		/* 0.571 */ `"db.r5d.xlarge"`,
		/* 0.590 */ `"db.m6idn.xlarge"`,
		/* 0.592 */ `"db.r5b.xlarge"`,
		/* 0.596 */ `"db.r6id.xlarge"`,
	}

	// These instance classes will be selected in order. Use sufficient criteria
	// with aws_rds_engine_version and aws_rds_orderable_db_instance to ensure
	// one is selected with the features you need.
	// Prices for mysql as of 2024-02-02 in us-west-2 which are representative of
	// relative prices in other regions. Prices are per hour.
	mainInstanceClasses = strings.Join(instanceClassesSlice, ", ")

	// Unused but leaving for reference in case some future test needs one of them.
	expensiveClassesSlice = []string{
		/* 0.608 */ `"db.m6g.2xlarge"`,
		/* 0.652 */ `"db.x2g.xlarge"`,
		/* 0.674 */ `"db.m7g.2xlarge"`,
		/* 0.684 */ `"db.m5.2xlarge"`,
		/* 0.684 */ `"db.m6i.2xlarge"`,
		/* 0.692 */ `"db.r6in.xlarge"`,
		/* 0.747 */ `"db.m6gd.2xlarge"`,
		/* 0.775 */ `"db.r6idn.xlarge"`,
		/* 0.838 */ `"db.m5d.2xlarge"`,
		/* 0.859 */ `"db.r6g.2xlarge"`,
		/* 0.897 */ `"db.m6id.2xlarge"`,
		/* 0.956 */ `"db.r7g.2xlarge"`,
		/* 0.960 */ `"db.r5.2xlarge"`,
		/* 0.960 */ `"db.r6i.2xlarge"`,
		/* 1.029 */ `"db.r6gd.2xlarge"`,
		/* 1.033 */ `"db.m6in.2xlarge"`,
		/* 1.143 */ `"db.r5d.2xlarge"`,
		/* 1.180 */ `"db.m6idn.2xlarge"`,
		/* 1.184 */ `"db.r5b.2xlarge"`,
		/* 1.193 */ `"db.r6id.2xlarge"`,
		/* 1.216 */ `"db.m6g.4xlarge"`,
		/* 1.304 */ `"db.x2g.2xlarge"`,
		/* 1.348 */ `"db.m7g.4xlarge"`,
		/* 1.368 */ `"db.m5.4xlarge"`,
		/* 1.368 */ `"db.m6i.4xlarge"`,
		/* 1.384 */ `"db.r6in.2xlarge"`,
		/* 1.493 */ `"db.m6gd.4xlarge"`,
		/* 1.551 */ `"db.r6idn.2xlarge"`,
		/* 1.676 */ `"db.m5d.4xlarge"`,
		/* 1.718 */ `"db.r6g.4xlarge"`,
		/* 1.794 */ `"db.m6id.4xlarge"`,
		/* 1.913 */ `"db.r7g.4xlarge"`,
		/* 1.920 */ `"db.r5.4xlarge"`,
		/* 1.920 */ `"db.r6i.4xlarge"`,
		/* 2.057 */ `"db.r6gd.4xlarge"`,
		/* 2.066 */ `"db.m6in.4xlarge"`,
		/* 2.221 */ `"db.x2iedn.xlarge"`,
		/* 2.286 */ `"db.r5d.4xlarge"`,
		/* 2.360 */ `"db.m6idn.4xlarge"`,
		/* 2.368 */ `"db.r5b.4xlarge"`,
		/* 2.386 */ `"db.r6id.4xlarge"`,
		/* 2.432 */ `"db.m6g.8xlarge"`,
		/* 2.608 */ `"db.x2g.4xlarge"`,
		/* 2.696 */ `"db.m7g.8xlarge"`,
		/* 2.736 */ `"db.m6i.8xlarge"`,
		/* 2.740 */ `"db.m5.8xlarge"`,
		/* 2.767 */ `"db.r6in.4xlarge"`,
		/* 2.987 */ `"db.m6gd.8xlarge"`,
		/* 3.101 */ `"db.r6idn.4xlarge"`,
		/* 3.352 */ `"db.m5d.8xlarge"`,
		/* 3.437 */ `"db.r6g.8xlarge"`,
		/* 3.587 */ `"db.m6id.8xlarge"`,
		/* 3.648 */ `"db.m6g.12xlarge"`,
		/* 3.825 */ `"db.r7g.8xlarge"`,
		/* 3.840 */ `"db.r5.8xlarge"`,
		/* 3.840 */ `"db.r6i.8xlarge"`,
		/* 4.044 */ `"db.m7g.12xlarge"`,
		/* 4.104 */ `"db.m5.12xlarge"`,
		/* 4.104 */ `"db.m6i.12xlarge"`,
		/* 4.114 */ `"db.r6gd.8xlarge"`,
		/* 4.131 */ `"db.m6in.8xlarge"`,
		/* 4.442 */ `"db.x2iedn.2xlarge"`,
		/* 4.480 */ `"db.m6gd.12xlarge"`,
		/* 4.571 */ `"db.r5d.8xlarge"`,
		/* 4.720 */ `"db.m6idn.8xlarge"`,
		/* 4.736 */ `"db.r5b.8xlarge"`,
		/* 4.771 */ `"db.r6id.8xlarge"`,
		/* 4.864 */ `"db.m6g.16xlarge"`,
		/* 5.028 */ `"db.m5d.12xlarge"`,
		/* 5.155 */ `"db.r6g.12xlarge"`,
		/* 5.216 */ `"db.x2g.8xlarge"`,
		/* 5.381 */ `"db.m6id.12xlarge"`,
		/* 5.392 */ `"db.m7g.16xlarge"`,
		/* 5.470 */ `"db.m5.16xlarge"`,
		/* 5.472 */ `"db.m6i.16xlarge"`,
		/* 5.534 */ `"db.r6in.8xlarge"`,
		/* 5.738 */ `"db.r7g.12xlarge"`,
		/* 5.760 */ `"db.r5.12xlarge"`,
		/* 5.760 */ `"db.r6i.12xlarge"`,
		/* 5.973 */ `"db.m6gd.16xlarge"`,
		/* 6.171 */ `"db.r6gd.12xlarge"`,
		/* 6.197 */ `"db.m6in.12xlarge"`,
		/* 6.203 */ `"db.r6idn.8xlarge"`,
		/* 6.705 */ `"db.m5d.16xlarge"`,
		/* 6.857 */ `"db.r5d.12xlarge"`,
		/* 6.874 */ `"db.r6g.16xlarge"`,
		/* 7.080 */ `"db.m6idn.12xlarge"`,
		/* 7.104 */ `"db.r5b.12xlarge"`,
		/* 7.157 */ `"db.r6id.12xlarge"`,
		/* 7.174 */ `"db.m6id.16xlarge"`,
		/* 7.650 */ `"db.r7g.16xlarge"`,
		/* 7.680 */ `"db.r5.16xlarge"`,
		/* 7.680 */ `"db.r6i.16xlarge"`,
		/* 7.824 */ `"db.x2g.12xlarge"`,
		/* 8.208 */ `"db.m5.24xlarge"`,
		/* 8.208 */ `"db.m6i.24xlarge"`,
		/* 8.229 */ `"db.r6gd.16xlarge"`,
		/* 8.262 */ `"db.m6in.16xlarge"`,
		/* 8.302 */ `"db.r6in.12xlarge"`,
		/* 8.883 */ `"db.x2iedn.4xlarge"`,
		/* 9.143 */ `"db.r5d.16xlarge"`,
		/* 9.304 */ `"db.r6idn.12xlarge"`,
		/* 9.440 */ `"db.m6idn.16xlarge"`,
		/* 9.472 */ `"db.r5b.16xlarge"`,
		/* 9.543 */ `"db.r6id.16xlarge"`,
	}
)
