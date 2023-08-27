package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccVerifiedAccessTrustProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var verifiedaccesstrustprovider ec2.DescribeVerifiedAccessTrustProvidersOutput
	resourceName := "aws_verifiedaccess_trust_provider.test"
	policyReferenceName := "test"
	trustProviderType := "user"
	userTrustProviderType := "iam-identity-center"
	description := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/sso.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessTrustProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderConfig_basic(policyReferenceName, trustProviderType, userTrustProviderType, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &verifiedaccesstrustprovider),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "policy_reference_name", policyReferenceName),
					resource.TestCheckResourceAttr(resourceName, "trust_provider_type", trustProviderType),
					resource.TestCheckResourceAttr(resourceName, "user_trust_provider_type", userTrustProviderType),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccVerifiedAccessTrustProvider_deviceOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var verifiedaccesstrustprovider ec2.DescribeVerifiedAccessTrustProvidersOutput
	resourceName := "aws_verifiedaccess_trust_provider.test"
	policyReferenceName := "test"
	trustProviderType := "device"
	deviceTrustProviderType := "jamf"
	tenantId := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessTrustProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderConfig_deviceOptions(policyReferenceName, trustProviderType, deviceTrustProviderType, tenantId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &verifiedaccesstrustprovider),
					resource.TestCheckResourceAttr(resourceName, "device_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "device_options.0.tenant_id", tenantId),
					resource.TestCheckResourceAttr(resourceName, "device_trust_provider_type", deviceTrustProviderType),
					resource.TestCheckResourceAttr(resourceName, "policy_reference_name", policyReferenceName),
					resource.TestCheckResourceAttr(resourceName, "trust_provider_type", trustProviderType),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccVerifiedAccessTrustProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var verifiedaccesstrustprovider ec2.DescribeVerifiedAccessTrustProvidersOutput
	resourceName := "aws_verifiedaccess_trust_provider.test"
	policyReferenceName := "test"
	trustProviderType := "user"
	userTrustProviderType := "iam-identity-center"
	description := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/sso.amazonaws.com")
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessTrustProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderConfig_basic(policyReferenceName, trustProviderType, userTrustProviderType, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &verifiedaccesstrustprovider),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVerifiedaccessTrustProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVerifiedAccessTrustProviderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedaccess_trust_provider" {
				continue
			}

			_, err := tfec2.FindVerifiedaccessTrustProviderByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return nil
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVerifiedAccessTrustProvider, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func TestAccVerifiedAccessTrustProvider_oidcOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var verifiedaccesstrustprovider ec2.DescribeVerifiedAccessTrustProvidersOutput
	resourceName := "aws_verifiedaccess_trust_provider.test"
	policyReferenceName := "test"
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
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessTrustProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderConfig_oidcOptions(policyReferenceName, trustProviderType, userTrustProviderType, authorizationEndpoint, clientId, clientSecret, issuer, scope, tokenEndpoint, userInfoEndpoint),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &verifiedaccesstrustprovider),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.authorization_endpoint", authorizationEndpoint),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.client_id", clientId),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.client_secret", clientSecret),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.issuer", issuer),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.scope", scope),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.token_endpoint", tokenEndpoint),
					resource.TestCheckResourceAttr(resourceName, "oidc_options.0.user_info_endpoint", userInfoEndpoint),
					resource.TestCheckResourceAttr(resourceName, "policy_reference_name", policyReferenceName),
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
	var verifiedaccesstrustprovider ec2.DescribeVerifiedAccessTrustProvidersOutput
	resourceName := "aws_verifiedaccess_trust_provider.test"
	policyReferenceName := "test"
	trustProviderType := "user"
	userTrustProviderType := "iam-identity-center"
	description := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/sso.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessTrustProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderConfig_tags1(policyReferenceName, trustProviderType, userTrustProviderType, description, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &verifiedaccesstrustprovider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccVerifiedAccessTrustProviderConfig_tags2(policyReferenceName, trustProviderType, userTrustProviderType, description, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &verifiedaccesstrustprovider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVerifiedAccessTrustProviderConfig_tags1(policyReferenceName, trustProviderType, userTrustProviderType, description, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderExists(ctx, resourceName, &verifiedaccesstrustprovider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func testAccCheckVerifiedAccessTrustProviderExists(ctx context.Context, name string, verifiedaccesstrustprovider *ec2.DescribeVerifiedAccessTrustProvidersOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessTrustProvider, name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessTrustProvider, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		resp, err := tfec2.FindVerifiedaccessTrustProviderByID(ctx, conn, rs.Primary.ID)
		fmt.Println(err)
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessTrustProvider, rs.Primary.ID, err)
		}

		*verifiedaccesstrustprovider = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
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
