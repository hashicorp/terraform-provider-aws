// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccWorkteam_cognitoConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var workteam sagemaker.Workteam
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workteam.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkteamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkteamConfig_cognito(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.cognito_member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.client_id", "aws_cognito_user_pool_client.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_pool", "aws_cognito_user_pool.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_group", "aws_cognito_user_group.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"workforce_name"},
			},
			{
				Config: testAccWorkteamConfig_cognitoUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.cognito_member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.client_id", "aws_cognito_user_pool_client.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_pool", "aws_cognito_user_pool.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_group", "aws_cognito_user_group.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "member_definition.1.cognito_member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.1.cognito_member_definition.0.client_id", "aws_cognito_user_pool_client.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.1.cognito_member_definition.0.user_pool", "aws_cognito_user_pool.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.1.cognito_member_definition.0.user_group", "aws_cognito_user_group.test2", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
				),
			},
			{
				Config: testAccWorkteamConfig_cognito(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.cognito_member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.client_id", "aws_cognito_user_pool_client.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_pool", "aws_cognito_user_pool.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_group", "aws_cognito_user_group.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
				),
			},
		},
	})
}

func testAccWorkteam_oidcConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var workteam sagemaker.Workteam
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workteam.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkteamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkteamConfig_oidc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.*", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"workforce_name"},
			},
			{
				Config: testAccWorkteamConfig_oidc2(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.*", rName),
					resource.TestCheckTypeSetElemAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.*", "test"),
				),
			},
			{
				Config: testAccWorkteamConfig_oidc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.*", rName)),
			},
		},
	})
}

func testAccWorkteam_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var workteam sagemaker.Workteam
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workteam.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkteamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkteamConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"workforce_name"},
			},
			{
				Config: testAccWorkteamConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccWorkteamConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccWorkteam_notificationConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var workteam sagemaker.Workteam
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workteam.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkteamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkteamConfig_notification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "notification_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "notification_configuration.0.notification_topic_arn", "aws_sns_topic.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"workforce_name"},
			},
			{
				Config: testAccWorkteamConfig_oidc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "notification_configuration.#", acctest.Ct1),
				),
			},
			{
				Config: testAccWorkteamConfig_notification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "notification_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "notification_configuration.0.notification_topic_arn", "aws_sns_topic.test", names.AttrARN),
				),
			},
		},
	})
}

func testAccWorkteam_workerAccessConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var workteam sagemaker.Workteam
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workteam.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkteamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkteamConfig_workerAccessConfiguration(rName, "Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "worker_access_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "worker_access_configuration.0.s3_presign.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "worker_access_configuration.0.s3_presign.0.iam_policy_constraints.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "worker_access_configuration.0.s3_presign.0.iam_policy_constraints.0.source_ip", "Enabled"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"workforce_name"},
			},
			{
				Config: testAccWorkteamConfig_workerAccessConfiguration(rName, "Disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "worker_access_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "worker_access_configuration.0.s3_presign.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "worker_access_configuration.0.s3_presign.0.iam_policy_constraints.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "worker_access_configuration.0.s3_presign.0.iam_policy_constraints.0.source_ip", "Disabled"),
				),
			},
		},
	})
}

func testAccWorkteam_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var workteam sagemaker.Workteam
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workteam.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkteamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkteamConfig_oidc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(ctx, resourceName, &workteam),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceWorkteam(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckWorkteamDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_workteam" {
				continue
			}

			_, err := tfsagemaker.FindWorkteamByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker Workteam %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWorkteamExists(ctx context.Context, n string, workteam *sagemaker.Workteam) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker Workteam ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		output, err := tfsagemaker.FindWorkteamByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*workteam = *output

		return nil
	}
}

func testAccWorkteamCognitoBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  name            = %[1]q
  generate_secret = true
  user_pool_id    = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_group" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_sagemaker_workforce" "test" {
  workforce_name = %[1]q

  cognito_config {
    client_id = aws_cognito_user_pool_client.test.id
    user_pool = aws_cognito_user_pool_domain.test.user_pool_id
  }
}
`, rName)
}

func testAccWorkteamConfig_cognito(rName string) string {
	return acctest.ConfigCompose(testAccWorkteamCognitoBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_workteam" "test" {
  workteam_name  = %[1]q
  workforce_name = aws_sagemaker_workforce.test.id
  description    = %[1]q

  member_definition {
    cognito_member_definition {
      client_id  = aws_cognito_user_pool_client.test.id
      user_pool  = aws_cognito_user_pool_domain.test.user_pool_id
      user_group = aws_cognito_user_group.test.id
    }
  }
}
`, rName))
}

