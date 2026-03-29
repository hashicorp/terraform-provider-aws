// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecret_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "secretsmanager", regexache.MustCompile(fmt.Sprintf("secret:%s-[[:alnum:]]+$", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "30"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_withNamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	resourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsecretsmanager.ResourceSecret(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSecretsManagerSecret_description(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccSecretConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_basicReplica(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckSecretDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_basicReplica(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecret_overwriteReplica(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 3) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckSecretDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_overwriteReplica(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", acctest.CtTrue),
				),
			},
			{
				Config: testAccSecretConfig_overwriteReplicaUpdate(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", acctest.CtTrue),
				),
			},
			{
				Config: testAccSecretConfig_overwriteReplica(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecret_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyID),
				),
			},
			{
				Config: testAccSecretConfig_kmsKeyIDUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_RecoveryWindowInDays_recreate(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_recoveryWindowInDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "0"),
				),
			},
			{
				Config: testAccSecretConfig_recoveryWindowInDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "0"),
				),
				Taint: []string{resourceName},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "San Holo feat. Duskus"),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy,
						regexache.MustCompile(`{"Action":"secretsmanager:GetSecretValue".+`)),
				),
			},
			{
				Config: testAccSecretConfig_policyEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Poliça"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
				),
			},
			{
				Config: testAccSecretConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, t, resourceName, &secret),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy,
						regexache.MustCompile(`{"Action":"secretsmanager:GetSecretValue".+`)),
				),
			},
		},
	})
}

func testAccCheckSecretDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SecretsManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_secretsmanager_secret" {
				continue
			}

			_, err := tfsecretsmanager.FindSecretByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Secrets Manager Secret %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSecretExists(ctx context.Context, t *testing.T, n string, v *secretsmanager.DescribeSecretOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecretsManagerClient(ctx)

		output, err := tfsecretsmanager.FindSecretByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).SecretsManagerClient(ctx)

	input := &secretsmanager.ListSecretsInput{}

	_, err := conn.ListSecrets(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSecretConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  description = %[1]q
  name        = %[2]q
}
`, description, rName)
}

func testAccSecretConfig_basicReplica(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q

  replica {
    region = data.aws_region.alternate.region
  }
}
`, rName))
}

func testAccSecretConfig_overwriteReplica(rName string, force_overwrite_replica_secret bool) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(3), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  provider                = awsalternate
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_key" "test2" {
  provider                = awsthird
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_secretsmanager_secret" "test" {
  name                           = %[1]q
  force_overwrite_replica_secret = %[2]t

  replica {
    kms_key_id = aws_kms_key.test.key_id
    region     = data.aws_region.alternate.region
  }
}
`, rName, force_overwrite_replica_secret))
}

func testAccSecretConfig_overwriteReplicaUpdate(rName string, force_overwrite_replica_secret bool) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(3), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  provider                = awsalternate
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_key" "test2" {
  provider                = awsthird
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

data "aws_region" "third" {
  provider = awsthird
}

resource "aws_secretsmanager_secret" "test" {
  name                           = %[1]q
  force_overwrite_replica_secret = %[2]t

  replica {
    kms_key_id = aws_kms_key.test2.key_id
    region     = data.aws_region.third.region
  }
}
`, rName, force_overwrite_replica_secret))
}

func testAccSecretConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}
`, rName)
}

func testAccSecretConfig_namePrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name_prefix = %[1]q
}
`, rName)
}

func testAccSecretConfig_kmsKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_secretsmanager_secret" "test" {
  kms_key_id = aws_kms_key.test1.id
  name       = %[1]q
}
`, rName)
}

func testAccSecretConfig_kmsKeyIDUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_secretsmanager_secret" "test" {
  kms_key_id = aws_kms_key.test2.id
  name       = %[1]q
}
`, rName)
}

func testAccSecretConfig_recoveryWindowInDays(rName string, recoveryWindowInDays int) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name                    = %[1]q
  recovery_window_in_days = %[2]d
}
`, rName, recoveryWindowInDays)
}

func testAccSecretConfig_policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Principal = {
        Service = "ec2.amazonaws.com"
      },
      Effect = "Allow"
      Sid    = ""
    }]
  })
}

resource "aws_secretsmanager_secret" "test" {
  name        = %[1]q
  description = "San Holo feat. Duskus"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "EnableAllPermissions"
      Effect = "Allow"
      Principal = {
        AWS = aws_iam_role.test.arn
      }
      Action   = "secretsmanager:GetSecretValue"
      Resource = "*"
    }]
  })
}
`, rName)
}

func testAccSecretConfig_policyEmpty(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Principal = {
        Service = "ec2.amazonaws.com"
      },
      Effect = "Allow"
      Sid    = ""
    }]
  })
}

resource "aws_secretsmanager_secret" "test" {
  name        = %[1]q
  description = "Poliça"

  policy = "{}"
}
`, rName)
}
