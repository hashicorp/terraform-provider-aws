// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package meta_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestFindRegionByEC2Endpoint(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	var testCases = []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "does-not-exist",
			ErrCount: 1,
		},
		{
			Value:    "ec2.does-not-exist.amazonaws.com",
			ErrCount: 1,
		},
		{
			Value:    "us-east-1", // lintignore:AWSAT003
			ErrCount: 1,
		},
		{
			Value:    "ec2.us-east-1.amazonaws.com", // lintignore:AWSAT003
			ErrCount: 0,
		},
	}

	for _, tc := range testCases {
		_, err := tfmeta.FindRegionByEC2Endpoint(ctx, tc.Value)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Value, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Value)
		}
	}
}

func TestFindRegionByName(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	var testCases = []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "does-not-exist",
			ErrCount: 1,
		},
		{
			Value:    "ec2.us-east-1.amazonaws.com", // lintignore:AWSAT003
			ErrCount: 1,
		},
		{
			Value:    "us-east-1", // lintignore:AWSAT003
			ErrCount: 0,
		},
		{
			Value:    "ap-southeast-5", // lintignore:AWSAT003
			ErrCount: 0,
		},
	}

	for _, tc := range testCases {
		_, err := tfmeta.FindRegionByName(ctx, tc.Value)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Value, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Value)
		}
	}
}

func TestAccMetaRegionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_region.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrRegionalHostnameService(dataSourceName, names.AttrEndpoint, names.EC2),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrDescription), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrEndpoint), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
		},
	})
}

func TestAccMetaRegionDataSource_endpoint(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_region.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartition(t, endpoints.AwsPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig_endpoint(endpoints.EuWest1RegionID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrDescription), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrEndpoint), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(endpoints.EuWest1RegionID)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(endpoints.EuWest1RegionID)),
				},
			},
		},
	})
}

func TestAccMetaRegionDataSource_endpointAndName(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_region.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartition(t, endpoints.AwsPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig_endpointAndName(endpoints.ApNortheast1RegionID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrDescription), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrEndpoint), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(endpoints.ApNortheast1RegionID)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(endpoints.ApNortheast1RegionID)),
				},
			},
		},
	})
}

func TestAccMetaRegionDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_region.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig_name(endpoints.UsWest1RegionID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrDescription), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrEndpoint), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(endpoints.UsWest1RegionID)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(endpoints.UsWest1RegionID)),
				},
			},
		},
	})
}

func TestAccMetaRegionDataSource_endpointAndRegion(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_region.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartition(t, endpoints.AwsPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig_endpointAndRegion(endpoints.ApSoutheast2RegionID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrDescription), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrEndpoint), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(endpoints.ApSoutheast2RegionID)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(endpoints.ApSoutheast2RegionID)),
				},
			},
		},
	})
}

func TestAccMetaRegionDataSource_region(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_region.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig_region(endpoints.UsGovEast1RegionID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrDescription), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrEndpoint), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(endpoints.UsGovEast1RegionID)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(endpoints.UsGovEast1RegionID)),
				},
			},
		},
	})
}

const testAccRegionDataSourceConfig_empty = `
data "aws_region" "test" {}
`

func testAccRegionDataSourceConfig_endpoint(region string) string {
	return fmt.Sprintf(`
data "aws_partition" "test" {}

data "aws_region" "test" {
  endpoint = "ec2.%[1]s.${data.aws_partition.test.dns_suffix}"
}
`, region)
}

func testAccRegionDataSourceConfig_endpointAndName(region string) string {
	return fmt.Sprintf(`
data "aws_partition" "test" {}

data "aws_region" "test" {
  endpoint = "ec2.%[1]s.${data.aws_partition.test.dns_suffix}"
  name     = %[1]q
}
`, region)
}

func testAccRegionDataSourceConfig_name(region string) string {
	return fmt.Sprintf(`
data "aws_region" "test" {
  name = %[1]q
}
`, region)
}

func testAccRegionDataSourceConfig_endpointAndRegion(region string) string {
	return fmt.Sprintf(`
data "aws_partition" "test" {}

data "aws_region" "test" {
  endpoint = "ec2.%[1]s.${data.aws_partition.test.dns_suffix}"
  region   = %[1]q
}
`, region)
}

func testAccRegionDataSourceConfig_region(region string) string {
	return fmt.Sprintf(`
data "aws_region" "test" {
  region = %[1]q
}
`, region)
}
