// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/devopsguru/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfdevopsguru "github.com/hashicorp/terraform-provider-aws/internal/service/devopsguru"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccServiceIntegration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_devopsguru_service_integration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationConfig_basic(string(types.OptInStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceIntegrationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logs_anomaly_detection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logs_anomaly_detection.0.opt_in_status", string(types.OptInStatusEnabled)),
					resource.TestCheckResourceAttr(resourceName, "ops_center.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ops_center.0.opt_in_status", string(types.OptInStatusEnabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceIntegrationConfig_basic(string(types.OptInStatusDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceIntegrationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logs_anomaly_detection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logs_anomaly_detection.0.opt_in_status", string(types.OptInStatusDisabled)),
					resource.TestCheckResourceAttr(resourceName, "ops_center.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ops_center.0.opt_in_status", string(types.OptInStatusDisabled)),
				),
			},
		},
	})
}

func testAccServiceIntegration_kms(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_devopsguru_service_integration.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationConfig_kmsCustomerManaged(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceIntegrationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "kms_server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "kms_server_side_encryption.0.kms_key_id", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "kms_server_side_encryption.0.opt_in_status", string(types.OptInStatusEnabled)),
					resource.TestCheckResourceAttr(resourceName, "kms_server_side_encryption.0.type", string(types.ServerSideEncryptionTypeCustomerManagedKey)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceIntegrationConfig_kmsAWSOwned(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceIntegrationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "kms_server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kms_server_side_encryption.0.opt_in_status", string(types.OptInStatusEnabled)),
					resource.TestCheckResourceAttr(resourceName, "kms_server_side_encryption.0.type", string(types.ServerSideEncryptionTypeAwsOwnedKmsKey)),
				),
			},
		},
	})
}

func testAccCheckServiceIntegrationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DevOpsGuruClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devopsguru_service_integration" {
				continue
			}

			out, err := tfdevopsguru.FindServiceIntegration(ctx, conn)
			if errs.IsA[*tfresource.EmptyResultError](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DevOpsGuru, create.ErrActionCheckingDestroyed, tfdevopsguru.ResNameServiceIntegration, rs.Primary.ID, err)
			}

			// Because the resource does not disable anything on destroy, add some checks here
			// to ensure the test cases tidied up appropriately.
			if logs := out.LogsAnomalyDetection; logs != nil && logs.OptInStatus != types.OptInStatusDisabled {
				return create.Error(names.DevOpsGuru, create.ErrActionCheckingDestroyed, tfdevopsguru.ResNameServiceIntegration, rs.Primary.ID, errors.New("logs_anomaly_detection not disabled"))
			}
			if oc := out.OpsCenter; oc != nil && oc.OptInStatus != types.OptInStatusDisabled {
				return create.Error(names.DevOpsGuru, create.ErrActionCheckingDestroyed, tfdevopsguru.ResNameServiceIntegration, rs.Primary.ID, errors.New("ops_center not disabled"))
			}
		}

		return nil
	}
}

func testAccCheckServiceIntegrationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameServiceIntegration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameServiceIntegration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DevOpsGuruClient(ctx)
		_, err := tfdevopsguru.FindServiceIntegration(ctx, conn)
		if err != nil {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameServiceIntegration, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccServiceIntegrationConfig_basic(optInStatus string) string {
	return fmt.Sprintf(`
resource "aws_devopsguru_service_integration" "test" {
  # Default to existing configured settings
  kms_server_side_encryption {}

  logs_anomaly_detection {
    opt_in_status = %[1]q
  }
  ops_center {
    opt_in_status = %[1]q
  }
}
`, optInStatus)
}

func testAccServiceIntegrationConfig_kmsCustomerManaged() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_devopsguru_service_integration" "test" {
  kms_server_side_encryption {
    kms_key_id    = aws_kms_key.test.arn
    opt_in_status = "ENABLED"
    type          = "CUSTOMER_MANAGED_KEY"
  }

  logs_anomaly_detection {
    opt_in_status = "DISABLED"
  }
  ops_center {
    opt_in_status = "DISABLED"
  }
}
`
}

func testAccServiceIntegrationConfig_kmsAWSOwned() string { // nosemgrep:ci.aws-in-func-name
	return `
resource "aws_devopsguru_service_integration" "test" {
  kms_server_side_encryption {
    opt_in_status = "ENABLED"
    type          = "AWS_OWNED_KMS_KEY"
  }

  logs_anomaly_detection {
    opt_in_status = "DISABLED"
  }
  ops_center {
    opt_in_status = "DISABLED"
  }
}
`
}
