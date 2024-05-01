// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import "strings"

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
)
