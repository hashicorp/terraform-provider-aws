// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDomainName_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName awstypes.DomainNameConfig
	appsyncCertDomain := acctest.SkipIfEnvVarNotSet(t, "AWS_APPSYNC_DOMAIN_NAME_CERTIFICATE_DOMAIN")
	rName := sdkacctest.RandString(8)
	acmCertificateResourceName := "data.aws_acm_certificate.test"
	resourceName := "aws_appsync_domain_name.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_basic(rName, appsyncCertDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, acmCertificateResourceName, names.AttrARN),
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

func testAccDomainName_description(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName awstypes.DomainNameConfig
	appsyncCertDomain := acctest.SkipIfEnvVarNotSet(t, "AWS_APPSYNC_DOMAIN_NAME_CERTIFICATE_DOMAIN")
	rName := sdkacctest.RandString(8)
	resourceName := "aws_appsync_domain_name.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_description(rName, appsyncCertDomain, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccDomainNameConfig_description(rName, appsyncCertDomain, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
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

func testAccDomainName_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName awstypes.DomainNameConfig
	appsyncCertDomain := acctest.SkipIfEnvVarNotSet(t, "AWS_APPSYNC_DOMAIN_NAME_CERTIFICATE_DOMAIN")
	rName := sdkacctest.RandString(8)
	resourceName := "aws_appsync_domain_name.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_basic(rName, appsyncCertDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, resourceName, &domainName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappsync.ResourceDomainName(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainNameDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_domain_name" {
				continue
			}

			_, err := tfappsync.FindDomainNameByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Appsync Domain Name %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainNameExists(ctx context.Context, n string, v *awstypes.DomainNameConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		output, err := tfappsync.FindDomainNameByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDomainNameConfig_base(domain string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = "*.%[1]s"
  most_recent = true
}
`, domain)
}

func testAccDomainNameConfig_description(rName, domain, desc string) string {
	return acctest.ConfigCompose(testAccDomainNameConfig_base(domain), fmt.Sprintf(`
resource "aws_appsync_domain_name" "test" {
  domain_name     = "%[2]s.%[1]s"
  certificate_arn = data.aws_acm_certificate.test.arn
  description     = %[3]q
}
`, domain, rName, desc))
}

func testAccDomainNameConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(testAccDomainNameConfig_base(domain), fmt.Sprintf(`
resource "aws_appsync_domain_name" "test" {
  domain_name     = "%[2]s.%[1]s"
  certificate_arn = data.aws_acm_certificate.test.arn
}
`, domain, rName))
}
