// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
		ErrorCheck:               acctest.ErrorCheck(t, controltower.EndpointsID),
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
	// leverage the control tower created "aws-controltower-BaselineCloudTrail" to confirm control tower is deployed
	var trails []string
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailConn(ctx)

	input := &cloudtrail.ListTrailsInput{}
	err := conn.ListTrailsPagesWithContext(ctx, input, func(page *cloudtrail.ListTrailsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, trail := range page.Trails {
			if trail == nil {
				continue
			}
			trails = append(trails, *trail.Name)
		}

		return !lastPage
	})

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	// Ensure there is a Control Tower trail
	ctTrail := false
	for _, t := range trails {
		if t == "aws-controltower-BaselineCloudTrail" {
			ctTrail = true
		}
	}
	if !ctTrail {
		t.Skip("skipping since Control Tower not found")
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
