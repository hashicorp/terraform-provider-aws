// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package signer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/aws-sdk-go-v2/service/signer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsigner "github.com/hashicorp/terraform-provider-aws/internal/service/signer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSignerSigningProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf signer.GetSigningProfileOutput
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	resourceName := "aws_signer_signing_profile.test_sp"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttrSet(resourceName, "platform_display_name"),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
					resource.TestCheckResourceAttr(resourceName, "revocation_record.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "signing_material"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	resourceName := "aws_signer_signing_profile.test_sp"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsigner.ResourceSigningProfile(), resourceName),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_namePrefix("tf_acc_test_prefix_"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
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
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	resourceName := "aws_signer_signing_profile.test_sp"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSigningProfileConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSignerSigningProfile_signatureValidityPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	var conf signer.GetSigningProfileOutput
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	resourceName := "aws_signer_signing_profile.test_sp"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_svp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.type", "DAYS"),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.value", acctest.Ct10),
				),
			},
		},
	})
}

func testAccPreCheckSingerSigningProfile(ctx context.Context, t *testing.T, platformID string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SignerClient(ctx)

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

func testAccCheckSigningProfileExists(ctx context.Context, n string, v *signer.GetSigningProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SignerClient(ctx)

		output, err := tfsigner.FindSigningProfileByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSigningProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SignerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_signer_signing_profile" {
				continue
			}

			_, err := tfsigner.FindSigningProfileByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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
