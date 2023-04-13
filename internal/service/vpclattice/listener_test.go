package vpclattice_test

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/vpclattice/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
func TestListenerExampleUnitTest(t *testing.T) {
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
			got, err := tfvpclattice.FunctionFromResource(testCase.Input)

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

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccVPCLatticeListener_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listener vpclattice.DescribeListenerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpclattice", regexp.MustCompile(`listener:+.`)),
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

func TestAccVPCLatticeListener_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listener vpclattice.DescribeListenerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_basic(rName, testAccListenerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(resourceName, &listener),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceListener(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckListenerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_listener" {
				continue
			}

			input := &vpclattice.DescribeListenerInput{
				ListenerId: aws.String(rs.Primary.ID),
			}
			_, err := conn.DescribeListener(ctx, &vpclattice.DescribeListenerInput{
				ListenerId: aws.String(rs.Primary.ID),
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameListener, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckListenerExists(ctx context.Context, name string, listener *vpclattice.DescribeListenerResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListener, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListener, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()
		resp, err := conn.DescribeListener(ctx, &vpclattice.DescribeListenerInput{
			ListenerId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListener, rs.Primary.ID, err)
		}

		*listener = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()

	input := &vpclattice.ListListenersInput{}
	_, err := conn.ListListeners(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckListenerNotRecreated(before, after *vpclattice.DescribeListenerResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.ListenerId), aws.StringValue(after.ListenerId); before != after {
			return create.Error(names.VPCLattice, create.ErrActionCheckingNotRecreated, tfvpclattice.ResNameListener, aws.StringValue(before.ListenerId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccListenerConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 0), `
resource "aws_vpclattice_service" "test" {
  name = %[1]q
}

resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "INSTANCE"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
}


`, rName)
}

func testAccListenerConfig_httpForwardRuleToArn(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), `
resource "aws_vpclattice_listener" "test" {
  name        = "test"
  service_arn = aws_vpclattice_service.test.arn
  default_action {
    forward {
      target_groups = [{
        target_group_identifier = aws_vpclattice_target_group.test.arn
        weight                  = 100
        }
      ]
    }
  }
}`, rName)
}

func testAccListenerConfig_httpForwardRuleToId(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), `
resource "aws_vpclattice_listener" "test" {
  name        = "test"
  protocol    = "HTTP"
  service_arn = aws_vpclattice_service.test.arn
  default_action {
    forward {
      target_groups = [{
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 100
        }
      ]
    }
  }
}`, rName)
}

func testAccListenerConfig_httpFixedResponse(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), `
resource "aws_vpclattice_listener" "test" {
  name        = "test"
  protocol    = "HTTP"
  service_arn = aws_vpclattice_service.test.arn
  default_action {
    fixed_response {
      status_code = 404
    }
  }
}`, rName)
}
