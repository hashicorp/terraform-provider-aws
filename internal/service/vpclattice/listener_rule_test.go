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
		CheckDestroy:             testAccChecklistenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAcclistenerRule_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, "priority", "20"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile(`service/svc-.*/listener/listener-.*/rule/rule.+`)),
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

func TestAccVPCLatticeListenerRule_fixedResponse(t *testing.T) {
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
		CheckDestroy:             testAccChecklistenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAcclistenerRule_fixedResponse(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "priority", "10"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.0.status_code", "404"),
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

func TestAccVPCLatticeListenerRule_methodMatch(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccChecklistenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAcclistenerRule_methodMatch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "priority", "40"),
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

func TestAccVPCLatticeListenerRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccChecklistenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRule_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccListenerRule_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAcclistenerRule_basic(rName string) string {
	return fmt.Sprintf(`
	resource "aws_vpclattice_listener_rule" "test" {
		name = %q
		listener_identifier = "listener-0f84b4608eae14610"
		service_identifier = "svc-05cad7dcf6ee78d45"
		priority     = 20
		match {
			http_match {

				header_matches {
					name = "example-header"
					case_sensitive = false
			
					match {
					  exact = "example-contains"
					}
				  }

				path_match {
					case_sensitive = true
					match {
						prefix = "/example-path"
					  }
				}
			}
		}	  
		action  {
			forward {
				target_groups{
					target_group_identifier = "tg-00153386728e69d10"
					weight = 1
				}
				target_groups{
					target_group_identifier = "tg-0fcd8d514d231b311"
					weight = 2
				}
			}
			
		}
	}
`, rName)
}

func testAcclistenerRule_fixedResponse(rName string) string {
	return fmt.Sprintf(`

	resource "aws_vpclattice_listener_rule" "test" {
		name = %q
		listener_identifier = "listener-0f84b4608eae14610"
		service_identifier = "svc-05cad7dcf6ee78d45"
		priority = 10
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
		action {
			fixed_response {
				status_code = 404
			}
		}
	}
`, rName)
}

func testAccCheckListenerRuleExists(ctx context.Context, name string, rule *vpclattice.GetRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListenerRule, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListenerRule, name, errors.New("not set"))
		}

		serviceIdentifier := rs.Primary.Attributes["service_identifier"]
		listenerIdentifier := rs.Primary.Attributes["listener_identifier"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()
		resp, err := conn.GetRule(ctx, &vpclattice.GetRuleInput{
			RuleIdentifier:     aws.String(rs.Primary.Attributes["arn"]),
			ListenerIdentifier: aws.String(listenerIdentifier),
			ServiceIdentifier:  aws.String(serviceIdentifier),
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListenerRule, rs.Primary.ID, err)
		}

		*rule = *resp

		return nil
	}
}

func testAccChecklistenerRuleDestroy(ctx context.Context) resource.TestCheckFunc {
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
				RuleIdentifier:     aws.String(rs.Primary.Attributes["arn"]),
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

func testAccListenerRule_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
	resource "aws_vpclattice_listener_rule" "test" {
		name = %q
		listener_identifier = "listener-0f84b4608eae14610"
		service_identifier = "svc-05cad7dcf6ee78d45"
		priority = 30
		match {
			http_match {
				path_match {
					case_sensitive = false
					match {
						prefix = "/example-path"
					}
				}
			}
		}
		action {
			fixed_response {
				status_code = 404
			}
		}
		tags = {
			%[2]q = %[3]q
		  }
	}
`, rName, tagKey1, tagValue1)
}

func testAccListenerRule_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
	resource "aws_vpclattice_listener_rule" "test" {
		name = %q
		listener_identifier = "listener-0f84b4608eae14610"
		service_identifier = "svc-05cad7dcf6ee78d45"
		priority = 30
		match {
			http_match {
				path_match {
					case_sensitive = false
					match {
						prefix = "/example-path"
					}
				}
			}
		}
		action {
			fixed_response {
				status_code = 404
			}
		}
		tags = {
			%[2]q = %[3]q
			%[4]q = %[5]q
		  }
	}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAcclistenerRule_methodMatch(rName string) string {
	return fmt.Sprintf(`
	resource "aws_vpclattice_listener_rule" "test" {
		name = %q
		listener_identifier = "listener-0f84b4608eae14610"
		service_identifier = "svc-05cad7dcf6ee78d45"
		priority = 40
		match {
			http_match {

				method = "POST"

				header_matches {
					name = "example-header"
					case_sensitive = false

					match {
						contains = "example-contains"
					}
				}

				path_match {
					case_sensitive = true
					match {
						prefix = "/example-path"
					}
				}

			}
		}
		action {
			forward {
				target_groups {
					target_group_identifier = "tg-00153386728e69d10"
					weight = 1
				}
				target_groups {
					target_group_identifier = "tg-0fcd8d514d231b311"
					weight = 2
				}
			}
		}
	}
`, rName)
}
