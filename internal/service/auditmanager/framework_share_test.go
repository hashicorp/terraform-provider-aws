// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package auditmanager_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfauditmanager "github.com/hashicorp/terraform-provider-aws/internal/service/auditmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAuditManagerFrameworkShare_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var frameworkShare types.AssessmentFrameworkShareRequest
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_framework_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFrameworkShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkShareConfig_basic(rName, acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkShareExists(ctx, t, resourceName, &frameworkShare),
					resource.TestCheckResourceAttrPair(resourceName, "destination_account", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "framework_id", "aws_auditmanager_framework.test", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
		},
	})
}

func TestAccAuditManagerFrameworkShare_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var frameworkShare types.AssessmentFrameworkShareRequest
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_framework_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFrameworkShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkShareConfig_basic(rName, acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkShareExists(ctx, t, resourceName, &frameworkShare),
					// Sleep briefly to prevent intermittent validation errors when revoking
					// a new framework share request
					acctest.CheckSleep(t, 10*time.Second),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfauditmanager.ResourceFrameworkShare, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAuditManagerFrameworkShare_optional(t *testing.T) {
	ctx := acctest.Context(t)
	var frameworkShare types.AssessmentFrameworkShareRequest
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_framework_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFrameworkShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkShareConfig_optional(rName, acctest.AlternateRegion(), "text"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkShareExists(ctx, t, resourceName, &frameworkShare),
					resource.TestCheckResourceAttrPair(resourceName, "destination_account", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "framework_id", "aws_auditmanager_framework.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "text"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
			{
				Config: testAccFrameworkShareConfig_basic(rName, acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkShareExists(ctx, t, resourceName, &frameworkShare),
					resource.TestCheckResourceAttrPair(resourceName, "destination_account", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "framework_id", "aws_auditmanager_framework.test", names.AttrID),
				),
			},
			{
				Config: testAccFrameworkShareConfig_optional(rName, acctest.AlternateRegion(), "text-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkShareExists(ctx, t, resourceName, &frameworkShare),
					resource.TestCheckResourceAttrPair(resourceName, "destination_account", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "framework_id", "aws_auditmanager_framework.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "text-updated"),
				),
			},
		},
	})
}

func testAccCheckFrameworkShareDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_auditmanager_framework_share" {
				continue
			}

			_, err := tfauditmanager.FindFrameworkShareByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Audit Manager Framework Share %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFrameworkShareExists(ctx context.Context, t *testing.T, n string, v *types.AssessmentFrameworkShareRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		output, err := tfauditmanager.FindFrameworkShareByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFrameworkShareConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_auditmanager_control" "test" {
  name = %[1]q

  control_mapping_sources {
    source_name          = %[1]q
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }
}

resource "aws_auditmanager_framework" "test" {
  name = %[1]q

  control_sets {
    name = %[1]q
    controls {
      id = aws_auditmanager_control.test.id
    }
  }
}
`, rName)
}

func testAccFrameworkShareConfig_basic(rName, destinationRegion string) string {
	return acctest.ConfigCompose(
		testAccFrameworkShareConfig_base(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_framework_share" "test" {
  destination_account = data.aws_caller_identity.current.account_id
  destination_region  = %[1]q
  framework_id        = aws_auditmanager_framework.test.id
}
`, destinationRegion))
}

func testAccFrameworkShareConfig_optional(rName, destinationRegion, comment string) string {
	return acctest.ConfigCompose(
		testAccFrameworkShareConfig_base(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_framework_share" "test" {
  destination_account = data.aws_caller_identity.current.account_id
  destination_region  = %[1]q
  framework_id        = aws_auditmanager_framework.test.id

  comment = %[2]q
}
`, destinationRegion, comment))
}
