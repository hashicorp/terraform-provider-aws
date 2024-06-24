// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneEngineVersionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"
	version := "1.1.0.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_basic(version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngine, "neptune"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, version),
					resource.TestCheckResourceAttrSet(dataSourceName, "engine_description"),
					resource.TestMatchResourceAttr(dataSourceName, "exportable_log_types.#", regexache.MustCompile(`^[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "parameter_group_family"),
					resource.TestMatchResourceAttr(dataSourceName, "supported_timezones.#", regexache.MustCompile(`^[0-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_log_exports_to_cloudwatch"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_read_replica"),
					resource.TestMatchResourceAttr(dataSourceName, "valid_upgrade_targets.#", regexache.MustCompile(`^[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "version_description"),
				),
			},
		},
	})
}

func TestAccNeptuneEngineVersionDataSource_preferred(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferred(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngine, "neptune"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "1.2.0.2"),
				),
			},
		},
	})
}

func TestAccNeptuneEngineVersionDataSource_defaultOnly(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_defaultOnly(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngine, "neptune"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrVersion),
				),
			},
		},
	})
}

func testAccEngineVersionPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

	input := &neptune.DescribeDBEngineVersionsInput{
		Engine:      aws.String("neptune"),
		DefaultOnly: aws.Bool(true),
	}

	_, err := conn.DescribeDBEngineVersionsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccEngineVersionDataSourceConfig_basic(version string) string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine  = "neptune"
  version = %[1]q
}
`, version)
}

func testAccEngineVersionDataSourceConfig_preferred() string {
	return `
data "aws_neptune_engine_version" "test" {
  preferred_versions = ["85.9.12", "1.2.0.2", "1.1.0.0"]
}
`
}

func testAccEngineVersionDataSourceConfig_defaultOnly() string {
	return `
data "aws_neptune_engine_version" "test" {}
`
}
