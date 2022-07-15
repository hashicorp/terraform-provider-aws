package transcribe_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/aws/aws-sdk-go-v2/service/transcribe/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftranscribe "github.com/hashicorp/terraform-provider-aws/internal/service/transcribe"
)


func TestVocabularyExampleUnitTest(t *testing.T) {
	testCases := []struct {
		TestName string
		Input    string
		Expected string
		Error    bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
			Error:    true,
		},
		{
			TestName: "descriptive name",
			Input:    "some input",
			Expected: "some output",
			Error:    false,
		},
		{
			TestName: "another descriptive name",
			Input:    "more input",
			Expected: "more output",
			Error:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got, err := tftranscribe.FunctionFromResource(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s) and no error, expected error", got)
			}

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestAccTranscribeVocabulary_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vocabulary transcribe.DescribeVocabularyResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_vocabulary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(transcribe.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, transcribe.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVocabularyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVocabularyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVocabularyExists(resourceName, &vocabulary),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transcribe", regexp.MustCompile(`vocabulary:+.`)),
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

func TestAccTranscribeVocabulary_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vocabulary transcribe.DescribeVocabularyResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_vocabulary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(transcribe.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, transcribe.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVocabularyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVocabularyConfig_basic(rName, testAccVocabularyVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVocabularyExists(resourceName, &vocabulary),
					acctest.CheckResourceDisappears(acctest.Provider, tftranscribe.ResourceVocabulary(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVocabularyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transcribe_vocabulary" {
			continue
		}

		input := &transcribe.DescribeVocabularyInput{
			VocabularyId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeVocabulary(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, transcribe.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}

		return names.Error(names.Transcribe, names.ErrActionCheckingDestroyed, tftranscribe.ResNameVocabulary, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckVocabularyExists(name string, vocabulary *transcribe.DescribeVocabularyResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return names.Error(names.Transcribe, names.ErrActionCheckingExistence, tftranscribe.ResNameVocabulary, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return names.Error(names.Transcribe, names.ErrActionCheckingExistence, tftranscribe.ResNameVocabulary, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn
		resp, err := conn.DescribeVocabulary(&transcribe.DescribeVocabularyInput{
			VocabularyId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return names.Error(names.Transcribe, names.ErrActionCheckingExistence, tftranscribe.ResNameVocabulary, rs.Primary.ID, err)
		}

		*vocabulary = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn

	input := &transcribe.ListVocabularysInput{}

	_, err := conn.ListVocabularys(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckVocabularyNotRecreated(before, after *transcribe.DescribeVocabularyResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.VocabularyId), aws.StringValue(after.VocabularyId); before != after {
			return names.Error(names.Transcribe, names.ErrActionCheckingNotRecreated, tftranscribe.ResNameVocabulary, aws.StringValue(before.VocabularyId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccVocabularyConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_transcribe_vocabulary" "test" {
  vocabulary_name             = %[1]q
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
