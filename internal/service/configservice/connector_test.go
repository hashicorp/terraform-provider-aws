// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccConnector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Connector
	resourceName := "aws_config_connector.test"
	clientID := os.Getenv("AZURE_CLIENT_IDENTIFIER")
	tenantID := os.Getenv("AZURE_TENANT_IDENTIFIER")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccConnectorPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_basic(clientID, tenantID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "config", regexache.MustCompile(`connector/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "azure.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "azure.0.client_identifier", clientID),
					resource.TestCheckResourceAttr(resourceName, "azure.0.tenant_identifier", tenantID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccConnector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Connector
	resourceName := "aws_config_connector.test"
	clientID := os.Getenv("AZURE_CLIENT_IDENTIFIER")
	tenantID := os.Getenv("AZURE_TENANT_IDENTIFIER")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccConnectorPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_basic(clientID, tenantID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfconfig.ResourceConnector, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccConnector_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Connector
	resourceName := "aws_config_connector.test"
	clientID := os.Getenv("AZURE_CLIENT_IDENTIFIER")
	tenantID := os.Getenv("AZURE_TENANT_IDENTIFIER")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccConnectorPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_tags1(clientID, tenantID, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", acctest.CtValue1),
				),
			},
			{
				Config: testAccConnectorConfig_tags2(clientID, tenantID, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", acctest.CtValue2),
				),
			},
		},
	})
}

func testAccConnectorPreCheck(t *testing.T) {
	if os.Getenv("AZURE_CLIENT_IDENTIFIER") == "" || os.Getenv("AZURE_TENANT_IDENTIFIER") == "" {
		t.Skip("AZURE_CLIENT_IDENTIFIER and AZURE_TENANT_IDENTIFIER must be set for ConfigService Connector acceptance tests")
	}
}

func testAccCheckConnectorExists(ctx context.Context, t *testing.T, n string, v *awstypes.Connector) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ConfigServiceClient(ctx)

		output, err := tfconfig.FindConnectorByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_connector" {
				continue
			}

			_, err := tfconfig.FindConnectorByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ConfigService Connector %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccConnectorConfig_basic(clientID, tenantID string) string {
	return fmt.Sprintf(`
resource "aws_config_connector" "test" {
  azure {
    client_identifier = %[1]q
    tenant_identifier = %[2]q
  }
}
`, clientID, tenantID)
}

func testAccConnectorConfig_tags1(clientID, tenantID, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_config_connector" "test" {
  azure {
    client_identifier = %[1]q
    tenant_identifier = %[2]q
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, clientID, tenantID, tagKey1, tagValue1)
}

func testAccConnectorConfig_tags2(clientID, tenantID, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_config_connector" "test" {
  azure {
    client_identifier = %[1]q
    tenant_identifier = %[2]q
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, clientID, tenantID, tagKey1, tagValue1, tagKey2, tagValue2)
}
