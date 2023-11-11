// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
)

func testAccDomainName_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName appsync.DomainNameConfig
	appsyncCertDomain := getCertDomain(t)

	rName := sdkacctest.RandString(8)
	acmCertificateResourceName := "data.aws_acm_certificate.test"
	resourceName := "aws_appsync_domain_name.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appsync.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_basic(rName, appsyncCertDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", acmCertificateResourceName, "arn"),
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
	var domainName appsync.DomainNameConfig
	appsyncCertDomain := getCertDomain(t)

	rName := sdkacctest.RandString(8)
	resourceName := "aws_appsync_domain_name.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appsync.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_description(rName, appsyncCertDomain, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccDomainNameConfig_description(rName, appsyncCertDomain, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
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
	var domainName appsync.DomainNameConfig
	appsyncCertDomain := getCertDomain(t)

	rName := sdkacctest.RandString(8)
	resourceName := "aws_appsync_domain_name.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appsync.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_domain_name" {
				continue
			}

			domainName, err := tfappsync.FindDomainNameByID(ctx, conn, rs.Primary.ID)
			if err == nil {
				if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
					return nil
				}
				return err
			}

			if domainName != nil && aws.StringValue(domainName.DomainName) == rs.Primary.ID {
				return fmt.Errorf("Appsync Domain Name ID %q still exists", rs.Primary.ID)
			}

			return nil
		}
		return nil
	}
}

func testAccCheckDomainNameExists(ctx context.Context, resourceName string, domainName *appsync.DomainNameConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Appsync Domain Name Not found in state: %s", resourceName)
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn(ctx)

		domain, err := tfappsync.FindDomainNameByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if domain == nil || domain.DomainName == nil {
			return fmt.Errorf("Appsync Domain Name %q not found", rs.Primary.ID)
		}

		*domainName = *domain

		return nil
	}
}

func testAccDomainNameBaseConfig(domain string) string {
	return acctest.ConfigAlternateRegionProvider() + fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  provider    = "awsalternate"
  domain      = "*.%[1]s"
  most_recent = true
}
`, domain)
}

func testAccDomainNameConfig_description(rName, domain, desc string) string {
	return testAccDomainNameBaseConfig(domain) + fmt.Sprintf(`
resource "aws_appsync_domain_name" "test" {
  domain_name     = "%[2]s.%[1]s"
  certificate_arn = data.aws_acm_certificate.test.arn
  description     = %[3]q
}
`, domain, rName, desc)
}

func testAccDomainNameConfig_basic(rName, domain string) string {
	return testAccDomainNameBaseConfig(domain) + fmt.Sprintf(`
resource "aws_appsync_domain_name" "test" {
  domain_name     = "%[2]s.%[1]s"
  certificate_arn = data.aws_acm_certificate.test.arn
}
`, domain, rName)
}
