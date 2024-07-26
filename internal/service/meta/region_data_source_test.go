// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestFindRegionByEC2Endpoint(t *testing.T) {
	t.Parallel()

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
		_, err := tfmeta.FindRegionByEndpoint(tc.Value)
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
	}

	for _, tc := range testCases {
		_, err := tfmeta.FindRegionByName(tc.Value)
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrDescription, regexache.MustCompile(`^.+$`)),
					acctest.CheckResourceAttrRegionalHostnameService(dataSourceName, names.AttrEndpoint, names.EC2),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, acctest.Region()),
				),
			},
		},
	})
}

func TestAccMetaRegionDataSource_endpoint(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_region.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig_endpoint(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrDescription, regexache.MustCompile(`^.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrEndpoint, regexache.MustCompile(fmt.Sprintf("^%s\\.[^.]+\\.%s$", names.EC2, acctest.PartitionDNSSuffix()))),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrName, regexache.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

func TestAccMetaRegionDataSource_endpointAndName(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_region.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig_endpointAndName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrDescription, regexache.MustCompile(`^.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrEndpoint, regexache.MustCompile(fmt.Sprintf("^ec2\\.[^.]+\\.%s$", acctest.PartitionDNSSuffix()))),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrName, regexache.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

func TestAccMetaRegionDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_region.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig_name(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrDescription, regexache.MustCompile(`^.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrEndpoint, regexache.MustCompile(fmt.Sprintf("^ec2\\.[^.]+\\.%s$", acctest.PartitionDNSSuffix()))),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrName, regexache.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

const testAccRegionDataSourceConfig_empty = `
data "aws_region" "test" {}
`

func testAccRegionDataSourceConfig_endpoint() string {
	return `
data "aws_partition" "test" {}

data "aws_regions" "test" {}

data "aws_region" "test" {
  endpoint = "ec2.${tolist(data.aws_regions.test.names)[0]}.${data.aws_partition.test.dns_suffix}"
}
`
}

func testAccRegionDataSourceConfig_endpointAndName() string {
	return `
data "aws_partition" "test" {}

data "aws_regions" "test" {}

data "aws_region" "test" {
  endpoint = "ec2.${tolist(data.aws_regions.test.names)[0]}.${data.aws_partition.test.dns_suffix}"
  name     = tolist(data.aws_regions.test.names)[0]
}
`
}

func testAccRegionDataSourceConfig_name() string {
	return `
data "aws_regions" "test" {}

data "aws_region" "test" {
  name = tolist(data.aws_regions.test.names)[0]
}
`
}
