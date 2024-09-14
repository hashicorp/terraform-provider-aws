// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftServerlessNamespace_emptyPlanWithArrayOutputs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withLogExportsOutput(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withNameOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "redshift-serverless", regexache.MustCompile("namespace/.+$")),
					resource.TestCheckResourceAttr(resourceName, "namespace_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "namespace_id"),
					resource.TestCheckResourceAttr(resourceName, "log_exports.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "admin_username", "admin"),
					resource.TestCheckResourceAttr(resourceName, "db_name", "dev"),
					resource.TestCheckResourceAttr(resourceName, "manage_admin_password", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNamespaceConfig_withMultipleAttrsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "redshift-serverless", regexache.MustCompile("namespace/.+$")),
					resource.TestCheckResourceAttr(resourceName, "namespace_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "namespace_id"),
					resource.TestCheckResourceAttr(resourceName, "log_exports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "iam_roles.*", "aws_iam_role.test.0", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "iam_roles.*", "aws_iam_role.test.1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "admin_username", "admin"),
					resource.TestCheckResourceAttr(resourceName, "db_name", "dev"),
					resource.TestCheckResourceAttr(resourceName, "manage_admin_password", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_defaultIAMRole(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withDefaultIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "namespace_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "default_iam_role_arn", "aws_iam_role.test.0", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNamespaceConfig_withoutDefaultIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "default_iam_role_arn"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_defaultUserAndCustomPassword(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withDefaultUserAndUnmanagedPassword(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_user_password", "Password123"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_customUserAndUnmanagedPassword(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withUnmanagedPasswordAndManageFlag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_user_password", "Password123"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_userPasswordWriteOnly(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_userPasswordWriteOnly(rName, "Password123", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
				),
			},
			{
				Config: testAccNamespaceConfig_userPasswordWriteOnly(rName, "Password123updated", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withTags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNamespaceConfig_withTags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccNamespaceConfig_withTags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withNameOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfredshiftserverless.ResourceNamespace, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_withWorkgroup(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withWorkgroup(rName),
				Check:  resource.ComposeTestCheckFunc(testAccCheckNamespaceExists(ctx, resourceName)),
			},
			{
				Config:   testAccNamespaceConfig_withWorkgroup(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_manageAdminPassword(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withUnmanagedPassword(rName, "testuser"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "admin_password_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "admin_username", "testuser"),
					resource.TestCheckResourceAttr(resourceName, "manage_admin_password", acctest.CtFalse),
				),
			},
			{
				Config: testAccNamespaceConfig_withManagedPassword(rName, "testuser"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "manage_admin_password", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "admin_password_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "admin_username", "testuser"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_replaceOnKMSKeyRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withKMSKey(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
				},
			},
			{
				Config: testAccNamespaceConfig_withNameOnly(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_iamRolesCleared(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withDefaultIAMRole(rName),
			},
			{
				Config: testAccNamespaceConfig_withNameOnly(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_passwordManagementDrift(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withManagedPassword(rName, "testuser"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
				},
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)
					input := &redshiftserverless.UpdateNamespaceInput{
						NamespaceName:       aws.String(rName),
						ManageAdminPassword: aws.Bool(false),
						AdminUsername:       aws.String("differentuser"),
						AdminUserPassword:   aws.String("Password124"),
					}
					_, err := conn.UpdateNamespace(ctx, input)
					if err != nil {
						t.Fatalf("Error introducing drift: %s", err)
					}
					// wait for the secret to be removed to avoid error on subsequent namespace update
					time.Sleep(30 * time.Second)
				},
				Config: testAccNamespaceConfig_withManagedPassword(rName, "testuser"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("admin_username"), knownvalue.StringExact("testuser")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("admin_user_password"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("manage_admin_password"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func TestAccRedshiftServerlessNamespace_adminUsernameDriftWithManagedPassword(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_withManagedPassword(rName, "testuser"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
				},
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)
					input := &redshiftserverless.UpdateNamespaceInput{
						NamespaceName:       aws.String(rName),
						ManageAdminPassword: aws.Bool(true),
						AdminUsername:       aws.String("differentuser"),
					}
					_, err := conn.UpdateNamespace(ctx, input)
					if err != nil {
						t.Fatalf("Error introducing drift: %s", err)
					}
					// repeat the API call to workaround current bug in the API where the secret is not created
					// upon the first update. Once the bug is fixed, this code won't be necessary but also won't hurt
					// as UpdateNamespace will then be idempotent.
					_, err = conn.UpdateNamespace(ctx, input)
					if err != nil {
						t.Fatalf("Error introducing drift: %s", err)
					}
					// wait for the secret to be removed to avoid error on subsequent namespace update
					time.Sleep(30 * time.Second)
				},
				Config: testAccNamespaceConfig_withManagedPassword(rName, "testuser"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("admin_username"), knownvalue.StringExact("testuser")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("admin_user_password"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("manage_admin_password"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func testAccCheckNamespaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshiftserverless_namespace" {
				continue
			}

			_, err := tfredshiftserverless.FindNamespaceByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Serverless Namespace %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNamespaceExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift Serverless Namespace is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)

		_, err := tfredshiftserverless.FindNamespaceByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccNamespaceConfig_baseIAMRole(rName string, n int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  count = %[2]d

  name = "%[1]s-${count.index}"
  path = "/service-role/"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": [
                    "redshift-serverless.amazonaws.com",
                    "redshift.amazonaws.com",
                    "sagemaker.amazonaws.com"
                ]
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test" {
  count = %[2]d

  role       = aws_iam_role.test[count.index].name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonRedshiftAllCommandsFullAccess"
}

`, rName, n)
}

func testAccNamespaceConfig_withLogExportsOutput(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

output "log_exports" {
  value = aws_redshiftserverless_namespace.test.log_exports
}
`, rName)
}

func testAccNamespaceConfig_withNameOnly(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

`, rName)
}

func testAccNamespaceConfig_withDefaultUserAndUnmanagedPassword(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name      = %[1]q
  admin_user_password = "Password123"
}
`, rName)
}

func testAccNamespaceConfig_withUnmanagedPasswordAndManageFlag(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name        = %[1]q
  admin_username        = "testuser"
  admin_user_password   = "Password123"
  manage_admin_password = false
}
`, rName)
}

func testAccNamespaceConfig_withUnmanagedPassword(rName, userName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name      = %[1]q
  admin_username      = %[2]q
  admin_user_password = "Password123"
}
`, rName, userName)
}

func testAccNamespaceConfig_withMultipleAttrsUpdated(rName string) string {
	return acctest.ConfigCompose(testAccNamespaceConfig_baseIAMRole(rName, 2), fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
  iam_roles      = aws_iam_role.test[*].arn
  log_exports    = ["userlog"]
}
`, rName))
}

func testAccNamespaceConfig_userPasswordWriteOnly(rName, password string, passwordVersion int) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name                 = %[1]q
  admin_username                 = "admin"
  admin_user_password_wo         = %[2]q
  admin_user_password_wo_version = %[3]d
}
`, rName, password, passwordVersion)
}

func testAccNamespaceConfig_withTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccNamespaceConfig_withTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccNamespaceConfig_withDefaultIAMRole(rName string) string {
	return acctest.ConfigCompose(testAccNamespaceConfig_baseIAMRole(rName, 1), fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name       = %[1]q
  default_iam_role_arn = aws_iam_role.test[0].arn
  iam_roles            = aws_iam_role.test[*].arn
}
`, rName))
}

func testAccNamespaceConfig_withoutDefaultIAMRole(rName string) string {
	return acctest.ConfigCompose(testAccNamespaceConfig_baseIAMRole(rName, 1), fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
  iam_roles      = aws_iam_role.test[*].arn
}
`, rName))
}

func testAccNamespaceConfig_withManagedPassword(rName, userName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name        = %[1]q
  admin_username        = %[2]q
  manage_admin_password = true
}
`, rName, userName)
}

func testAccNamespaceConfig_withWorkgroup(rName string) string {
	return acctest.ConfigCompose(testAccNamespaceConfig_baseIAMRole(rName, 2), fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name       = %[1]q
  default_iam_role_arn = aws_iam_role.test[0].arn
  iam_roles            = aws_iam_role.test[*].arn
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

output "db_name" {
  value = aws_redshiftserverless_namespace.test.db_name
}
`, rName))
}

func testAccNamespaceConfig_withKMSKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
  kms_key_id     = aws_kms_key.test.arn
}
`, rName)
}
