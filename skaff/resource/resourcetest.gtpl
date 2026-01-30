// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package {{ .ServicePackage }}_test

{{- if .IncludeComments }}
// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.
{{- end }}

import (
{{- if .IncludeComments }}
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the service/{{ .SDKPackage }}/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
{{- end }}
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}"
	awstypes "github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
{{- if .IncludeComments }}

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
{{- end }}
	tf{{ .ServicePackage }} "github.com/hashicorp/terraform-provider-aws/internal/service/{{ .ServicePackage }}"
)
{{ if .IncludeComments }}
// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations
{{- end }}
{{- if .IncludeComments }}

// TIP: ==== UNIT TESTS ====
// This is an example of a unit test. Its name is not prefixed with
// "TestAcc" like an acceptance test.
//
// Unlike acceptance tests, unit tests do not access AWS and are focused on a
// function (or method). Because of this, they are quick and cheap to run.
//
// In designing a resource's implementation, isolate complex bits from AWS bits
// so that they can be tested through a unit test. We encourage more unit tests
// in the provider.
//
// Cut and dry functions using well-used patterns, like typical flatteners and
// expanders, don't need unit testing. However, if they are complex or
// intricate, they should be unit tested.
{{- end }}
func Test{{ .Resource }}ExampleUnitTest(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			got, err := tf{{ .ServicePackage }}.FunctionFromResource(testCase.Input)

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
{{ if .IncludeComments }}
// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance tests access AWS and cost money to run.
{{- end }}
func TestAcc{{ .Service }}{{ .Resource }}_basic(t *testing.T) {
	ctx := acctest.Context(t)
    {{- if .IncludeComments }}
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
    {{- end }}
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var {{ .ResourceLower }} {{ .SDKPackage }}.Describe{{ .ResourceAWS }}Response
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_{{ .ServicePackage }}_{{ .ResourceSnake }}.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.{{ .Service }}EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.{{ .Service }}ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheck{{ .Resource }}Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAcc{{ .Resource }}Config_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheck{{ .Resource }}Exists(ctx, t, resourceName, &{{ .ResourceLower }}),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					{{- if .IncludeComments }}
					// TIP: If the ARN can be partially or completely determined by the parameters passed, e.g. it contains the
					// value of `rName`, either include the values in the regex or check for an exact match using `acctest.CheckResourceAttrRegionalARN`
					{{- end }}
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "{{ .ServicePackage }}", regexache.MustCompile(`{{ .ResourceLower }}:.+$`)),
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

func TestAcc{{ .Service }}{{ .Resource }}_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var {{ .ResourceLower }} {{ .SDKPackage }}.Describe{{ .ResourceAWS }}Response
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_{{ .ServicePackage }}_{{ .ResourceSnake }}.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.{{ .Service }}EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.{{ .Service }}ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheck{{ .Resource }}Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAcc{{ .Resource }}Config_basic(rName, testAcc{{ .Resource }}VersionNewer),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheck{{ .Resource }}Exists(ctx, t, resourceName, &{{ .ResourceLower }}),
					{{- if .IncludeComments }}
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var Resource{{ .Resource}} = new{{ .Resource }}Resource
					{{- end }}
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tf{{ .ServicePackage }}.Resource{{ .Resource }}, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheck{{ .Resource }}Destroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).{{ .Service }}Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_{{ .ServicePackage }}_{{ .ResourceSnake }}" {
				continue
			}

			{{ if .IncludeComments }}
			// TIP: ==== FINDERS ====
			// The find function should be exported. Since it won't be used outside of the package, it can be exported
			// in the `exports_test.go` file.
			{{- end }}
			_, err := tf{{ .ServicePackage }}.Find{{ .Resource }}ByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
			    return create.Error(names.{{ .Service }}, create.ErrActionCheckingDestroyed, tf{{ .ServicePackage }}.ResName{{ .Resource }}, rs.Primary.ID, err)
			}

			return create.Error(names.{{ .Service }}, create.ErrActionCheckingDestroyed, tf{{ .ServicePackage }}.ResName{{ .Resource }}, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheck{{ .Resource }}Exists(ctx context.Context, t *testing.T, name string, {{ .ResourceLower }} *{{ .SDKPackage }}.Describe{{ .ResourceAWS }}Response) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.{{ .Service }}, create.ErrActionCheckingExistence, tf{{ .ServicePackage }}.ResName{{ .Resource }}, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.{{ .Service }}, create.ErrActionCheckingExistence, tf{{ .ServicePackage }}.ResName{{ .Resource }}, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).{{ .Service }}Client(ctx)

		resp, err := tf{{ .ServicePackage }}.Find{{ .Resource }}ByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.{{ .Service }}, create.ErrActionCheckingExistence, tf{{ .ServicePackage }}.ResName{{ .Resource }}, rs.Primary.ID, err)
		}

		*{{ .ResourceLower }} = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).{{ .Service }}Client(ctx)

	input := &{{ .SDKPackage }}.List{{ .ResourceAWS }}sInput{}

	_, err := conn.List{{ .ResourceAWS }}s(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheck{{ .Resource }}NotRecreated(before, after *{{ .SDKPackage }}.Describe{{ .ResourceAWS }}Response) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.{{ .ResourceAWS }}Id), aws.ToString(after.{{ .ResourceAWS }}Id); before != after {
			return create.Error(names.{{ .Service }}, create.ErrActionCheckingNotRecreated, tf{{ .ServicePackage }}.ResName{{ .Resource }}, aws.ToString(before.{{ .ResourceAWS }}Id), errors.New("recreated"))
		}

		return nil
	}
}

func testAcc{{ .Resource }}Config_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_{{ .ServicePackage }}_{{ .ResourceSnake }}" "test" {
  {{ .ResourceSnake }}_name             = %[1]q
  engine_type             = "Active{{ .Service }}"
  engine_version          = %[2]q
  host_instance_type      = "{{ .ServicePackage }}.t2.micro"
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
