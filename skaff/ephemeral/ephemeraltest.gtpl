// Copyright (c) HashiCorp, Inc.
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
	// using the services/{{ .SDKPackage }}/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
{{- end }}
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}"
	"github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
    "github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
{{- if .IncludeComments }}

	// TIP: You will often need to import the package that this test file lives
    // in. Since it is in the "test" context, it must import the package to use
    // any normal context constants, variables, or functions.
{{- end }}
	tf{{ .ServicePackage }} "github.com/hashicorp/terraform-provider-aws/internal/service/{{ .ServicePackage }}"
	"github.com/hashicorp/terraform-provider-aws/names"
)
{{ if .IncludeComments }}
// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this ephemeral resource's maintainability by following this
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
{{ if .IncludeComments }}

// TIP: ==== UNIT TESTS ====
// This is an example of a unit test. Its name is not prefixed with
// "TestAcc" like an acceptance test.
//
// Unlike acceptance tests, unit tests do not access AWS and are focused on a
// function (or method). Because of this, they are quick and cheap to run.
//
// In designing a ephemeral resource's implementation, isolate complex bits from AWS bits
// so that they can be tested through a unit test. We encourage more unit tests
// in the provider.
//
// Cut and dry functions using well-used patterns, like typical flatteners and
// expanders, don't need unit testing. However, if they are complex or
// intricate, they should be unit tested.
{{- end }}
func Test{{ .EphemeralResource }}ExampleUnitTest(t *testing.T) {
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
			got, err := tf{{ .ServicePackage }}.FunctionFromEphemeralResource(testCase.Input)

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
// standard functionality of the ephemeral resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// ephemeral resource name.
//
// Acceptance test access AWS and cost money to run.
{{- end }}
func TestAcc{{ .Service }}{{ .EphemeralResource }}DataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
    {{- if .IncludeComments }}
    // TIP: This is a long-running test guard for tests that run longer than
    // 300s (5 min) generally.
    {{- end }}
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	echoResourceName := "echo.test"
    dataPath := tfjsonpath.New("data")
    secretString := "super-secret"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.{{ .Service }}EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.{{ .Service }}ServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
        	tfversion.SkipBelow(tfversion.Version1_10_0),
        },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccTestEphemeralResourceConfig_basic(rName, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey(names.AttrCreatedDate), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_string"), knownvalue.StringExact(secretString)),
				},
			},
		},
	})
}

func testAcc{{ .EphemeralResource }}EphemeralResourceConfig_basic(rName, version string) string {
	return acctest.ConfigCompose(
	    acctest.ConfigWithEchoProvider("ephemeral.aws_{{ .ServicePackage }}_{{ .EphemeralResourceSnake }}.test"),
	    fmt.Sprintf(`
ephemeral "aws_{{ .ServicePackage }}_{{ .EphemeralResourceSnake }}" "test" {
  {{ .EphemeralResourceSnake }}_name             = %[1]q
  engine_type             = "Active{{ .Service }}"
  engine_version          = %[2]q
  host_instance_type      = "{{ .ServicePackage }}.t2.micro"
}
`, rName, version))
}
