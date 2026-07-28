// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"testing"

	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// The Secrets Manager integration is account-wide. Do not run this test in
// parallel with other tests that manage the same integration.
func TestAccODBAutonomousDatabaseSecretsManagerIntegration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var role odbtypes.OciIamRole
	resourceName := "aws_odb_autonomous_database_secrets_manager_integration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccAutonomousDatabaseSecretsManagerIntegrationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutonomousDatabaseSecretsManagerIntegrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutonomousDatabaseSecretsManagerIntegrationConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseSecretsManagerIntegrationExists(ctx, t, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(odbtypes.OciIamRoleStatusAvailable)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAutonomousDatabaseSecretsManagerIntegrationPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)

	_, err := tfodb.FindAutonomousDatabaseSecretsManagerIntegration(ctx, conn)
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	t.Skip("skipping acceptance testing: AWS Secrets Manager integration already exists; this test destroys the account-level integration")
}

func testAccCheckAutonomousDatabaseSecretsManagerIntegrationExists(ctx context.Context, t *testing.T, name string, role *odbtypes.OciIamRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAutonomousDatabaseSecretsManagerIntegration, name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAutonomousDatabaseSecretsManagerIntegration, name, errors.New("ID not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
		found, err := tfodb.FindAutonomousDatabaseSecretsManagerIntegration(ctx, conn)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAutonomousDatabaseSecretsManagerIntegration, rs.Primary.ID, err)
		}

		*role = *found
		return nil
	}
}

func testAccCheckAutonomousDatabaseSecretsManagerIntegrationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_autonomous_database_secrets_manager_integration" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
			_, err := tfodb.FindAutonomousDatabaseSecretsManagerIntegration(ctx, conn)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameAutonomousDatabaseSecretsManagerIntegration, rs.Primary.ID, err)
			}
			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameAutonomousDatabaseSecretsManagerIntegration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccAutonomousDatabaseSecretsManagerIntegrationConfigBasic() string {
	return `
resource "aws_odb_autonomous_database_secrets_manager_integration" "test" {}
`
}
