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
	"fmt"
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

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.

func TestAccVpcLatticeListener_httpFixedResponse(t *testing.T) {
	ctx := acctest.Context(t)

	var listener vpclattice.GetListenerOutput
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
				Config: testAccListenerConfig_httpFixedResponse(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.0.status_code", "404"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile(`service/svc-.*/listener/listener-.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				//ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})

}

func TestAccVpcLatticeListener_httpForwardRuleToId(t *testing.T) {
	ctx := acctest.Context(t)

	var listener vpclattice.GetListenerOutput
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
				Config: testAccListenerConfig_httpForwardRuleToId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					// resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.status_code", "404"),
					// resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					// resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
					// 	"console_access": "false",
					// 	"groups.#":       "0",
					// 	"username":       "Test",
					// 	"password":       "TestTest1234",
					// }),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile(`service\/svc-.*\/listener\/listener-.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				//ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})

}

func TestAccVpcLatticeListener_httpForwardRuleMultipleTargetGroups(t *testing.T) {
	ctx := acctest.Context(t)
	targetGroupName0 := fmt.Sprintf("testtargetgroup-%s", sdkacctest.RandString(10))
	targetGroupName1 := fmt.Sprintf("testtargetgroup-%s", sdkacctest.RandString(10))

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"
	targetGroupResourceName := "aws_vpclattice_target_group.test"
	targetGroup1ResourceName := "aws_vpclattice_target_group.test1"

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
				Config: testAccListenerConfig_httpForwardRuleMultiTarget(rName, targetGroupName0, targetGroupName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "80"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_groups.0.weight", "80"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.1.target_group_identifier", targetGroup1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_groups.1.weight", "20"),
					// resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					// resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
					// 	"console_access": "false",
					// 	"groups.#":       "0",
					// 	"username":       "Test",
					// 	"password":       "TestTest1234",
					// }),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile(`service\/svc-.*\/listener\/listener-.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				//ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})

}

// func TestAccVPCLatticeListener_basic(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	// TIP: This is a long-running test guard for tests that run longer than
// 	// 300s (5 min) generally.
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var listener vpclattice.GetListenerOutput
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_vpclattice_listener.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
// 			testAccPreCheck(ctx, t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckListenerDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccListenerConfig_basic(rName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckListenerExists(ctx, resourceName, &listener),
// 					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
// 					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
// 					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
// 						"console_access": "false",
// 						"groups.#":       "0",
// 						"username":       "Test",
// 						"password":       "TestTest1234",
// 					}),
// 					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpclattice", regexp.MustCompile(`listener:+.`)),
// 				),
// 			},
// 			{
// 				ResourceName:            resourceName,
// 				ImportState:             true,
// 				ImportStateVerify:       true,
// 				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
// 			},
// 		},
// 	})
// }

// func TestAccVPCLatticeListener_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var listener vpclattice.GetListenerOutput
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_vpclattice_listener.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
// 			testAccPreCheck(t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckListenerDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccListenerConfig_basic(rName, testAccListenerVersionNewer),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckListenerExists(resourceName, &listener),
// 					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceListener(), resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

func testAccCheckListenerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_listener" {
				continue
			}

			// input := &vpclattice.GetListenerInput{
			// 	ListenerIdentifier: aws.String(rs.Primary.ID),
			// }
			_, err := conn.GetListener(ctx, &vpclattice.GetListenerInput{
				ListenerIdentifier: aws.String(rs.Primary.Attributes["listener_id"]),
				ServiceIdentifier:  aws.String(rs.Primary.Attributes["service_identifier"]),
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

func testAccCheckListenerExists(ctx context.Context, name string, listener *vpclattice.GetListenerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListener, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListener, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()
		resp, err := conn.GetListener(ctx, &vpclattice.GetListenerInput{
			ListenerIdentifier: aws.String(rs.Primary.Attributes["listener_id"]),
			ServiceIdentifier:  aws.String(rs.Primary.Attributes["service_identifier"]),
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListener, rs.Primary.ID, err)
		}

		*listener = *resp

		return nil
	}
}

// func testAccPreCheck(ctx context.Context, t *testing.T) {
// 	conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()

// 	input := &vpclattice.ListListenersInput{}
// 	_, err := conn.ListListeners(ctx, input)

// 	if acctest.PreCheckSkipError(err) {
// 		t.Skipf("skipping acceptance testing: %s", err)
// 	}

// 	if err != nil {
// 		t.Fatalf("unexpected PreCheck error: %s", err)
// 	}
// }

func testAccListenerConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 0), fmt.Sprintf(`
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


`, rName))
}

func testAccListenerConfig_httpForwardRuleToArn(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name        = %[1]q
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
}`, rName))
}

func testAccListenerConfig_httpForwardRuleToId(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  port               = 80
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 100
      }
    }
  }
}
`, rName))
}

func testAccListenerConfig_httpForwardRuleMultiTarget(rName string, targetGroupName0 string, targetGroupName1 string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test1" {
  name = %[2]q
  type = "INSTANCE"

  config {
    port           = 8080
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
}

resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  port               = 80
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 80
      }
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test1.id
        weight                  = 20
      }
    }
  }
}
`, rName, targetGroupName1))
}

func testAccListenerConfig_httpFixedResponse(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  port               = 80
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    fixed_response {
      status_code = 404
    }
  }
}
`, rName))
}
