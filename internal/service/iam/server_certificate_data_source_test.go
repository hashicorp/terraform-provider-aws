// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestResourceSortByExpirationDate(t *testing.T) {
	t.Parallel()

	certs := []awstypes.ServerCertificateMetadata{
		{
			ServerCertificateName: aws.String("oldest"),
			Expiration:            aws.Time(time.Now()),
		},
		{
			ServerCertificateName: aws.String("latest"),
			Expiration:            aws.Time(time.Now().Add(3 * time.Hour)),
		},
		{
			ServerCertificateName: aws.String("in between"),
			Expiration:            aws.Time(time.Now().Add(2 * time.Hour)),
		},
	}
	sort.Sort(tfiam.CertificateByExpiration(certs))
	if aws.ToString(certs[0].ServerCertificateName) != "latest" {
		t.Fatalf("Expected first item to be %q, but was %q", "latest", *certs[0].ServerCertificateName)
	}
}

func TestAccIAMServerCertificateDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateDataSourceConfig_cert(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iam_server_certificate.test_cert", names.AttrARN),
					resource.TestCheckResourceAttrSet("data.aws_iam_server_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttrSet("data.aws_iam_server_certificate.test", names.AttrID),
					resource.TestCheckResourceAttrSet("data.aws_iam_server_certificate.test", names.AttrName),
					resource.TestCheckResourceAttrSet("data.aws_iam_server_certificate.test", names.AttrPath),
					resource.TestCheckResourceAttrSet("data.aws_iam_server_certificate.test", "upload_date"),
					resource.TestCheckResourceAttr("data.aws_iam_server_certificate.test", names.AttrCertificateChain, ""),
					resource.TestMatchResourceAttr("data.aws_iam_server_certificate.test", "certificate_body", regexache.MustCompile("^-----BEGIN CERTIFICATE-----")),
				),
			},
		},
	})
}

func TestAccIAMServerCertificateDataSource_matchNamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccServerCertificateDataSourceConfig_certMatchNamePrefix,
				ExpectError: regexache.MustCompile(`Search for AWS IAM server certificate returned no results`),
			},
		},
	})
}

func TestAccIAMServerCertificateDataSource_path(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	path := "/test-path/"
	pathPrefix := "/test-path/"

	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateDataSourceConfig_certPath(rName, path, pathPrefix, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_server_certificate.test", names.AttrPath, path),
				),
			},
		},
	})
}

func testAccServerCertificateDataSourceConfig_cert(rName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test_cert" {
  name             = "%[1]s"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_iam_server_certificate" "test" {
  name   = aws_iam_server_certificate.test_cert.name
  latest = true
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccServerCertificateDataSourceConfig_certPath(rName, path, pathPrefix, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test_cert" {
  name             = "%[1]s"
  path             = "%[2]s"
  certificate_body = "%[3]s"
  private_key      = "%[4]s"
}

data "aws_iam_server_certificate" "test" {
  name        = aws_iam_server_certificate.test_cert.name
  path_prefix = "%[5]s"
  latest      = true
}
`, rName, path, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), pathPrefix)
}

var testAccServerCertificateDataSourceConfig_certMatchNamePrefix = `
data "aws_iam_server_certificate" "test" {
  name_prefix = "MyCert"
  latest      = true
}
`
