// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneEnvironmentBlueprintDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var environmentblueprint datazone.GetEnvironmentBlueprintOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_datazone_environment_blueprint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintExists(ctx, dataSourceName, &environmentblueprint),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, "blueprint_provider"),
				),
			},
		},
	})
}

func testAccCheckEnvironmentBlueprintExists(ctx context.Context, name string, environmentblueprint *datazone.GetEnvironmentBlueprintOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.DSNameEnvironmentBlueprint, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.DSNameEnvironmentBlueprint, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := conn.GetEnvironmentBlueprint(ctx, &datazone.GetEnvironmentBlueprintInput{
			DomainIdentifier: aws.String(rs.Primary.Attributes["domain_id"]),
			Identifier:       aws.String(rs.Primary.Attributes[names.AttrID]),
		})

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.DSNameEnvironmentBlueprint, rs.Primary.ID, err)
		}

		*environmentblueprint = *resp

		return nil
	}
}

func testAccCheckEnvironmentBlueprintDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_environment_blueprint" {
				continue
			}

			_, err := conn.GetEnvironmentBlueprint(ctx, &datazone.GetEnvironmentBlueprintInput{
				DomainIdentifier: aws.String(rs.Primary.Attributes["domain_id"]),
				Identifier:       aws.String(rs.Primary.Attributes[names.AttrID]),
			})
			if tfdatazone.IsResourceMissing(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameEnvironmentBlueprintConfiguration, rs.Primary.ID, err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameEnvironmentBlueprintConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccEnvironmentBlueprintDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDomainConfig_basic(rName),
		`
data "aws_datazone_environment_blueprint" "test" {
  domain_id = aws_datazone_domain.test.id
  name      = "DefaultDataLake"
  managed   = true
}
`,
	)
}
