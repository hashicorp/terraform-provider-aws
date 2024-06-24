// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfauditmanager "github.com/hashicorp/terraform-provider-aws/internal/service/auditmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestCanBeRevoked(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		status types.ShareRequestStatus
		want   bool
	}{
		{"active", types.ShareRequestStatusActive, true},
		{"declined", types.ShareRequestStatusDeclined, false},
		{"expiring", types.ShareRequestStatusExpiring, true},
		{"expired", types.ShareRequestStatusExpired, false},
		{"failed", types.ShareRequestStatusFailed, false},
		{"replicating", types.ShareRequestStatusReplicating, true},
		{"revoked", types.ShareRequestStatusRevoked, false},
		{"shared", types.ShareRequestStatusShared, true},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := tfauditmanager.CanBeRevoked(string(testCase.status)); got != testCase.want {
				t.Errorf("CanBeRevoked() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestAccAuditManagerFrameworkShare_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var frameworkShare types.AssessmentFrameworkShareRequest
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_framework_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFrameworkShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkShareConfig_basic(rName, acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkShareExists(ctx, resourceName, &frameworkShare),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_framework_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFrameworkShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkShareConfig_basic(rName, acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkShareExists(ctx, resourceName, &frameworkShare),
					// Sleep briefly to prevent intermittent validation errors when revoking
					// a new framework share request
					acctest.CheckSleep(t, 10*time.Second),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfauditmanager.ResourceFrameworkShare, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAuditManagerFrameworkShare_optional(t *testing.T) {
	ctx := acctest.Context(t)
	var frameworkShare types.AssessmentFrameworkShareRequest
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_framework_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFrameworkShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkShareConfig_optional(rName, acctest.AlternateRegion(), "text"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkShareExists(ctx, resourceName, &frameworkShare),
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
					testAccCheckFrameworkShareExists(ctx, resourceName, &frameworkShare),
					resource.TestCheckResourceAttrPair(resourceName, "destination_account", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "framework_id", "aws_auditmanager_framework.test", names.AttrID),
				),
			},
			{
				Config: testAccFrameworkShareConfig_optional(rName, acctest.AlternateRegion(), "text-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkShareExists(ctx, resourceName, &frameworkShare),
					resource.TestCheckResourceAttrPair(resourceName, "destination_account", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "framework_id", "aws_auditmanager_framework.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "text-updated"),
				),
			},
		},
	})
}

func testAccCheckFrameworkShareDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AuditManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_auditmanager_framework_share" {
				continue
			}

			_, err := tfauditmanager.FindFrameworkShareByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *retry.NotFoundError
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.AuditManager, create.ErrActionCheckingDestroyed, tfauditmanager.ResNameFrameworkShare, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFrameworkShareExists(ctx context.Context, name string, frameworkShare *types.AssessmentFrameworkShareRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameFrameworkShare, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameFrameworkShare, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AuditManagerClient(ctx)
		resp, err := tfauditmanager.FindFrameworkShareByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameFrameworkShare, rs.Primary.ID, err)
		}

		*frameworkShare = *resp

		return nil
	}
}

func testAccFrameworkShareConfigBase(rName string) string {
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
		testAccFrameworkShareConfigBase(rName),
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
		testAccFrameworkShareConfigBase(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_framework_share" "test" {
  destination_account = data.aws_caller_identity.current.account_id
  destination_region  = %[1]q
  framework_id        = aws_auditmanager_framework.test.id

  comment = %[2]q
}
`, destinationRegion, comment))
}
