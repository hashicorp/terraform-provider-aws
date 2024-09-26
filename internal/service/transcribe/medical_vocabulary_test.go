// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transcribe_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftranscribe "github.com/hashicorp/terraform-provider-aws/internal/service/transcribe"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTranscribeMedicalVocabulary_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var medicalVocabulary transcribe.GetMedicalVocabularyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_medical_vocabulary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TranscribeEndpointID)
			testAccMedicalVocabularyPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TranscribeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMedicalVocabularyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMedicalVocabularyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMedicalVocabularyExists(ctx, resourceName, &medicalVocabulary),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "download_uri"),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, "en-US"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vocabulary_file_uri", "download_uri"},
			},
		},
	})
}

func TestAccTranscribeMedicalVocabulary_updateS3URI(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var medicalVocabulary transcribe.GetMedicalVocabularyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_medical_vocabulary.test"

	file1 := "test1.txt"
	file2 := "test2.txt"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TranscribeEndpointID)
			testAccMedicalVocabularyPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TranscribeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMedicalVocabularyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMedicalVocabularyConfig_updateFile(rName, file1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMedicalVocabularyExists(ctx, resourceName, &medicalVocabulary),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "vocabulary_file_uri", "s3://"+rName+"/transcribe/test1.txt"),
				),
			},
			{
				Config: testAccMedicalVocabularyConfig_updateFile(rName, file2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMedicalVocabularyExists(ctx, resourceName, &medicalVocabulary),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "vocabulary_file_uri", "s3://"+rName+"/transcribe/test2.txt"),
				),
			},
		},
	})
}

func TestAccTranscribeMedicalVocabulary_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var medicalVocabulary transcribe.GetMedicalVocabularyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_medical_vocabulary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TranscribeEndpointID)
			testAccMedicalVocabularyPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TranscribeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMedicalVocabularyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMedicalVocabularyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMedicalVocabularyExists(ctx, resourceName, &medicalVocabulary),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccMedicalVocabularyConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMedicalVocabularyExists(ctx, resourceName, &medicalVocabulary),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccMedicalVocabularyConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMedicalVocabularyExists(ctx, resourceName, &medicalVocabulary),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccTranscribeMedicalVocabulary_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var medicalVocabulary transcribe.GetMedicalVocabularyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_medical_vocabulary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TranscribeEndpointID)
			testAccMedicalVocabularyPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TranscribeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMedicalVocabularyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMedicalVocabularyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMedicalVocabularyExists(ctx, resourceName, &medicalVocabulary),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftranscribe.ResourceMedicalVocabulary(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMedicalVocabularyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transcribe_medical_vocabulary" {
				continue
			}

			_, err := tftranscribe.FindMedicalVocabularyByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Expected Transcribe MedicalVocabulary to be destroyed, %s found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMedicalVocabularyExists(ctx context.Context, name string, medicalVocabulary *transcribe.GetMedicalVocabularyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transcribe MedicalVocabulary is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeClient(ctx)
		resp, err := tftranscribe.FindMedicalVocabularyByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("Error describing Transcribe MedicalVocabulary: %s", err.Error())
		}

		*medicalVocabulary = *resp

		return nil
	}
}

func testAccMedicalVocabularyPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeClient(ctx)

	input := &transcribe.ListMedicalVocabulariesInput{}

	_, err := conn.ListMedicalVocabularies(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccMedicalVocabularyBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "object1" {
  bucket = aws_s3_bucket.test.id
  key    = "transcribe/test1.txt"
  source = "test-fixtures/medical_vocabulary_test1.txt"
}

resource "aws_s3_object" "object2" {
  bucket = aws_s3_bucket.test.id
  key    = "transcribe/test2.txt"
  source = "test-fixtures/medical_vocabulary_test2.txt"
}

`, rName)
}

func testAccMedicalVocabularyConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccMedicalVocabularyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_transcribe_medical_vocabulary" "test" {
  vocabulary_name     = %[1]q
  language_code       = "en-US"
  vocabulary_file_uri = "s3://${aws_s3_bucket.test.id}/${aws_s3_object.object1.key}"

  tags = {
    tag1 = "value1"
    tag2 = "value3"
  }

  depends_on = [
    aws_s3_object.object1
  ]
}
`, rName))
}

func testAccMedicalVocabularyConfig_updateFile(rName, fileName string) string {
	return acctest.ConfigCompose(
		testAccMedicalVocabularyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_transcribe_medical_vocabulary" "test" {
  vocabulary_name     = %[1]q
  language_code       = "en-US"
  vocabulary_file_uri = "s3://${aws_s3_bucket.test.id}/transcribe/%[2]s"

  tags = {
    tag1 = "value1"
    tag2 = "value3"
  }

  depends_on = [
    aws_s3_object.object1,
    aws_s3_object.object2
  ]
}
`, rName, fileName))
}

func testAccMedicalVocabularyConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(
		testAccMedicalVocabularyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_transcribe_medical_vocabulary" "test" {
  vocabulary_name     = %[1]q
  language_code       = "en-US"
  vocabulary_file_uri = "s3://${aws_s3_bucket.test.id}/${aws_s3_object.object1.key}"

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [
    aws_s3_object.object1
  ]
}
`, rName, key1, value1))
}

func testAccMedicalVocabularyConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccMedicalVocabularyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_transcribe_medical_vocabulary" "test" {
  vocabulary_name     = %[1]q
  language_code       = "en-US"
  vocabulary_file_uri = "s3://${aws_s3_bucket.test.id}/${aws_s3_object.object1.key}"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [
    aws_s3_object.object1
  ]
}
`, rName, key1, value1, key2, value2))
}
