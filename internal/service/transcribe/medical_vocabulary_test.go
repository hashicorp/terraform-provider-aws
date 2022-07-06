package transcribe_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftranscribe "github.com/hashicorp/terraform-provider-aws/internal/service/transcribe"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTranscribeMedicalVocabulary_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var medicalVocabulary transcribe.GetMedicalVocabularyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_medical_vocabulary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.TranscribeEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.TranscribeEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMedicalVocabularyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMedicalVocabularyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMedicalVocabularyExists(resourceName, &medicalVocabulary),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transcribe", regexp.MustCompile(`medicalvocabulary:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccTranscribeMedicalVocabulary_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var medicalvocabulary transcribe.DescribeMedicalVocabularyResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_medicalvocabulary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(transcribe.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, transcribe.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMedicalVocabularyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMedicalVocabularyConfig_basic(rName, testAccMedicalVocabularyVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMedicalVocabularyExists(resourceName, &medicalvocabulary),
					acctest.CheckResourceDisappears(acctest.Provider, tftranscribe.ResourceMedicalVocabulary(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMedicalVocabularyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transcribe_medicalvocabulary" {
			continue
		}

		input := &transcribe.DescribeMedicalVocabularyInput{
			MedicalVocabularyId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeMedicalVocabulary(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, transcribe.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected Transcribe MedicalVocabulary to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckMedicalVocabularyExists(name string, medicalvocabulary *transcribe.DescribeMedicalVocabularyResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transcribe MedicalVocabulary is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn
		resp, err := conn.DescribeMedicalVocabulary(&transcribe.DescribeMedicalVocabularyInput{
			MedicalVocabularyId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("Error describing Transcribe MedicalVocabulary: %s", err.Error())
		}

		*medicalvocabulary = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn

	input := &transcribe.ListMedicalVocabularysInput{}

	_, err := conn.ListMedicalVocabularys(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckMedicalVocabularyNotRecreated(before, after *transcribe.DescribeMedicalVocabularyResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.MedicalVocabularyId), aws.StringValue(after.MedicalVocabularyId); before != after {
			return fmt.Errorf("Transcribe MedicalVocabulary (%s/%s) recreated", before, after)
		}

		return nil
	}
}

func testAccMedicalVocabularyBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.test.id
  key    = "transcribe/test.txt"
  source = "test.txt"
}

`, rName)
}

func testAccMedicalVocabularyConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_transcribe_medical_vocabulary" "test" {
  medicalvocabulary_name             = %[1]q
  engine_type             = "ActiveTranscribe"
  engine_version          = %[2]q
  host_instance_type      = "transcribe.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}
