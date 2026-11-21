// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs("data.aws_continuum_integration.by_name", tfjsonpath.New(names.AttrID), "aws_continuum_integration.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs("data.aws_continuum_environment.by_id", tfjsonpath.New(names.AttrName), "aws_continuum_environment.test", tfjsonpath.New(names.AttrName), compare.ValuesSame()),
					statecheck.ExpectKnownValue("data.aws_continuum_code_scans.test", tfjsonpath.New("environment_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.aws_continuum_code_issues.test", tfjsonpath.New("environment_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.aws_continuum_code_findings.test", tfjsonpath.New("environment_id"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_basic(rName, rName, "PT0H"), `
data "aws_continuum_integration" "by_name" {
  name = aws_continuum_integration.test.name
}

data "aws_continuum_environment" "by_id" {
  id = aws_continuum_environment.test.id
}

data "aws_continuum_code_scans" "test" {
  environment_id = aws_continuum_environment.test.id
}

data "aws_continuum_code_issues" "test" {
  environment_id = aws_continuum_environment.test.id
}

data "aws_continuum_code_findings" "test" {
  environment_id = aws_continuum_environment.test.id
}
`)
}
