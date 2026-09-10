// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfopensearchserverless "github.com/hashicorp/terraform-provider-aws/internal/service/opensearchserverless"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessSecurityConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var securityconfig types.SecurityConfigDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckSecurityConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfig_basic(rName, "test-fixtures/idp-metadata.xml"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "saml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "iam_federation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iam_identity_center_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "saml_options.0.session_timeout"),
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

func TestAccOpenSearchServerlessSecurityConfig_update(t *testing.T) {
	ctx := acctest.Context(t)
	var securityconfig types.SecurityConfigDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckSecurityConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfig_update(rName, "test-fixtures/idp-metadata.xml", names.AttrDescription, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "saml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout", "60"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
				),
			},
			{
				Config: testAccSecurityConfig_update(rName, "test-fixtures/idp-metadata.xml", "description updated", 40),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "saml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout", "40"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description updated"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var securityconfig types.SecurityConfigDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckSecurityConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfig_basic(rName, "test-fixtures/idp-metadata.xml"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfopensearchserverless.ResourceSecurityConfig, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityConfig_iamFederationOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var securityconfig types.SecurityConfigDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckSecurityConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfig_iamFederationOptionsWithGroupAttributeOnly(rName, "test-1"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.SecurityConfigTypeIamfederation)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "iam_federation_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "iam_federation_options.0.group_attribute", "test-1"),
					resource.TestCheckNoResourceAttr(resourceName, "iam_federation_options.0.user_attribute"),
				),
			},
			{
				Config: testAccSecurityConfig_iamFederationOptions(rName, "test-1", "test-a"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.SecurityConfigTypeIamfederation)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "iam_federation_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "iam_federation_options.0.group_attribute", "test-1"),
					resource.TestCheckResourceAttr(resourceName, "iam_federation_options.0.user_attribute", "test-a"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSecurityConfig_iamFederationOptions(rName, "test-2", "test-b"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.SecurityConfigTypeIamfederation)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "iam_federation_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "iam_federation_options.0.group_attribute", "test-2"),
					resource.TestCheckResourceAttr(resourceName, "iam_federation_options.0.user_attribute", "test-b"),
				),
			},
			{
				Config: testAccSecurityConfig_iamFederationOptionsWithGroupAttributeOnly(rName, "test-1"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.SecurityConfigTypeIamfederation)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "iam_federation_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "iam_federation_options.0.group_attribute", "test-1"),
					resource.TestCheckNoResourceAttr(resourceName, "iam_federation_options.0.user_attribute"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityConfig_iamIdentityCenterOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var securityconfig types.SecurityConfigDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
			testAccPreCheckSecurityConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfig_iamIdentityCenterOptionsWithoutGroupAndUserAttribute(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.SecurityConfigTypeIamidentitycenter)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "iam_identity_center_options.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "iam_identity_center_options.0.instance_arn"),
					resource.TestCheckResourceAttr(resourceName, "iam_identity_center_options.0.group_attribute", string(types.IamIdentityCenterGroupAttributeGroupId)),
					resource.TestCheckResourceAttr(resourceName, "iam_identity_center_options.0.user_attribute", string(types.IamIdentityCenterUserAttributeUserId)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSecurityConfig_iamIdentityCenterOptions(rName, string(types.IamIdentityCenterGroupAttributeGroupName), string(types.IamIdentityCenterUserAttributeUserName)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.SecurityConfigTypeIamidentitycenter)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "iam_identity_center_options.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "iam_identity_center_options.0.instance_arn"),
					resource.TestCheckResourceAttr(resourceName, "iam_identity_center_options.0.group_attribute", string(types.IamIdentityCenterGroupAttributeGroupName)),
					resource.TestCheckResourceAttr(resourceName, "iam_identity_center_options.0.user_attribute", string(types.IamIdentityCenterUserAttributeUserName)),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityConfig_upgradeV6_0_0(t *testing.T) {
	ctx := acctest.Context(t)
	var securityconfig types.SecurityConfigDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckSecurityConfig(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		CheckDestroy: testAccCheckSecurityConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.94.1",
					},
				},
				Config: testAccSecurityConfig_samlOptions(rName, "test-fixtures/idp-metadata.xml", 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "saml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "saml_options.session_timeout", "60"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccSecurityConfig_samlOptions(rName, "test-fixtures/idp-metadata.xml", 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "saml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout", "60"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckSecurityConfigDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_security_config" {
				continue
			}

			_, err := tfopensearchserverless.FindSecurityConfigByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingDestroyed, tfopensearchserverless.ResNameSecurityConfig, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckSecurityConfigExists(ctx context.Context, t *testing.T, name string, securityconfig *types.SecurityConfigDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityConfig, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityConfig, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)
		resp, err := tfopensearchserverless.FindSecurityConfigByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityConfig, rs.Primary.ID, err)
		}

		*securityconfig = *resp

		return nil
	}
}

func testAccPreCheckSecurityConfig(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

	input := &opensearchserverless.ListSecurityConfigsInput{
		Type: types.SecurityConfigTypeSaml,
	}
	_, err := conn.ListSecurityConfigs(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSecurityConfig_basic(rName string, samlOptions string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_config" "test" {
  name = %[1]q
  type = "saml"
  saml_options {
    metadata = file("%[2]s")
  }
}
`, rName, samlOptions)
}

func testAccSecurityConfig_update(rName, samlOptions, description string, sessionTimeout int) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_config" "test" {
  name        = %[1]q
  description = %[3]q
  type        = "saml"

  saml_options {
    metadata        = file("%[2]s")
    session_timeout = %[4]d
  }
}
`, rName, samlOptions, description, sessionTimeout)
}

func testAccSecurityConfig_samlOptions(rName, samlOptions string, sessionTimeout int) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_config" "test" {
  name = %[1]q
  type = "saml"

  saml_options {
    metadata        = file("%[2]s")
    session_timeout = %[3]d
  }
}
`, rName, samlOptions, sessionTimeout)
}

func testAccSecurityConfig_iamFederationOptions(rName, groupAttribute, userAttribute string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_opensearchserverless_security_config" "test" {
  name = %[1]q
  type = "iamfederation"

  iam_federation_options {
    group_attribute = %[2]q
    user_attribute  = %[3]q
  }
}
`, rName, groupAttribute, userAttribute)
}

func testAccSecurityConfig_iamFederationOptionsWithGroupAttributeOnly(rName, groupAttribute string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_opensearchserverless_security_config" "test" {
  name = %[1]q
  type = "iamfederation"

  iam_federation_options {
    group_attribute = %[2]q
  }
}
`, rName, groupAttribute)
}

func testAccSecurityConfig_iamIdentityCenterOptions(rName, groupAttribute, userAttribute string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_opensearchserverless_security_config" "test" {
  name = %[1]q
  type = "iamidentitycenter"

  iam_identity_center_options {
    instance_arn    = tolist(data.aws_ssoadmin_instances.test.arns)[0]
    group_attribute = %[2]q
    user_attribute  = %[3]q
  }
}
`, rName, groupAttribute, userAttribute)
}

func testAccSecurityConfig_iamIdentityCenterOptionsWithoutGroupAndUserAttribute(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_opensearchserverless_security_config" "test" {
  name = %[1]q
  type = "iamidentitycenter"

  iam_identity_center_options {
    instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  }
}
`, rName)
}
