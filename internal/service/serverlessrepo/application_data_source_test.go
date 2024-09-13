// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serverlessrepo_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServerlessRepoApplicationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_serverlessapplicationrepository_application.secrets_manager_postgres_single_user_rotator"
	appARN := testAccCloudFormationApplicationID()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServerlessRepoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfig_basic(appARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationIDDataSource(datasourceName),
					resource.TestCheckResourceAttr(datasourceName, names.AttrName, "SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttrSet(datasourceName, "semantic_version"),
					resource.TestCheckResourceAttrSet(datasourceName, "source_code_url"),
					resource.TestCheckResourceAttrSet(datasourceName, "template_url"),
					resource.TestCheckResourceAttrSet(datasourceName, "required_capabilities.#"),
				),
			},
		},
	})
}

func TestAccServerlessRepoApplicationDataSource_versioned(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_serverlessapplicationrepository_application.secrets_manager_postgres_single_user_rotator"
	appARN := testAccCloudFormationApplicationID()

	const (
		version1 = "1.0.13"
		version2 = "1.1.36"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServerlessRepoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfig_versioned(appARN, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationIDDataSource(datasourceName),
					resource.TestCheckResourceAttr(datasourceName, names.AttrName, "SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr(datasourceName, "semantic_version", version1),
					resource.TestCheckResourceAttrSet(datasourceName, "source_code_url"),
					resource.TestCheckResourceAttrSet(datasourceName, "template_url"),
					resource.TestCheckResourceAttr(datasourceName, "required_capabilities.#", acctest.Ct0),
				),
			},
			{
				Config: testAccApplicationDataSourceConfig_versioned(appARN, version2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationIDDataSource(datasourceName),
					resource.TestCheckResourceAttr(datasourceName, names.AttrName, "SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr(datasourceName, "semantic_version", version2),
					resource.TestCheckResourceAttrSet(datasourceName, "source_code_url"),
					resource.TestCheckResourceAttrSet(datasourceName, "template_url"),
					resource.TestCheckResourceAttr(datasourceName, "required_capabilities.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(datasourceName, "required_capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(datasourceName, "required_capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
		},
	})
}

func testAccCheckApplicationIDDataSource(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Serverless Repository Application data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AMI data source ID not set")
		}
		return nil
	}
}

func testAccApplicationDataSourceConfig_basic(appARN string) string {
	return fmt.Sprintf(`
data "aws_serverlessapplicationrepository_application" "secrets_manager_postgres_single_user_rotator" {
  application_id = %[1]q
}
`, appARN)
}

func testAccApplicationDataSourceConfig_versioned(appARN, version string) string {
	return fmt.Sprintf(`
data "aws_serverlessapplicationrepository_application" "secrets_manager_postgres_single_user_rotator" {
  application_id   = %[1]q
  semantic_version = %[2]q
}
`, appARN, version)
}