func testAccWorkteamConfig_cognitoUpdated(rName string) string {
	return acctest.ConfigCompose(testAccWorkteamCognitoBaseConfig(rName), fmt.Sprintf(`
resource "aws_cognito_user_group" "test2" {
  name         = "%[1]s-2"
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_sagemaker_workteam" "test" {
  workteam_name  = %[1]q
  workforce_name = aws_sagemaker_workforce.test.id
  description    = %[1]q

  member_definition {
    cognito_member_definition {
      client_id  = aws_cognito_user_pool_client.test.id
      user_pool  = aws_cognito_user_pool_domain.test.user_pool_id
      user_group = aws_cognito_user_group.test.id
    }
  }

  member_definition {
    cognito_member_definition {
      client_id  = aws_cognito_user_pool_client.test.id
      user_pool  = aws_cognito_user_pool_domain.test.user_pool_id
      user_group = aws_cognito_user_group.test2.id
    }
  }
}
`, rName))
}

func testAccWorkteamOIDCBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_workforce" "test" {
  workforce_name = %[1]q

  oidc_config {
    authorization_endpoint = "https://example.com"
    client_id              = %[1]q
    client_secret          = %[1]q
    issuer                 = "https://example.com"
    jwks_uri               = "https://example.com"
    logout_endpoint        = "https://example.com"
    token_endpoint         = "https://example.com"
    user_info_endpoint     = "https://example.com"
  }
}
`, rName)
}

func testAccWorkteamConfig_oidc(rName string) string {
	return acctest.ConfigCompose(testAccWorkteamOIDCBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_workteam" "test" {
  workteam_name  = %[1]q
  workforce_name = aws_sagemaker_workforce.test.id
  description    = %[1]q

  member_definition {
    oidc_member_definition {
      groups = [%[1]q]
    }
  }
}
`, rName))
}

func testAccWorkteamConfig_oidc2(rName, group string) string {
	return acctest.ConfigCompose(testAccWorkteamOIDCBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_workteam" "test" {
  workteam_name  = %[1]q
  workforce_name = aws_sagemaker_workforce.test.id
  description    = %[1]q

  member_definition {
    oidc_member_definition {
      groups = [%[1]q, %[2]q]
    }
  }
}
`, rName, group))
}

func testAccWorkteamConfig_notification(rName string) string {
	return acctest.ConfigCompose(testAccWorkteamOIDCBaseConfig(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn = aws_sns_topic.test.arn

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Id" : "default",
    "Statement" : [
      {
        "Sid" : "%[1]s",
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "sagemaker.amazonaws.com"
        },
        "Action" : [
          "sns:Publish"
        ],
        "Resource" : aws_sns_topic.test.arn
      }
    ]
  })
}

resource "aws_sagemaker_workteam" "test" {
  workteam_name  = %[1]q
  workforce_name = aws_sagemaker_workforce.test.id
  description    = %[1]q

  member_definition {
    oidc_member_definition {
      groups = [%[1]q]
    }
  }

  notification_configuration {
    notification_topic_arn = aws_sns_topic.test.arn
  }
}
`, rName))
}

func testAccWorkteamConfig_workerAccessConfiguration(rName, status string) string {
	return acctest.ConfigCompose(testAccWorkteamOIDCBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_workteam" "test" {
  workteam_name  = %[1]q
  workforce_name = aws_sagemaker_workforce.test.id
  description    = %[1]q

  member_definition {
    oidc_member_definition {
      groups = [%[1]q]
    }
  }

  worker_access_configuration {
    s3_presign {
      iam_policy_constraints {
        source_ip = %[2]q
      }
    }
  }
}
`, rName, status))
}

func testAccWorkteamConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccWorkteamOIDCBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_workteam" "test" {
  workteam_name  = %[1]q
  workforce_name = aws_sagemaker_workforce.test.id
  description    = %[1]q

  member_definition {
    oidc_member_definition {
      groups = [%[1]q]
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccWorkteamConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccWorkteamOIDCBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_workteam" "test" {
  workteam_name  = %[1]q
  workforce_name = aws_sagemaker_workforce.test.id
  description    = %[1]q

  member_definition {
    oidc_member_definition {
      groups = [%[1]q]
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
