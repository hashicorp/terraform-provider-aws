// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/amplify/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDomainAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var domain types.DomainAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_basic(rName, domainName, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "amplify", regexache.MustCompile(`apps/.+/domains/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_sub_domain", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "sub_domain.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name":    rName,
						names.AttrPrefix: "",
					}),
					resource.TestCheckResourceAttr(resourceName, "wait_for_verification", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
		},
	})
}

func testAccDomainAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var domain types.DomainAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_basic(rName, domainName, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfamplify.ResourceDomainAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDomainAssociation_update(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var domain types.DomainAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_basic(rName, domainName, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "amplify", regexache.MustCompile(`apps/.+/domains/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_sub_domain", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "sub_domain.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name":    rName,
						names.AttrPrefix: "",
					}),
					resource.TestCheckResourceAttr(resourceName, "wait_for_verification", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
			{
				Config: testAccDomainAssociationConfig_updated(rName, domainName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "amplify", regexache.MustCompile(`apps/.+/domains/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_sub_domain", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "sub_domain.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name":    rName,
						names.AttrPrefix: "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name":    fmt.Sprintf("%s-2", rName),
						names.AttrPrefix: "www",
					}),
					resource.TestCheckResourceAttr(resourceName, "wait_for_verification", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccDomainAssociation_certificateSettings(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var domain types.DomainAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_certificateSettings(rName, domainName, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "amplify", regexache.MustCompile(`apps/.+/domains/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.type", "AMPLIFY_MANAGED"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_settings.0.certificate_verification_dns_record"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
		},
	})
}

func testAccCheckDomainAssociationExists(ctx context.Context, n string, v *types.DomainAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyClient(ctx)

		output, err := tfamplify.FindDomainAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_id"], rs.Primary.Attributes[names.AttrDomainName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDomainAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_amplify_domain_association" {
				continue
			}

			_, err := tfamplify.FindDomainAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_id"], rs.Primary.Attributes[names.AttrDomainName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Amplify Domain Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDomainAssociationConfig_basic(rName, domainName string, enableAutoSubDomain bool, waitForVerification bool) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = %[2]q

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  enable_auto_sub_domain = %[3]t
  wait_for_verification  = %[4]t
}
`, rName, domainName, enableAutoSubDomain, waitForVerification)
}

func testAccDomainAssociationConfig_updated(rName, domainName string, enableAutoSubDomain bool, waitForVerification bool) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

resource "aws_amplify_branch" "test2" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%[1]s-2"
}

resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = %[2]q

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  sub_domain {
    branch_name = aws_amplify_branch.test2.branch_name
    prefix      = "www"
  }

  enable_auto_sub_domain = %[3]t
  wait_for_verification  = %[4]t
}
`, rName, domainName, enableAutoSubDomain, waitForVerification)
}

func testAccDomainAssociationConfig_certificateSettings(rName, domainName string, enableAutoSubDomain bool, waitForVerification bool) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = %[2]q

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  certificate_settings {
    type = "AMPLIFY_MANAGED"
  }

  enable_auto_sub_domain = %[3]t
  wait_for_verification  = %[4]t
}
`, rName, domainName, enableAutoSubDomain, waitForVerification)
}
