// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package signer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/aws-sdk-go-v2/service/signer/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsigner "github.com/hashicorp/terraform-provider-aws/internal/service/signer"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSignerSigningProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf signer.GetSigningProfileOutput
	rName := fmt.Sprintf("tf_acc_test_%d", acctest.RandInt(t))
	resourceName := "aws_signer_signing_profile.test_sp"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, t, resourceName, &conf),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "signer", "/signing-profiles/{name}"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttrSet(resourceName, "platform_display_name"),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
					resource.TestCheckResourceAttr(resourceName, "revocation_record.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "signing_material"),
					resource.TestCheckResourceAttr(resourceName, "signing_parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttrSet(resourceName, "version_arn"),
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

func TestAccSignerSigningProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf signer.GetSigningProfileOutput
	rName := fmt.Sprintf("tf_acc_test_%d", acctest.RandInt(t))
	resourceName := "aws_signer_signing_profile.test_sp"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, t, resourceName, &conf),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsigner.ResourceSigningProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSignerSigningProfile_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_signer_signing_profile.test_sp"

	var conf signer.GetSigningProfileOutput

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, t, resourceName, &conf),
					acctest.CheckResourceAttrNameGeneratedWithPrefix(resourceName, names.AttrName, "terraform_"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform_"),
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

func TestAccSignerSigningProfile_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf signer.GetSigningProfileOutput
	resourceName := "aws_signer_signing_profile.test_sp"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_namePrefix("tf_acc_test_prefix_"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, t, resourceName, &conf),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf_acc_test_prefix_"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf_acc_test_prefix_"),
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

func TestAccSignerSigningProfile_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf signer.GetSigningProfileOutput
	rName := fmt.Sprintf("tf_acc_test_%d", acctest.RandInt(t))
	resourceName := "aws_signer_signing_profile.test_sp"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, t, resourceName, &conf),
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
				Config: testAccSigningProfileConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSigningProfileConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSignerSigningProfile_signatureValidityPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	var conf signer.GetSigningProfileOutput
	rName := fmt.Sprintf("tf_acc_test_%d", acctest.RandInt(t))
	resourceName := "aws_signer_signing_profile.test_sp"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_svp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.type", "DAYS"),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.value", "10"),
				),
			},
		},
	})
}

func TestAccSignerSigningProfile_signingParameters(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domainName := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var conf signer.GetSigningProfileOutput
	rName := fmt.Sprintf("tf_acc_test_%d", acctest.RandInt(t))
	resourceName := "aws_signer_signing_profile.test_sp"
	certificateName := "aws_acm_certificate.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AmazonFreeRTOS-Default")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_signingParameters(rName, rootDomain, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, t, resourceName, &conf),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "signer", "/signing-profiles/{name}"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AmazonFreeRTOS-Default"),
					resource.TestCheckResourceAttrPair(resourceName, "signing_material.0.certificate_arn", certificateName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "signing_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "signing_parameters.param1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "signing_parameters.param2", acctest.CtValue2),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttrSet(resourceName, "version_arn"),
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

func testAccPreCheckSingerSigningProfile(ctx context.Context, t *testing.T, platformID string) {
	conn := acctest.ProviderMeta(ctx, t).SignerClient(ctx)

	input := &signer.ListSigningPlatformsInput{}

	pages := signer.NewListSigningPlatformsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			t.Fatalf("unexpected PreCheck error: %s", err)
		}

		if page == nil {
			t.Skip("skipping acceptance testing: empty response")
		}

		for _, platform := range page.Platforms {
			if platform == (types.SigningPlatform{}) {
				continue
			}

			if aws.ToString(platform.PlatformId) == platformID {
				return
			}
		}
	}

	t.Skipf("skipping acceptance testing: Signing Platform (%s) not found", platformID)
}

func testAccCheckSigningProfileExists(ctx context.Context, t *testing.T, n string, v *signer.GetSigningProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SignerClient(ctx)

		output, err := tfsigner.FindSigningProfileByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSigningProfileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SignerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_signer_signing_profile" {
				continue
			}

			_, err := tfsigner.FindSigningProfileByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Signer Signing Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSigningProfileConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = %[1]q
}`, rName)
}

func testAccSigningProfileConfig_nameGenerated() string {
	return `
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}`
}

func testAccSigningProfileConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name_prefix = %[1]q
}`, namePrefix)
}

func testAccSigningProfileConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}`, rName, tagKey1, tagValue1)
}

func testAccSigningProfileConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccSigningProfileConfig_svp(rName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name_prefix = %[1]q

  signature_validity_period {
    value = 10
    type  = "DAYS"
  }
}
`, rName)
}

func testAccSigningProfileConfig_signingParameters(rName, rootDomain, domainName string) string {
	return fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = %[2]q
  private_zone = false
}

resource "aws_acm_certificate" "test" {
  domain_name       = %[3]q
  validation_method = "DNS"
}

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn]
}

resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AmazonFreeRTOS-Default"
  name        = %[1]q

  signing_material {
    certificate_arn = aws_acm_certificate.test.arn
  }
  signing_parameters = {
    "param1" = "value1"
    "param2" = "value2"
  }
  depends_on = [aws_acm_certificate_validation.test]
}`, rName, rootDomain, domainName)
}
