// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMSAMLProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName(t))
	idpEntityIdModified := fmt.Sprintf("https://%s", acctest.RandomDomainName(t))
	resourceName := "aws_iam_saml_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSAMLProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderConfig_basic(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "iam", fmt.Sprintf("saml-provider/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "saml_metadata_document"),
					resource.TestCheckResourceAttrSet(resourceName, "saml_provider_uuid"),
					resource.TestCheckResourceAttrSet(resourceName, "valid_until"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccSAMLProviderConfig_update(rName, idpEntityIdModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "saml_metadata_document"),
					resource.TestCheckResourceAttrSet(resourceName, "saml_provider_uuid"),
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

func TestAccIAMSAMLProvider_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName(t))
	resourceName := "aws_iam_saml_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSAMLProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderConfig_tags1(rName, idpEntityId, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
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
				Config: testAccSAMLProviderConfig_tags2(rName, idpEntityId, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSAMLProviderConfig_tags1(rName, idpEntityId, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccIAMSAMLProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName(t))
	resourceName := "aws_iam_saml_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSAMLProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderConfig_basic(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceSAMLProvider(), resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMSAMLProvider_contextProvider(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName(t))
	resourceName := "aws_iam_saml_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSAMLProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderConfig_contextProvider(rName, idpEntityId, "subject", "user"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "context_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "context_provider.0.subject_source_key", "subject"),
					resource.TestCheckResourceAttr(resourceName, "context_provider.0.subject_type_value", "user"),
					resource.TestCheckResourceAttr(resourceName, "context_provider.0.key_definition.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "context_provider.0.key_definition.*", map[string]string{
						"source_key": names.AttrEmail,
						"target_key": "Email",
					}),
					resource.TestCheckResourceAttr(resourceName, "saml_key_hashes.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "saml_key_hashes.0.issuer_url"),
					resource.TestCheckResourceAttr(resourceName, "saml_key_hashes.0.key_hashes.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSAMLProviderConfig_contextProvider(rName, idpEntityId, "nameid", "workload"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "context_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "context_provider.0.subject_source_key", "nameid"),
					resource.TestCheckResourceAttr(resourceName, "context_provider.0.subject_type_value", "workload"),
				),
			},
			{
				Config: testAccSAMLProviderConfig_basic(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "context_provider.#", "0"),
				),
			},
		},
	})
}

func TestAccIAMSAMLProvider_identityCenterInstance(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName(t))
	resourceName := "aws_iam_saml_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSAMLProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderConfig_contextProvider(rName, idpEntityId, "subject", "user"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_center_instance_arn", ""),
				),
			},
			{
				Config: testAccSAMLProviderConfig_identityCenterInstance(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "identity_center_instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttrSet(resourceName, "identity_store_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSAMLProviderConfig_identityCenterInstanceSubject(rName, idpEntityId, "nameid"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "identity_center_instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttr(resourceName, "context_provider.0.subject_source_key", "nameid"),
				),
			},
			{
				Config: testAccSAMLProviderConfig_basic(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "context_provider.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "identity_center_instance_arn", ""),
				),
			},
		},
	})
}

func TestAccIAMSAMLProvider_contextProviderKeyDefinitions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName(t))
	resourceName := "aws_iam_saml_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSAMLProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderConfig_contextProviderKeyDefinitionsThree(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "context_provider.0.key_definition.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "context_provider.0.key_definition.*", map[string]string{
						"source_key": names.AttrEmail,
						"target_key": names.AttrEmail, // Defaults to the source key when omitted.
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "context_provider.0.key_definition.*", map[string]string{
						"source_key": "department",
						"target_key": "Department",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "context_provider.0.key_definition.*", map[string]string{
						"source_key": "groups",
						"target_key": "Groups",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSAMLProviderConfig_contextProviderKeyDefinitionsTwo(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "context_provider.0.key_definition.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "context_provider.0.key_definition.*", map[string]string{
						"source_key": "title",
						"target_key": "Title",
					}),
				),
			},
		},
	})
}

func testAccCheckSAMLProviderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_saml_provider" {
				continue
			}

			_, err := tfiam.FindSAMLProviderByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM SAML Provider %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSAMLProviderExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		_, err := tfiam.FindSAMLProviderByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccSAMLProviderConfig_basic(rName, idpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}
`, rName, idpEntityId)
}

func testAccSAMLProviderConfig_update(rName, idpEntityIdModified string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata-modified.xml.tpl", { entity_id_modified = %[2]q })
}
`, rName, idpEntityIdModified)
}

func testAccSAMLProviderConfig_tags1(rName, idpEntityId, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, idpEntityId, tagKey1, tagValue1)
}

func testAccSAMLProviderConfig_tags2(rName, idpEntityId, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, idpEntityId, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccSAMLProviderConfig_contextProvider(rName, idpEntityId, subjectSourceKey, subjectTypeValue string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  context_provider {
    subject_source_key = %[3]q
    subject_type_value = %[4]q

    key_definition {
      source_key = "email"
      target_key = "Email"
    }
  }
}
`, rName, idpEntityId, subjectSourceKey, subjectTypeValue)
}

func testAccSAMLProviderConfig_contextProviderKeyDefinitionsThree(rName, idpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  context_provider {
    subject_source_key = "subject"
    subject_type_value = "user"

    key_definition {
      source_key = "email"
    }

    key_definition {
      source_key = "department"
      target_key = "Department"
    }

    key_definition {
      source_key = "groups"
      target_key = "Groups"
    }
  }
}
`, rName, idpEntityId)
}

func testAccSAMLProviderConfig_contextProviderKeyDefinitionsTwo(rName, idpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  context_provider {
    subject_source_key = "subject"
    subject_type_value = "user"

    key_definition {
      source_key = "groups"
      target_key = "Groups"
    }

    key_definition {
      source_key = "title"
      target_key = "Title"
    }
  }
}
`, rName, idpEntityId)
}

func testAccSAMLProviderConfig_identityCenterInstanceSubject(rName, idpEntityId, subjectSourceKey string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  context_provider {
    subject_source_key = %[3]q
    subject_type_value = "user"
  }
}
`, rName, idpEntityId, subjectSourceKey)
}

func testAccSAMLProviderConfig_identityCenterInstance(rName, idpEntityId string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  context_provider {
    subject_source_key = "subject"
    subject_type_value = "user"
  }
}
`, rName, idpEntityId)
}
