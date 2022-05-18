package sagemaker_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccWorkteam_cognitoConfig(t *testing.T) {
	var workteam sagemaker.Workteam
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workteam.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkteamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkteamCognitoConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.cognito_member_definition.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.client_id", "aws_cognito_user_pool_client.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_pool", "aws_cognito_user_pool.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_group", "aws_cognito_user_group.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"workforce_name"},
			},
			{
				Config: testAccWorkteamCognitoUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.cognito_member_definition.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.client_id", "aws_cognito_user_pool_client.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_pool", "aws_cognito_user_pool.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_group", "aws_cognito_user_group.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "member_definition.1.cognito_member_definition.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.1.cognito_member_definition.0.client_id", "aws_cognito_user_pool_client.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.1.cognito_member_definition.0.user_pool", "aws_cognito_user_pool.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.1.cognito_member_definition.0.user_group", "aws_cognito_user_group.test2", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
				),
			},
			{
				Config: testAccWorkteamCognitoConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.cognito_member_definition.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.client_id", "aws_cognito_user_pool_client.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_pool", "aws_cognito_user_pool.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "member_definition.0.cognito_member_definition.0.user_group", "aws_cognito_user_group.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
				),
			},
		},
	})
}

func testAccWorkteam_oidcConfig(t *testing.T) {
	var workteam sagemaker.Workteam
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workteam.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkteamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkteamOIDCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.#", "1"),
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
				Config: testAccWorkteamOIDC2Config(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.*", rName),
					resource.TestCheckTypeSetElemAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.*", "test"),
				),
			},
			{
				Config: testAccWorkteamOIDCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "member_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "member_definition.0.oidc_member_definition.0.groups.*", rName)),
			},
		},
	})
}

func testAccWorkteam_tags(t *testing.T) {
	var workteam sagemaker.Workteam
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workteam.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkteamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkteamTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"workforce_name"},
			},
			{
				Config: testAccWorkteamTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWorkteamTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccWorkteam_notificationConfig(t *testing.T) {
	var workteam sagemaker.Workteam
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workteam.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkteamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkteamNotificationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "notification_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "notification_configuration.0.notification_topic_arn", "aws_sns_topic.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"workforce_name"},
			},
			{
				Config: testAccWorkteamOIDCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "notification_configuration.#", "1"),
				),
			},
			{
				Config: testAccWorkteamNotificationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					resource.TestCheckResourceAttr(resourceName, "workteam_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workteam/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "notification_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "notification_configuration.0.notification_topic_arn", "aws_sns_topic.test", "arn"),
				),
			},
		},
	})
}

func testAccWorkteam_disappears(t *testing.T) {
	var workteam sagemaker.Workteam
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workteam.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkteamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkteamOIDCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkteamExists(resourceName, &workteam),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceWorkteam(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckWorkteamDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_workteam" {
			continue
		}

		_, err := tfsagemaker.FindWorkteamByName(conn, rs.Primary.ID)

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

func testAccCheckWorkteamExists(n string, workteam *sagemaker.Workteam) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker Workteam ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

		output, err := tfsagemaker.FindWorkteamByName(conn, rs.Primary.ID)

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

func testAccWorkteamCognitoConfig(rName string) string {
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

func testAccWorkteamCognitoUpdatedConfig(rName string) string {
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

func testAccWorkteamOIDCConfig(rName string) string {
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

func testAccWorkteamOIDC2Config(rName, group string) string {
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

func testAccWorkteamNotificationConfig(rName string) string {
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
        "Resource" : "${aws_sns_topic.test.arn}"
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

func testAccWorkteamTags1Config(rName, tagKey1, tagValue1 string) string {
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

func testAccWorkteamTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
