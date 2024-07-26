// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneOrderableDBInstanceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_orderable_db_instance.test"
	engine := "neptune"
	engineVersion := "1.1.0.0"
	licenseModel := "amazon-license"
	class := "db.t4g.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckOrderableDBInstance(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableDBInstanceDataSourceConfig_basic(class, engine, engineVersion, licenseModel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngine, engine),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngineVersion, engineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "license_model", licenseModel),
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", class),
				),
			},
		},
	})
}

func TestAccNeptuneOrderableDBInstanceDataSource_preferred(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_orderable_db_instance.test"
	engine := "neptune"
	engineVersion := "1.2.0.2"
	licenseModel := "amazon-license"
	preferredOption := "db.r5.2xlarge"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckOrderableDBInstance(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableDBInstanceDataSourceConfig_preferred(engine, engineVersion, licenseModel, preferredOption),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngine, engine),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngineVersion, engineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "license_model", licenseModel),
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", preferredOption),
				),
			},
		},
	})
}

func testAccPreCheckOrderableDBInstance(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

	input := &neptune.DescribeOrderableDBInstanceOptionsInput{
		Engine: aws.String("mysql"),
	}

	_, err := conn.DescribeOrderableDBInstanceOptionsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOrderableDBInstanceDataSourceConfig_basic(class, engine, version, license string) string {
	return fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  instance_class = %[1]q
  engine         = %[2]q
  engine_version = %[3]q
  license_model  = %[4]q
}
`, class, engine, version, license)
}

func testAccOrderableDBInstanceDataSourceConfig_preferred(engine, version, license, preferredOption string) string {
	return fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  engine         = %[1]q
  engine_version = %[2]q
  license_model  = %[3]q

  preferred_instance_classes = [
    "db.xyz.xlarge",
    %[4]q,
    "db.t3.small",
  ]
}
`, engine, version, license, preferredOption)
}
