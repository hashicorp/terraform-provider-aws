// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudtrail "github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccControlTowerControlsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_controltower_controls.test"
	ouName := "Security"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccControlsDataSourceConfig_id(ouName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "enabled_controls.#"),
				),
			},
		},
	})
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	// Leverage the Control Tower created "aws-controltower-BaselineCloudTrail" to confirm Control Tower is deployed.
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailClient(ctx)
	_, err := tfcloudtrail.FindTrailInfoByName(ctx, conn, "aws-controltower-BaselineCloudTrail")

	if tfresource.NotFound(err) {
		t.Skip("skipping since Control Tower not found")
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccControlsDataSourceConfig_id(ouName string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

data "aws_organizations_organizational_units" "test" {
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

data "aws_controltower_controls" "test" {
  target_identifier = [
    for x in data.aws_organizations_organizational_units.test.children :
    x.arn if x.name == "%[1]s"
  ][0]
}
`, ouName)
}
