package signer_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSignerSigningProfile_basic(t *testing.T) {
	resourceName := "aws_signer_signing_profile.test_sp"
	rString := sdkacctest.RandString(48)
	profileName := fmt.Sprintf("tf_acc_sp_basic_%s", rString)

	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, signer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_providedName(profileName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(resourceName, &conf),
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
	resourceName := "aws_signer_signing_profile.test_sp"
	namePrefix := "tf_acc_sp_basic_"

	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, signer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_basic(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
				),
			},
		},
	})
}

func TestAccSignerSigningProfile_generateName(t *testing.T) {
	resourceName := "aws_signer_signing_profile.test_sp"

	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, signer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_generateName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
				),
			},
		},
	})
}

func TestAccSignerSigningProfile_tags(t *testing.T) {
	resourceName := "aws_signer_signing_profile.test_sp"
	namePrefix := "tf_acc_sp_basic_"

	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, signer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_tags(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "value2"),
				),
			},
			{
				Config: testAccSigningProfileConfig_updateTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "prod"),
				),
			},
		},
	})
}

func TestAccSignerSigningProfile_signatureValidityPeriod(t *testing.T) {
	resourceName := "aws_signer_signing_profile.test_sp"
	namePrefix := "tf_acc_sp_basic_"

	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, signer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileConfig_svp(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.type", "DAYS"),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.value", "10"),
				),
			},
			{
				Config: testAccSigningProfileConfig_updateSVP(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.type", "MONTHS"),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.value", "10"),
				),
			},
		},
	})
}

func testAccPreCheckSingerSigningProfile(t *testing.T, platformID string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SignerConn

	input := &signer.ListSigningPlatformsInput{}

	output, err := conn.ListSigningPlatforms(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if output == nil {
		t.Skip("skipping acceptance testing: empty response")
	}

	for _, platform := range output.Platforms {
		if platform == nil {
			continue
		}

		if aws.StringValue(platform.PlatformId) == platformID {
			return
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

func testAccCheckSigningProfileExists(res string, sp *signer.GetSigningProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Signing profile not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Signing Profile with that ARN does not exist")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SignerConn

		params := &signer.GetSigningProfileInput{
			ProfileName: aws.String(rs.Primary.ID),
		}

		getSp, err := conn.GetSigningProfile(params)
		if err != nil {
			return err
		}

		*sp = *getSp

		return nil
	}
}

func testAccCheckSigningProfileDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SignerConn

	time.Sleep(5 * time.Second)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_signer_signing_profile" {
			continue
		}

		out, err := conn.GetSigningProfile(&signer.GetSigningProfileInput{
			ProfileName: aws.String(rs.Primary.ID),
		})

		if *out.Status != signer.SigningProfileStatusCanceled && err == nil {
			return fmt.Errorf("Signing Profile not cancelled%s", *out.ProfileName)
		}

	}

	return nil
}
