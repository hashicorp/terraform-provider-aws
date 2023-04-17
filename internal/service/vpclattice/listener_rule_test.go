package vpclattice_test

import (
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
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeListenerRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListeningRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListeningRule_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, resourceName, &listenerRule),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile(`service/.+/listener/.+/rule/.+`)),
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

func testAccListeningRule_basic(rName string) string {
	return fmt.Sprintf(`
	resource "aws_vpclattice_listener_rule" "test" {
		name = %q
		listener_identifier = "listener-0238d4e9479096392"
		service_identifier = "svc-05cad7dcf6ee78d45"
		priority     = 93	  
		action  {
			forward {
				target_groups{
					target_group_identifier = "tg-00153386728e69d10"
					weight = 1
				}
			}
			
		}
		match {
			http_match {
				path_match {
					case_sensitive = false
					match {
						exact = "/example-path"
					  }
				}
			}
		}
	}
`, rName)
}

func testAccCheckListenerRuleExists(ctx context.Context, name string, listenerRule *vpclattice.GetRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListenerRule, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListenerRule, name, errors.New("not set"))
		}

		listenerRuleResource, ok := s.RootModule().Resources["aws_vpclattice_listener_rule.test"]
		if !ok {
			return fmt.Errorf("Not found: %s", "aws_vpclattice_listener_rule.test")
		}

		listenerIdentifier := listenerRuleResource.Primary.Attributes["listener_identifier"]
		serviceIdentifier := listenerRuleResource.Primary.Attributes["service_identifier"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()
		resp, err := conn.GetRule(ctx, &vpclattice.GetRuleInput{
			RuleIdentifier:     aws.String(rs.Primary.ID),
			ListenerIdentifier: aws.String(listenerIdentifier),
			ServiceIdentifier:  aws.String(serviceIdentifier),
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListenerRule, rs.Primary.ID, err)
		}

		*listenerRule = *resp

		return nil
	}
}

func testAccCheckListeningRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_listener_rule" {
				continue
			}

			listenerRuleResource, ok := s.RootModule().Resources["aws_vpclattice_listener_rule.test"]
			if !ok {
				return fmt.Errorf("Not found: %s", "aws_vpclattice_listener_rule.test")
			}

			listenerIdentifier := listenerRuleResource.Primary.Attributes["listener_identifier"]
			serviceIdentifier := listenerRuleResource.Primary.Attributes["service_identifier"]

			_, err := conn.GetRule(ctx, &vpclattice.GetRuleInput{
				RuleIdentifier:     aws.String(rs.Primary.ID),
				ListenerIdentifier: aws.String(listenerIdentifier),
				ServiceIdentifier:  aws.String(serviceIdentifier),
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameListenerRule, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}
