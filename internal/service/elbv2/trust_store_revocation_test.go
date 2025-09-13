// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2TrustStoreRevocation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribeTrustStoreRevocation
	resourceName := "aws_lb_trust_store_revocation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreRevocationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreRevocationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreRevocationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "trust_store_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "revocation_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func testAccCheckTrustStoreRevocationExists(ctx context.Context, n string, v *awstypes.DescribeTrustStoreRevocation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		trustStoreARN := rs.Primary.Attributes["trust_store_arn"]
		revocationID, err := strconv.ParseInt(rs.Primary.Attributes["revocation_id"], 10, 64)

		if err != nil {
			return err
		}

		output, err := tfelbv2.FindTrustStoreRevocationByTwoPartKey(ctx, conn, trustStoreARN, revocationID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTrustStoreRevocationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lb_trust_store_revocation" {
				continue
			}

			trustStoreARN := rs.Primary.Attributes["trust_store_arn"]
			revocationID, err := strconv.ParseInt(rs.Primary.Attributes["revocation_id"], 10, 64)

			if err != nil {
				return err
			}

			_, err = tfelbv2.FindTrustStoreRevocationByTwoPartKey(ctx, conn, trustStoreARN, revocationID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELBv2 Trust Store Revocation %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTrustStoreRevocationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTrustStoreConfig_baseS3BucketCA(rName), fmt.Sprintf(`
resource "aws_lb_trust_store" "test" {
  name                             = %[1]q
  ca_certificates_bundle_s3_bucket = aws_s3_bucket.test.bucket
  ca_certificates_bundle_s3_key    = aws_s3_object.test.key
}

resource "aws_s3_object" "crl" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "%[1]s-crl.pem"
  content = <<EOT
-----BEGIN X509 CRL-----
MIIC/jCB5wIBATANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMCR0IxFzAVBgNV
BAgMDldlc3QgWW9ya3NoaXJlMQ4wDAYDVQQHDAVMZWVkczEZMBcGA1UECgwQU0VM
Ri1TSUdORUQtUk9PVDENMAsGA1UECwwETUVTSDEhMB8GA1UEAwwYc2VydmVyLXJv
b3QtY2EgLSByb290IENBFw0yMzA1MDMyMDE2MzJaFw0zMzA0MzAyMDE2MzJaoC8w
LTAfBgNVHSMEGDAWgBQbSz4WFRLmobqssr+mJO5suezVUTAKBgNVHRQEAwIBATAN
BgkqhkiG9w0BAQsFAAOCAgEAICif1em35UW2dYc6gxy8qbEqGgRKxWWaRzvfHpFK
3mkilV/bIXHqqoeaKFvijmPndVBd2TRFKKUZfNmcwUOISF9EJB5e9bx+J0yLv2ab
ovcES1P16R6k84IaIELcHu3Oib3ob0+KQulPbLR4uUvm1sabcj5dweYbgz7wdqWp
FAcDqgwYx9I7gwIcflEUAKx3mSJ426/cMW/yYTDr4Jgdr+GFIGwCJK9ggyo0CXOT
y3ZqM1yHWbQoe8K++La1ZGM+JOI2/8qta67BUx9jovNZMIsqVUMhTLMfVGsZRnon
3EnF0RP/eNr7Q1ajieOcxqbB8/XH5JsVpDUPMEj5DAht/h/CpsIXF6tcmNvzv5HM
NNKNnCNO6tQKCAF0S/BFHz+P4SW1oUId0dcxo2dIMBmdAy/mEano30JdtY1ZfIKA
ihAxK05gplnp1QQgyThoj7D3u8LTQSzo5V9rPX65CQCCK9RhaO00VEHfZmrzuWEV
W0OQgeAWPNFi/bZ/SqMln6CO9J6U60e/rvwxRHkkMAS7cR09XVnXm2sPAEDnXg52
gas9OVAJsw3d6UlMtC8cCJe0MPYHsySKaezK92mDOTQmTpsntHbzvGEF4VXs2rWK
mblwMFiUDFIa5K9gMRKksXpzRHvOvDe4+ZJvop1k7r5tU4iAYZkNgTGjiMt3WjwD
4wE=
-----END X509 CRL-----
EOT
}

resource "aws_lb_trust_store_revocation" "test" {
  trust_store_arn = aws_lb_trust_store.test.arn

  revocations_s3_bucket = aws_s3_bucket.test.bucket
  revocations_s3_key    = aws_s3_object.crl.key
}
`, rName))
}
