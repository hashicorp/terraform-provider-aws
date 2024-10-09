// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedAccessTrustProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessTrustProvider
	resourceName := "aws_verifiedaccess_trust_provider.test"

	trustProviderType := "user"
	userTrustProviderType := "iam-identity-center"
	description := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckVerifiedAccess(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/sso.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessTrustProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderConfig_basic("test", trustProviderType, userTrustProviderType, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "policy_reference_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "trust_provider_type", trustProviderType),
					resource.TestCheckResourceAttr(resourceName, "user_trust_provider_type", userTrustProviderType),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccVerifiedAccessTrustProvider_deviceOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessTrustProvider
	resourceName := "aws_verifiedaccess_trust_provider.test"

	trustProviderType := "device"
	deviceTrustProviderType := "jamf"
	tenantId := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessTrustProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderConfig_deviceOptions("test", trustProviderType, deviceTrustProviderType, tenantId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "device_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "device_options.0.tenant_id", tenantId),
					resource.TestCheckResourceAttr(resourceName, "device_trust_provider_type", deviceTrustProviderType),
					resource.TestCheckResourceAttr(resourceName, "policy_reference_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "trust_provider_type", trustProviderType),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccVerifiedAccessTrustProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessTrustProvider
	resourceName := "aws_verifiedaccess_trust_provider.test"

	trustProviderType := "user"
	userTrustProviderType := "iam-identity-center"
	description := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/sso.amazonaws.com")
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessTrustProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderConfig_basic("test", trustProviderType, userTrustProviderType, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVerifiedAccessTrustProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVerifiedAccessTrustProvider_oidcOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessTrustProvider
	resourceName := "aws_verifiedaccess_trust_provider.test"

	trustProviderType := "user"
	userTrustProviderType := "oidc"
	authorizationEndpoint := "https://authorization.example.com"
	clientId := sdkacctest.RandString(10)
	clientSecret := sdkacctest.RandString(10)
	issuer := "https://issuer.example.com"
	scope := sdkacctest.RandString(10)
	tokenEndpoint := "https://token.example.com"
	userInfoEndpoint := "https://user.example.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessTrustProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderConfig_oidcOptions("test", trustProviderType, userTrustProviderType, authorizationEndpoint, clientId, clientSecret, issuer, scope, tokenEndpoint, userInfoEndpoint),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.authorization_endpoint", authorizationEndpoint),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.client_id", clientId),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.client_secret", clientSecret),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.issuer", issuer),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.scope", scope),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.token_endpoint", tokenEndpoint),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.user_info_endpoint", userInfoEndpoint),
					resource.TestCheckResourceAttr(resourceName, "policy_reference_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "trust_provider_type", trustProviderType),
					resource.TestCheckResourceAttr(resourceName, "user_trust_provider_type", userTrustProviderType),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oidc_options.0.client_secret"},
			},
		},
	})
}

func TestAccVerifiedAccessTrustProvider_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessTrustProvider
	resourceName := "aws_verifiedaccess_trust_provider.test"

	trustProviderType := "user"
	userTrustProviderType := "iam-identity-center"
	description := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/sso.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessTrustProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderConfig_tags1("test", trustProviderType, userTrustProviderType, description, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccVerifiedAccessTrustProviderConfig_tags2("test", trustProviderType, userTrustProviderType, description, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVerifiedAccessTrustProviderConfig_tags1("test", trustProviderType, userTrustProviderType, description, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCheckVerifiedAccessTrustProviderExists(ctx context.Context, n string, v *types.VerifiedAccessTrustProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVerifiedAccessTrustProviderByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVerifiedAccessTrustProviderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedaccess_trust_provider" {
				continue
			}

			_, err := tfec2.FindVerifiedAccessTrustProviderByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Verified Access Trust Provider %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckVerifiedAccess(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVerifiedAccessTrustProvidersInput{}
	_, err := conn.DescribeVerifiedAccessTrustProviders(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVerifiedAccessTrustProviderConfig_basic(policyReferenceName, trustProviderType, userTrustProviderType, description string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_trust_provider" "test" {
  description              = %[4]q
  policy_reference_name    = %[1]q
  trust_provider_type      = %[2]q
  user_trust_provider_type = %[3]q
}
`, policyReferenceName, trustProviderType, userTrustProviderType, description)
}

func testAccVerifiedAccessTrustProviderConfig_deviceOptions(policyReferenceName, trustProviderType, deviceTrustProviderType, tenantId string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_trust_provider" "test" {
  device_options {
    tenant_id = %[4]q
  }
  device_trust_provider_type = %[3]q
  policy_reference_name      = %[1]q
  trust_provider_type        = %[2]q
}
`, policyReferenceName, trustProviderType, deviceTrustProviderType, tenantId)
}

func testAccVerifiedAccessTrustProviderConfig_oidcOptions(policyReferenceName, trustProviderType, userTrustProviderType, authorizationEndpoint, clientId, clientSecret, issuer, scope, tokenEndpoint, userInfoEndpoint string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_trust_provider" "test" {
  oidc_options {
    authorization_endpoint = %[4]q
    client_id              = %[5]q
    client_secret          = %[6]q
    issuer                 = %[7]q
    scope                  = %[8]q
    token_endpoint         = %[9]q
    user_info_endpoint     = %[10]q
  }
  policy_reference_name    = %[1]q
  trust_provider_type      = %[2]q
  user_trust_provider_type = %[3]q
}
`, policyReferenceName, trustProviderType, userTrustProviderType, authorizationEndpoint, clientId, clientSecret, issuer, scope, tokenEndpoint, userInfoEndpoint)
}

func testAccVerifiedAccessTrustProviderConfig_tags1(policyReferenceName, trustProviderType, userTrustProviderType, description, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_trust_provider" "test" {
  description              = %[4]q
  policy_reference_name    = %[1]q
  trust_provider_type      = %[2]q
  user_trust_provider_type = %[3]q
  tags = {
    %[5]q = %[6]q
  }
}
`, policyReferenceName, trustProviderType, userTrustProviderType, description, tagKey1, tagValue1)
}

func testAccVerifiedAccessTrustProviderConfig_tags2(policyReferenceName, trustProviderType, userTrustProviderType, description, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_trust_provider" "test" {
  description              = %[4]q
  policy_reference_name    = %[1]q
  trust_provider_type      = %[2]q
  user_trust_provider_type = %[3]q
  tags = {
    %[5]q = %[6]q
    %[7]q = %[8]q
  }
}
`, policyReferenceName, trustProviderType, userTrustProviderType, description, tagKey1, tagValue1, tagKey2, tagValue2)
}
