// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acm_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccACMCertificateDataSource_byDomain(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byDomain(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainNoMatch(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateDataSourceConfig_byDomainNoMatch(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				ExpectError: regexache.MustCompile(`reading ACM Certificates: empty result`),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainMultiple(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateDataSourceConfig_byDomainMultiple(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				ExpectError: regexache.MustCompile(`2 matching ACM Certificates found`),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainAndTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byDomainAndTags(domain, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainAndStatuses(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byDomainAndStatuses(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainAndKeyTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 4096)
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byDomainAndKeyTypes(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainAndTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byDomainAndTypes(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainAndTypesNoMatch(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateDataSourceConfig_byDomainAndTypesNoMatch(domain, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				ExpectError: regexache.MustCompile(`reading ACM Certificates: empty result`),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byDomainAndKeyTypesMostRecent(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test3"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 4096)
	domain := acctest.RandomDomain().String()

	// Create 3 certificates, a minute apart.
	certificate1 := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)
	time.Sleep(1 * time.Minute)
	certificate2 := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)
	time.Sleep(1 * time.Minute)
	certificate3 := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccCertificateDataSourceConfig_importedCertificate(acctest.TLSPEMEscapeNewlines(certificate1), acctest.TLSPEMEscapeNewlines(key), 1),
					testAccCertificateDataSourceConfig_importedCertificate(acctest.TLSPEMEscapeNewlines(certificate2), acctest.TLSPEMEscapeNewlines(key), 2),
					testAccCertificateDataSourceConfig_importedCertificate(acctest.TLSPEMEscapeNewlines(certificate3), acctest.TLSPEMEscapeNewlines(key), 3),
				),
				Check: resource.ComposeTestCheckFunc(
					// Sleep an additional minute after resource creation.
					acctest.CheckSleep(t, 1*time.Minute),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccCertificateDataSourceConfig_importedCertificate(acctest.TLSPEMEscapeNewlines(certificate1), acctest.TLSPEMEscapeNewlines(key), 1),
					testAccCertificateDataSourceConfig_importedCertificate(acctest.TLSPEMEscapeNewlines(certificate2), acctest.TLSPEMEscapeNewlines(key), 2),
					testAccCertificateDataSourceConfig_importedCertificate(acctest.TLSPEMEscapeNewlines(certificate3), acctest.TLSPEMEscapeNewlines(key), 3),
					testAccCertificateDataSourceConfig_byDomainAndKeyTypesMostRecent(domain),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byTags(rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byTagsNoMatch(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048) // ListCertificates: Default filtering returns only RSA_2048 certificates.
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomDomain().String())
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateDataSourceConfig_byTagsNoMatch(rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				ExpectError: regexache.MustCompile(`no matching ACM Certificate found`),
			},
		},
	})
}

func TestAccACMCertificateDataSource_byTagsAndKeyTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 4096)
	domain := acctest.RandomDomain().String()
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_byTagsAndKeyTypes(rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDomain, domain),
				),
			},
		},
	})
}

func testAccCertificateDataSourceConfig_byDomain(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain = %[1]q

  depends_on = [aws_acm_certificate.test]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainNoMatch(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain = "not.%[1]s"

  depends_on = [aws_acm_certificate.test]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainMultiple(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test1" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_acm_certificate" "test2" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain = %[1]q

  depends_on = [aws_acm_certificate.test1, aws_acm_certificate.test2]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainAndTags(domain, rName, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[3]s"
  private_key      = "%[4]s"

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
    Name = %[2]q
  }
}

data "aws_acm_certificate" "test" {
  domain = %[1]q

  tags = {
    Key1 = "Value1"
    Name = aws_acm_certificate.test.tags["Name"]
  }
}
`, domain, rName, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainAndStatuses(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain   = %[1]q
  statuses = ["EXPIRED", "ISSUED"]

  depends_on = [aws_acm_certificate.test]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainAndKeyTypes(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain    = %[1]q
  key_types = ["RSA_4096"]

  depends_on = [aws_acm_certificate.test]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_importedCertificate(certificate, key string, i int) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test%[1]d" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, i, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainAndKeyTypesMostRecent(domain string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = %[1]q
  key_types   = ["RSA_4096"]
  most_recent = true
}
`, domain)
}

func testAccCertificateDataSourceConfig_byDomainAndTypes(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain = %[1]q
  types  = ["IMPORTED", "PRIVATE"]

  depends_on = [aws_acm_certificate.test]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_byDomainAndTypesNoMatch(domain, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_acm_certificate" "test" {
  domain = %[1]q
  types  = ["AMAZON_ISSUED"]

  depends_on = [aws_acm_certificate.test]
}
`, domain, certificate, key)
}

func testAccCertificateDataSourceConfig_byTags(rName, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
    Name = %[1]q
  }
}

data "aws_acm_certificate" "test" {
  tags = {
    Key1 = "Value1"
    Name = aws_acm_certificate.test.tags["Name"]
  }
}
`, rName, certificate, key)
}

func testAccCertificateDataSourceConfig_byTagsNoMatch(rName, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
    Name = %[1]q
  }
}

data "aws_acm_certificate" "test" {
  tags = {
    Key1 = "Value1"
    Key3 = "Value3"
    Name = aws_acm_certificate.test.tags["Name"]
  }
}
`, rName, certificate, key)
}

func testAccCertificateDataSourceConfig_byTagsAndKeyTypes(rName, certificate, key string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  tags = {
    Name = %[1]q
  }
}

data "aws_acm_certificate" "test" {
  key_types = ["RSA_4096"]

  tags = {
    Name = aws_acm_certificate.test.tags["Name"]
  }
}
`, rName, certificate, key)
}
