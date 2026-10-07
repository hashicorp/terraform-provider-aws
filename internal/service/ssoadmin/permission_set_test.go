// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsretry "github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	smithy "github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminPermissionSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT1H"),
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

func TestAccSSOAdminPermissionSet_duplicateName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_permission_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config:      testAccPermissionSetConfig_duplicateName(rName),
				ExpectError: regexache.MustCompile(`ConflictException: PermissionSet with name .* already exists`),
			},
		},
	})
}

func TestAccSSOAdminPermissionSet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
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
				Config: testAccPermissionSetConfig_tags2(rName, acctest.CtKey1, "updatedvalue1", acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, "updatedvalue1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPermissionSetConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
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

func TestAccSSOAdminPermissionSet_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
			{
				Config: testAccPermissionSetConfig_updateDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
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

func TestAccSSOAdminPermissionSet_updateRelayState(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", ""),
				),
			},
			{
				Config: testAccPermissionSetConfig_updateRelayState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
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

func TestAccSSOAdminPermissionSet_updateSessionDuration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccPermissionSetConfig_updateSessionDuration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT2H"),
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

// TestAccSSOAdminPermissionSet_RelayState_updateSessionDuration validates
// the resource's unchanged values (primarily relay_state) after updating the session_duration argument
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17411
func TestAccSSOAdminPermissionSet_RelayState_updateSessionDuration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_relayState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT1H"),
				),
			},
			{
				Config: testAccPermissionSetConfig_relayStateUpdateSessionDuration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT2H"),
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

func TestAccSSOAdminPermissionSet_mixedPolicyAttachments(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccPermissionSetConfig_mixedPolicyAttachments(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, t, resourceName),
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

func testAccCheckPermissionSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_permission_set" {
				continue
			}

			permissionSetARN, instanceARN, err := tfssoadmin.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfssoadmin.FindPermissionSetByTwoPartKey(ctx, conn, permissionSetARN, instanceARN)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSO Permission Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSOAdminPermissionSetExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		permissionSetARN, instanceARN, err := tfssoadmin.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

		_, err = tfssoadmin.FindPermissionSetByTwoPartKey(ctx, conn, permissionSetARN, instanceARN)

		return err
	}
}

func testAccPermissionSetConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccPermissionSetConfig_duplicateName(rName string) string {
	return testAccPermissionSetConfig_basic(rName) + fmt.Sprintf(`
resource "aws_ssoadmin_permission_set" "duplicate" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccPermissionSetConfig_updateDescription(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  description  = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccPermissionSetConfig_updateRelayState(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state  = "https://example.com"
}
`, rName)
}

func testAccPermissionSetConfig_updateSessionDuration(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  session_duration = "PT2H"
}
`, rName)
}

func testAccPermissionSetConfig_relayState(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  description      = %[1]q
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state      = "https://example.com"
  session_duration = "PT1H"
}
`, rName)
}

func testAccPermissionSetConfig_relayStateUpdateSessionDuration(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  description      = %[1]q
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state      = "https://example.com"
  session_duration = "PT2H"
}
`, rName)
}

func testAccPermissionSetConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccPermissionSetConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccPermissionSetConfig_mixedPolicyAttachments(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AlexaForBusinessDeviceSetup"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
    ]
  }
}
resource "aws_ssoadmin_permission_set_inline_policy" "test" {
  inline_policy      = data.aws_iam_policy_document.test.json
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`, rName)
}

func TestPermissionSetRetryer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		errorCode        string
		message          string
		expectedAttempts int
		expectedError    bool
	}{
		{
			name:             "duplicate permission set",
			errorCode:        "ConflictException",
			message:          "PermissionSet with name test already exists.",
			expectedAttempts: 1,
			expectedError:    true,
		},
		{
			name:             "permission set conflict without duplicate",
			errorCode:        "ConflictException",
			message:          "PermissionSet with name test is being updated.",
			expectedAttempts: 2,
		},
		{
			name:             "conflict for another resource",
			errorCode:        "ConflictException",
			message:          "Another resource already exists.",
			expectedAttempts: 2,
		},
		{
			name:             "throttling with duplicate message",
			errorCode:        "ThrottlingException",
			message:          "PermissionSet with name test already exists.",
			expectedAttempts: 2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			attempts := 0
			cfg := aws.Config{
				Region:      "us-east-1",
				Credentials: aws.AnonymousCredentials{},
				Retryer: func() aws.Retryer {
					return awsretry.NewStandard(func(o *awsretry.StandardOptions) {
						o.MaxAttempts = 2
						o.Backoff = awsretry.BackoffDelayerFunc(func(int, error) (time.Duration, error) {
							return 0, nil
						})
					})
				},
				HTTPClient: smithyhttp.ClientDoFunc(func(request *http.Request) (*http.Response, error) {
					attempts++
					statusCode := http.StatusOK
					body := `{}`
					headers := http.Header{"Content-Type": []string{"application/x-amz-json-1.1"}}
					if attempts == 1 {
						statusCode = http.StatusBadRequest
						headers.Set("X-Amzn-ErrorType", testCase.errorCode)
						body = fmt.Sprintf(`{"Message":%q}`, testCase.message)
					}
					return &http.Response{
						StatusCode: statusCode,
						Header:     headers,
						Body:       io.NopCloser(strings.NewReader(body)),
						Request:    request,
					}, nil
				}),
			}
			ctx := t.Context()
			factory, ok := tfssoadmin.ServicePackage(ctx).(interface {
				NewClient(context.Context, map[string]any) (*ssoadmin.Client, error)
			})
			if !ok {
				t.Fatal("SSO Admin service package has no client factory")
			}
			conn, err := factory.NewClient(ctx, map[string]any{
				"aws_sdkv2_config": &cfg,
				names.AttrEndpoint: "",
				names.AttrRegion:   cfg.Region,
			})
			if err != nil {
				t.Fatalf("creating SSO Admin client: %s", err)
			}

			_, err = conn.CreatePermissionSet(ctx, &ssoadmin.CreatePermissionSetInput{
				InstanceArn: aws.String("arn:aws:sso:::instance/ssoins-1234567890123456"),
				Name:        aws.String("test"),
			})

			if testCase.expectedError {
				apiErr, ok := errors.AsType[smithy.APIError](err)
				if !ok {
					t.Fatalf("expected API error, got %v", err)
				}
				if got, want := apiErr.ErrorCode(), testCase.errorCode; got != want {
					t.Errorf("error code = %q, want %q", got, want)
				}
				if got, want := apiErr.ErrorMessage(), testCase.message; got != want {
					t.Errorf("error message = %q, want %q", got, want)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got, want := attempts, testCase.expectedAttempts; got != want {
				t.Errorf("attempts = %d, want %d", got, want)
			}
		})
	}
}
