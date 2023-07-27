// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package signer_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/aws-sdk-go-v2/service/signer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsigner "github.com/hashicorp/terraform-provider-aws/internal/service/signer"
)

func TestAccSignerSigningProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_signer_signing_profile.test_sp"
	rString := sdkacctest.RandString(48)
	profileName := fmt.Sprintf("tf_acc_sp_basic_%s", rString)

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
				Config: testAccSigningProfileConfig_providedName(profileName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, "name",
						regexp.MustCompile("^[a-zA-Z0-9_]{0,64}$")),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccSignerSigningProfile_generateNameWithNamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_signer_signing_profile.test_sp"
	namePrefix := "tf_acc_sp_basic_"

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
				Config: testAccSigningProfileConfig_basic(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
				),
			},
		},
	})
}

func TestAccSignerSigningProfile_generateName(t *testing.T) {
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
				Config: testAccSigningProfileConfig_generateName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
				),
			},
		},
	})
}

func TestAccSignerSigningProfile_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_signer_signing_profile.test_sp"
	namePrefix := "tf_acc_sp_basic_"

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
				Config: testAccSigningProfileConfig_tags(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "value2"),
				),
			},
			{
				Config: testAccSigningProfileConfig_updateTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "prod"),
				),
			},
		},
	})
}

func TestAccSignerSigningProfile_signatureValidityPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_signer_signing_profile.test_sp"
	namePrefix := "tf_acc_sp_basic_"

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
				Config: testAccSigningProfileConfig_svp(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.type", "DAYS"),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.value", "10"),
				),
			},
			{
				Config: testAccSigningProfileConfig_updateSVP(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.type", "MONTHS"),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.value", "10"),
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

func testAccSigningProfileConfig_basic(namePrefix string) string {
	return testAccSigningProfileBaseConfig(namePrefix)
}

func testAccSigningProfileConfig_generateName() string {
	return `
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}`
}

func testAccSigningProfileConfig_providedName(profileName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = "%s"
}`, profileName)
}

func testAccSigningProfileConfig_tags(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name_prefix = "%s"
  tags = {
    "tag1" = "value1"
    "tag2" = "value2"
  }
}`, namePrefix)
}

func testAccSigningProfileConfig_svp(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name_prefix = "%s"

  signature_validity_period {
    value = 10
    type  = "DAYS"
  }
}
`, namePrefix)
}

func testAccSigningProfileConfig_updateSVP() string {
	return `
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"

  signature_validity_period {
    value = 10
    type  = "MONTHS"
  }
}
`
}

func testAccSigningProfileConfig_updateTags() string {
	return `
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  tags = {
    "tag1" = "prod"
  }
}
`
}

func testAccSigningProfileBaseConfig(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name_prefix = "%s"
}
`, namePrefix)
}

func testAccCheckSigningProfileExists(ctx context.Context, res string, sp *signer.GetSigningProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Signing profile not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Signing Profile with that ARN does not exist")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SignerClient(ctx)

		getSp, err := tfsigner.FindSigningProfileByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*sp = *getSp

		return nil
	}
}

func testAccCheckSigningProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SignerClient(ctx)

		time.Sleep(5 * time.Second)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_signer_signing_profile" {
				continue
			}

			out, err := tfsigner.FindSigningProfileByName(ctx, conn, rs.Primary.ID)

			if out.Status != types.SigningProfileStatusCanceled && err == nil {
				return fmt.Errorf("Signing Profile not cancelled%s", *out.ProfileName)
			}
		}

		return nil
	}
}
