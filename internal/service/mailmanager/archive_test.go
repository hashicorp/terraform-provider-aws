// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmailmanager "github.com/hashicorp/terraform-provider-aws/internal/service/mailmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMailManagerArchive_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_archive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccArchivePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", regexache.MustCompile(`mailmanager-archive/.+`)),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_timestamp"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_updated_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrKMSKeyARN),
					resource.TestCheckResourceAttr(resourceName, "archive_state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "retention.#", "0"),
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

func TestAccMailManagerArchive_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_archive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccArchivePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfmailmanager.ResourceArchive, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccMailManagerArchive_update(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_archive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccArchivePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccArchiveConfig_basic(rNameUpdated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
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

func TestAccMailManagerArchive_retention(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_archive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccArchivePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_retention(rName, "ONE_YEAR"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retention.0.retention_period", "ONE_YEAR"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccArchiveConfig_retention(rName, "TWO_YEARS"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retention.0.retention_period", "TWO_YEARS"),
				),
			},
		},
	})
}

func testAccCheckArchiveDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mailmanager_archive" {
				continue
			}
			_, err := tfmailmanager.FindArchiveByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return smarterr.NewError(err)
			}
			return smarterr.NewError(errors.New("SES Mail Manager Archive still exists"))
		}
		return nil
	}
}

func testAccCheckArchiveExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return smarterr.NewError(errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return smarterr.NewError(errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)

		_, err := tfmailmanager.FindArchiveByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return smarterr.NewError(err)
		}
		return nil
	}
}

func testAccArchivePreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)

	var input mailmanager.ListArchivesInput

	_, err := conn.ListArchives(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccArchiveConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_archive" "test" {
  name = %[1]q
}
`, rName)
}

func testAccArchiveConfig_retention(rName, retentionPeriod string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_archive" "test" {
  name = %[1]q

  retention {
    retention_period = %[2]q
  }
}
`, rName, retentionPeriod)
}
