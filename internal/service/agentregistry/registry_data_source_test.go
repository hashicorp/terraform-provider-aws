// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAgentRegistryRegistryDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	dataSourceName := "data.aws_agentregistry_registry.test"
	resourceName := "aws_agentregistry_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "registry_arn", resourceName, "registry_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "registry_id", resourceName, "registry_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, "discovery_configuration.#", resourceName, "discovery_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "discovery_configuration.0.authorizer_type", resourceName, "discovery_configuration.0.authorizer_type"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated_at"),
				),
			},
		},
	})
}

func testAccRegistryDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_agentregistry_registry" "test" {
  name        = %[1]q
  description = "data source test"

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}

data "aws_agentregistry_registry" "test" {
  registry_id = aws_agentregistry_registry.test.registry_id
}
`, rName)
}
