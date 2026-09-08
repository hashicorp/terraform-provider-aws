// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const resNameDataProvider = "Data Provider"

func TestAccDMSDataProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dms_data_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheckDataProvider(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dms", regexache.MustCompile(`data-provider:.+$`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrCreationTime, regexache.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z$`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrName, regexache.MustCompile(`^dp-\d+$`)),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "postgres"),
					resource.TestCheckNoResourceAttr(resourceName, "virtual"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.server_name", rName+".example.com"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.port", "5432"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.database_name", "example"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTags+".%", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTagsAll+".%", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDMSDataProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dms_data_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheckDataProvider(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdms.ResourceDataProvider, resourceName),
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

func TestAccDMSDataProvider_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dms_data_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheckDataProvider(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_update(rName, "first description", rName+".example.com", 5432, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "first description"),
					resource.TestCheckResourceAttr(resourceName, "virtual", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.server_name", rName+".example.com"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.port", "5432"),
				),
			},
			{
				Config: testAccDataProviderConfig_update(rName, "second description", "updated."+rName+".example.com", 5433, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "second description"),
					resource.TestCheckResourceAttr(resourceName, "virtual", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.server_name", "updated."+rName+".example.com"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgresql_settings.0.port", "5433"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDMSDataProvider_virtual(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dms_data_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheckDataProvider(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_update(rName, names.AttrDescription, rName+".example.com", 5432, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "virtual", acctest.CtFalse),
				),
			},
			{
				// Promoting a non-virtual data provider to virtual requires replacement.
				Config: testAccDataProviderConfig_update(rName, names.AttrDescription, rName+".example.com", 5432, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "virtual", acctest.CtTrue),
				),
			},
			{
				// Demoting a virtual data provider to non-virtual is an in-place update.
				Config: testAccDataProviderConfig_update(rName, names.AttrDescription, rName+".example.com", 5432, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "virtual", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckDataProviderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_data_provider" {
				continue
			}

			ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)
			arn := rs.Primary.Attributes[names.AttrARN]
			_, err := tfdms.FindDataProviderByARN(ctx, conn, arn)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.DMS, create.ErrActionCheckingDestroyed, resNameDataProvider, arn, err)
			}

			return create.Error(names.DMS, create.ErrActionCheckingDestroyed, resNameDataProvider, arn, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDataProviderExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, resNameDataProvider, name, errors.New("not found"))
		}

		arn := rs.Primary.Attributes[names.AttrARN]
		if arn == "" {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, resNameDataProvider, name, errors.New("arn not set"))
		}

		ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)
		_, err := tfdms.FindDataProviderByARN(ctx, conn, arn)
		if err != nil {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, resNameDataProvider, arn, err)
		}

		return nil
	}
}

func testAccPreCheckDataProvider(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)
	var input databasemigrationservice.DescribeDataProvidersInput
	_, err := conn.DescribeDataProviders(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDataProviderConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_data_provider" "test" {
  engine = "postgres"

  settings {
    postgresql_settings {
      server_name   = "%[1]s.example.com"
      port          = 5432
      database_name = "example"
    }
  }
}
`, rName)
}

func testAccDataProviderConfig_update(rName, description, serverName string, port int, virtual bool) string {
	return fmt.Sprintf(`
resource "aws_dms_data_provider" "test" {
  name        = %[1]q
  description = %[2]q
  engine      = "postgres"
  virtual     = %[5]t

  settings {
    postgresql_settings {
      server_name   = %[3]q
      port          = %[4]d
      database_name = "example"
    }
  }
}
`, rName, description, serverName, port, virtual)
}
