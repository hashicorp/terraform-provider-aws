// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspacesweb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestPortalARNFromIdentityProviderARN(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		identityProviderARN string
		wantPortalARN       string
		wantErr             bool
	}{
		"empty ARN": {
			wantErr: true,
		},
		"unparsable ARN": {
			identityProviderARN: "test",
			wantErr:             true,
		},
		"invalid ARN service": {
			identityProviderARN: "arn:aws:workspaces:us-west-2:123456789012:identityProvider/portal-123/ip-456", //lintignore:AWSAT003,AWSAT005
			wantErr:             true,
		},
		"invalid ARN resource parts": {
			identityProviderARN: "arn:aws:workspaces-web:us-west-2:123456789012:browserSettings/bs-789", //lintignore:AWSAT003,AWSAT005
			wantErr:             true,
		},
		"valid ARN": {
			identityProviderARN: "arn:aws:workspaces-web:us-west-2:123456789012:identityProvider/portal-123/ip-456", //lintignore:AWSAT003,AWSAT005
			wantPortalARN:       "arn:aws:workspaces-web:us-west-2:123456789012:portal/portal-123",                  //lintignore:AWSAT003,AWSAT005
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tfworkspacesweb.PortalARNFromIdentityProviderARN(testCase.identityProviderARN)

			if got, want := err != nil, testCase.wantErr; !cmp.Equal(got, want) {
				t.Errorf("PortalARNFromIdentityProviderARN(%s) err %t, want %t", testCase.identityProviderARN, got, want)
			}
			if err == nil {
				if diff := cmp.Diff(got, testCase.wantPortalARN); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}

func TestAccWorkSpacesWebIdentityProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var identityProvider awstypes.IdentityProvider
	resourceName := "aws_workspacesweb_identity_provider.test"
	portalResourceName := "aws_workspacesweb_portal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityProviderConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, t, resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", string(awstypes.IdentityProviderTypeSaml)),
					resource.TestCheckResourceAttrSet(resourceName, "identity_provider_details.MetadataFile"),
					resource.TestCheckResourceAttrPair(resourceName, "portal_arn", portalResourceName, "portal_arn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "identity_provider_arn", "workspaces-web", regexache.MustCompile(`identityProvider/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "identity_provider_arn"),
				ImportStateVerifyIdentifierAttribute: "identity_provider_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebIdentityProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var identityProvider awstypes.IdentityProvider
	resourceName := "aws_workspacesweb_identity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityProviderConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, t, resourceName, &identityProvider),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceIdentityProvider, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWorkSpacesWebIdentityProvider_oidc_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var identityProvider awstypes.IdentityProvider
	resourceName := "aws_workspacesweb_identity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityProviderConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, t, resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_name", "test-updated"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", string(awstypes.IdentityProviderTypeOidc)),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.client_id", "test-client-id"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.client_secret", "test-client-secret"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.oidc_issuer", "https://accounts.google.com"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.authorize_scopes", "openid, email"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.attributes_request_method", "POST"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "identity_provider_arn"),
				ImportStateVerifyIdentifierAttribute: "identity_provider_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebIdentityProvider_update(t *testing.T) {
	ctx := acctest.Context(t)
	var identityProvider awstypes.IdentityProvider
	resourceName := "aws_workspacesweb_identity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityProviderConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, t, resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", string(awstypes.IdentityProviderTypeSaml)),
					resource.TestCheckResourceAttrSet(resourceName, "identity_provider_details.MetadataFile"),
				),
			},
			{
				Config: testAccIdentityProviderConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityProviderExists(ctx, t, resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_name", "test-updated"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", string(awstypes.IdentityProviderTypeOidc)),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.client_id", "test-client-id"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.client_secret", "test-client-secret"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.oidc_issuer", "https://accounts.google.com"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.authorize_scopes", "openid, email"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.attributes_request_method", "POST"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "identity_provider_arn"),
				ImportStateVerifyIdentifierAttribute: "identity_provider_arn",
			},
		},
	})
}

func testAccCheckIdentityProviderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_identity_provider" {
				continue
			}

			_, _, err := tfworkspacesweb.FindIdentityProviderByARN(ctx, conn, rs.Primary.Attributes["identity_provider_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Web Identity Provider %s still exists", rs.Primary.Attributes["identity_provider_arn"])
		}

		return nil
	}
}

func testAccCheckIdentityProviderExists(ctx context.Context, t *testing.T, n string, v *awstypes.IdentityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, _, err := tfworkspacesweb.FindIdentityProviderByARN(ctx, conn, rs.Primary.Attributes["identity_provider_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccIdentityProviderConfig_basic() string {
	return `
resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "aws_workspacesweb_identity_provider" "test" {
  identity_provider_name = "test"
  identity_provider_type = "SAML"
  portal_arn             = aws_workspacesweb_portal.test.portal_arn

  identity_provider_details = {
    MetadataFile = file("./testfixtures/saml-metadata.xml")
  }
}
`
}

func testAccIdentityProviderConfig_updated() string {
	return `
resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "aws_workspacesweb_identity_provider" "test" {
  identity_provider_name = "test-updated"
  identity_provider_type = "OIDC"
  portal_arn             = aws_workspacesweb_portal.test.portal_arn

  identity_provider_details = {
    client_id                 = "test-client-id"
    client_secret             = "test-client-secret"
    oidc_issuer               = "https://accounts.google.com"
    attributes_request_method = "POST"
    authorize_scopes          = "openid, email"
  }
}
`
}
