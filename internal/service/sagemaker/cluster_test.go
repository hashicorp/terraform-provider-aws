// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSageMakerCluster_basic(t *testing.T) {
}

func TestAccSageMakerCluster_disappears(t *testing.T) {
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return nil
}

func testAccCheckClusterExists(ctx context.Context, name string, cluster *sagemaker.DescribeClusterOutput) resource.TestCheckFunc {
	return nil
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
}

func testAccCheckClusterNotRecreated(before, after *sagemaker.DescribeClusterOutput) resource.TestCheckFunc {
	return nil
}

func testAccClusterConfig_basic(rName, version string) string {
	return ""
}
