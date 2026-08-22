// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/interconnect"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// envVarRegion optionally overrides the Region used by the Interconnect acceptance tests.
// Interconnect Environments are Region-scoped catalog entries that cannot be created, so
// the tests must run in a Region where Environments are already available. For example:
//
//	INTERCONNECT_REGION=us-east-1 make testacc PKG=interconnect
const envVarRegion = "INTERCONNECT_REGION"

func TestAccInterconnectEnvironmentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_interconnect_environment.test"
	environmentsDataSourceName := "data.aws_interconnect_environments.test"
	region := testAccRegion()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckEnvironments(ctx, t, region) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDataSourceConfig_basic(region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "environment_id", environmentsDataSourceName, "environments.0.environment_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrLocation, environmentsDataSourceName, "environments.0.location"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, environmentsDataSourceName, "environments.0.state"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, environmentsDataSourceName, "environments.0.type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "interconnect_provider", environmentsDataSourceName, "environments.0.interconnect_provider"),
					resource.TestCheckResourceAttrSet(dataSourceName, "bandwidths.#"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, region),
				),
			},
		},
	})
}

// testAccRegion returns the Region set in the INTERCONNECT_REGION environment variable,
// falling back to the default acceptance testing Region when it is not set.
func testAccRegion() string {
	if region := os.Getenv(envVarRegion); region != "" {
		return region
	}

	return acctest.Region()
}

// testAccPreCheckEnvironments skips the test unless the given Region has at least one
// Interconnect Environment. Environments are AWS-provided catalog entries that cannot be
// created, so a test that looks one up has nothing to fall back on when none exist.
func testAccPreCheckEnvironments(ctx context.Context, t *testing.T, region string) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).InterconnectClient(ctx)

	input := interconnect.ListEnvironmentsInput{}
	output, err := conn.ListEnvironments(ctx, &input, func(o *interconnect.Options) {
		o.Region = region
	})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if len(output.Environments) == 0 {
		t.Skipf("skipping acceptance testing: no Interconnect Environments in %s, set %s to a Region that has them", region, envVarRegion)
	}
}

func testAccEnvironmentDataSourceConfig_basic(region string) string {
	return fmt.Sprintf(`
data "aws_interconnect_environments" "test" {
  region = %[1]q
}

data "aws_interconnect_environment" "test" {
  region         = %[1]q
  environment_id = data.aws_interconnect_environments.test.environments[0].environment_id
}
`, region)
}
