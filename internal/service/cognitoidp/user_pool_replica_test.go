// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserPoolReplica_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var replica awstypes.UserPoolReplicaType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_replica.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckIdentityProvider(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckUserPoolReplicaDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolReplicaConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolReplicaExists(ctx, t, resourceName, &replica),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrUserPoolID),
					resource.TestCheckResourceAttrSet(resourceName, "region_name"),
					resource.TestCheckResourceAttr(resourceName, "role", "SECONDARY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "INACTIVE"),
					resource.TestCheckResourceAttrSet(resourceName, "user_pool_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccUserPoolReplicaImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrUserPoolID,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolReplica_status(t *testing.T) {
	ctx := acctest.Context(t)
	var replica awstypes.UserPoolReplicaType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_replica.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckIdentityProvider(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckUserPoolReplicaDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolReplicaConfig_status(rName, "ACTIVE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolReplicaExists(ctx, t, resourceName, &replica),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
				),
			},
			{
				Config: testAccUserPoolReplicaConfig_status(rName, "INACTIVE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolReplicaExists(ctx, t, resourceName, &replica),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "INACTIVE"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolReplica_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var replica awstypes.UserPoolReplicaType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_replica.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckIdentityProvider(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckUserPoolReplicaDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolReplicaConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolReplicaExists(ctx, t, resourceName, &replica),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcognitoidp.ResourceUserPoolReplica, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccUserPoolReplicaImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return rs.Primary.Attributes[names.AttrUserPoolID] + "," + rs.Primary.Attributes["region_name"], nil
	}
}

func testAccCheckUserPoolReplicaExists(ctx context.Context, t *testing.T, n string, v *awstypes.UserPoolReplicaType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		output, err := tfcognitoidp.FindUserPoolReplicaByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes["region_name"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckUserPoolReplicaDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user_pool_replica" {
				continue
			}

			_, err := tfcognitoidp.FindUserPoolReplicaByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes["region_name"])
			if retry.NotFound(err) {
				continue
			}
			// The associated KMS key may be scheduled for deletion (disabled) by the time
			// this check runs, causing Cognito to return a KMS validation error instead of
			// ResourceNotFoundException. Treat this as confirmation the replica is gone.
			if err != nil && strings.Contains(err.Error(), "is in an invalid state") {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito User Pool Replica %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccUserPoolReplicaConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  multi_region            = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Cognito to use this key"
        Effect = "Allow"
        Principal = {
          Service = "cognito-idp.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant",
          "kms:DescribeKey",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_kms_replica_key" "test" {
  provider        = awsalternate
  primary_key_arn = aws_kms_key.test.arn
  description     = %[1]q

  deletion_window_in_days = 7

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Cognito to use this key"
        Effect = "Allow"
        Principal = {
          Service = "cognito-idp.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant",
          "kms:DescribeKey",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  key_configuration {
    key_type    = "CUSTOMER_MANAGED_KEY"
    kms_key_arn = aws_kms_key.test.arn
  }
}

resource "aws_cognito_user_pool_replica" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  region_name  = data.aws_region.alternate.region

  depends_on = [aws_kms_replica_key.test]
}
`, rName))
}

func testAccUserPoolReplicaConfig_status(rName, status string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  multi_region            = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Cognito to use this key"
        Effect = "Allow"
        Principal = {
          Service = "cognito-idp.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant",
          "kms:DescribeKey",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_kms_replica_key" "test" {
  provider        = awsalternate
  primary_key_arn = aws_kms_key.test.arn
  description     = %[1]q

  deletion_window_in_days = 7

  # When a Cognito user pool replica is activated, AWS takes control of this key
  # and may disable it from Terraform's perspective. Ignore drift on enabled to
  # prevent a non-empty plan from blocking subsequent test steps.
  lifecycle {
    ignore_changes = [enabled]
  }

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Cognito to use this key"
        Effect = "Allow"
        Principal = {
          Service = "cognito-idp.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant",
          "kms:DescribeKey",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  key_configuration {
    key_type    = "CUSTOMER_MANAGED_KEY"
    kms_key_arn = aws_kms_key.test.arn
  }
}

resource "aws_cognito_user_pool_replica" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  region_name  = data.aws_region.alternate.region
  status       = %[2]q

  depends_on = [aws_kms_replica_key.test]
}
`, rName, status))
}

func TestAccCognitoIDPUserPoolReplica_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var replica awstypes.UserPoolReplicaType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_replica.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckIdentityProvider(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckUserPoolReplicaDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolReplicaConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolReplicaExists(ctx, t, resourceName, &replica),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccUserPoolReplicaConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolReplicaExists(ctx, t, resourceName, &replica),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccUserPoolReplicaConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  multi_region            = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Cognito to use this key"
        Effect = "Allow"
        Principal = {
          Service = "cognito-idp.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant",
          "kms:DescribeKey",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_kms_replica_key" "test" {
  provider        = awsalternate
  primary_key_arn = aws_kms_key.test.arn
  description     = %[1]q

  deletion_window_in_days = 7

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Cognito to use this key"
        Effect = "Allow"
        Principal = {
          Service = "cognito-idp.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant",
          "kms:DescribeKey",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  key_configuration {
    key_type    = "CUSTOMER_MANAGED_KEY"
    kms_key_arn = aws_kms_key.test.arn
  }
}

resource "aws_cognito_user_pool_replica" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  region_name  = data.aws_region.alternate.region

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_kms_replica_key.test]
}
`, rName, tagKey1, tagValue1))
}

func testAccUserPoolReplicaConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  multi_region            = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Cognito to use this key"
        Effect = "Allow"
        Principal = {
          Service = "cognito-idp.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant",
          "kms:DescribeKey",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_kms_replica_key" "test" {
  provider        = awsalternate
  primary_key_arn = aws_kms_key.test.arn
  description     = %[1]q

  deletion_window_in_days = 7

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Cognito to use this key"
        Effect = "Allow"
        Principal = {
          Service = "cognito-idp.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant",
          "kms:DescribeKey",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  key_configuration {
    key_type    = "CUSTOMER_MANAGED_KEY"
    kms_key_arn = aws_kms_key.test.arn
  }
}

resource "aws_cognito_user_pool_replica" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  region_name  = data.aws_region.alternate.region

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_kms_replica_key.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
