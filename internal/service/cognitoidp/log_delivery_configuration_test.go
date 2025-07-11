// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPLogDeliveryConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var logDeliveryConfiguration awstypes.LogDeliveryConfigurationType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognitoidp_log_delivery_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogDeliveryConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLogDeliveryConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogDeliveryConfigurationExists(ctx, resourceName, &logDeliveryConfiguration),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.event_source", "userNotification"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.log_level", "ERROR"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccLogDeliveryConfigurationImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrUserPoolID,
			},
		},
	})
}

func TestAccCognitoIDPLogDeliveryConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var logDeliveryConfiguration awstypes.LogDeliveryConfigurationType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognitoidp_log_delivery_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogDeliveryConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLogDeliveryConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogDeliveryConfigurationExists(ctx, resourceName, &logDeliveryConfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceLogDeliveryConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLogDeliveryConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognitoidp_log_delivery_configuration" {
				continue
			}

			_, err := tfcognitoidp.FindLogDeliveryConfigurationByUserPoolID(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID])

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.CognitoIDP, create.ErrActionCheckingDestroyed, tfcognitoidp.ResNameLogDeliveryConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckLogDeliveryConfigurationExists(ctx context.Context, name string, logDeliveryConfiguration *awstypes.LogDeliveryConfigurationType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CognitoIDP, create.ErrActionCheckingExistence, tfcognitoidp.ResNameLogDeliveryConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CognitoIDP, create.ErrActionCheckingExistence, tfcognitoidp.ResNameLogDeliveryConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		resp, err := tfcognitoidp.FindLogDeliveryConfigurationByUserPoolID(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID])

		if err != nil {
			return create.Error(names.CognitoIDP, create.ErrActionCheckingExistence, tfcognitoidp.ResNameLogDeliveryConfiguration, rs.Primary.ID, err)
		}

		*logDeliveryConfiguration = *resp

		return nil
	}
}

func testAccLogDeliveryConfigurationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrUserPoolID], nil
	}
}

func testAccLogDeliveryConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cognitoidp_log_delivery_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  log_configurations {
    event_source = "userNotification"
    log_level    = "ERROR"

    cloud_watch_logs_configuration {
      log_group_arn = aws_cloudwatch_log_group.test.arn
    }
  }
}
`, rName)
}
