package transcribe_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/aws/aws-sdk-go-v2/service/transcribe/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftranscribe "github.com/hashicorp/terraform-provider-aws/internal/service/transcribe"
	"github.com/hashicorp/terraform-provider-aws/names"
)


func TestVocabularyFilterExampleUnitTest(t *testing.T) {
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

func TestAccTranscribeVocabularyFilter_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vocabularyfilter transcribe.DescribeVocabularyFilterResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_vocabulary_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.TranscribeEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.TranscribeEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVocabularyFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVocabularyFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVocabularyFilterExists(resourceName, &vocabularyfilter),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transcribe", regexp.MustCompile(`vocabularyfilter:+.`)),
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

func TestAccTranscribeVocabularyFilter_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vocabularyfilter transcribe.DescribeVocabularyFilterResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_vocabulary_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.TranscribeEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.TranscribeEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVocabularyFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVocabularyFilterConfig_basic(rName, testAccVocabularyFilterVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVocabularyFilterExists(resourceName, &vocabularyfilter),
					acctest.CheckResourceDisappears(acctest.Provider, tftranscribe.ResourceVocabularyFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVocabularyFilterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transcribe_vocabulary_filter" {
			continue
		}

		input := &transcribe.DescribeVocabularyFilterInput{
			VocabularyFilterId: aws.String(rs.Primary.ID),
		}
		_, err := conn.DescribeVocabularyFilter(ctx, &transcribe.DescribeVocabularyFilterInput{
			VocabularyFilterId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			var nfe *types.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return nil
			}
			return err
		}

		return names.Error(names.Transcribe, names.ErrActionCheckingDestroyed, tftranscribe.ResNameVocabularyFilter, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckVocabularyFilterExists(name string, vocabularyfilter *transcribe.DescribeVocabularyFilterResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return names.Error(names.Transcribe, names.ErrActionCheckingExistence, tftranscribe.ResNameVocabularyFilter, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return names.Error(names.Transcribe, names.ErrActionCheckingExistence, tftranscribe.ResNameVocabularyFilter, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn
		ctx := context.Background()
		resp, err := conn.DescribeVocabularyFilter(ctx, &transcribe.DescribeVocabularyFilterInput{
			VocabularyFilterId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return names.Error(names.Transcribe, names.ErrActionCheckingExistence, tftranscribe.ResNameVocabularyFilter, rs.Primary.ID, err)
		}

		*vocabularyfilter = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn
	ctx := context.Background()

	input := &transcribe.ListVocabularyFiltersInput{}
	_, err := conn.ListVocabularyFilters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckVocabularyFilterNotRecreated(before, after *transcribe.DescribeVocabularyFilterResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.VocabularyFilterId), aws.StringValue(after.VocabularyFilterId); before != after {
			return names.Error(names.Transcribe, names.ErrActionCheckingNotRecreated, tftranscribe.ResNameVocabularyFilter, aws.StringValue(before.VocabularyFilterId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccVocabularyFilterConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_transcribe_vocabulary_filter" "test" {
  vocabulary_filter_name             = %[1]q
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
