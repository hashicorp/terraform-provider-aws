// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNetworkInsightsAccessScope_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_access_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAccessScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAccessScopeConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkInsightsAccessScopeExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNFormat(resourceName, tfjsonpath.New(names.AttrARN), "ec2", "network-insights-access-scope/{id}"),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("match_paths"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrSource: knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"packet_header_statement": knownvalue.ListExact([]knownvalue.Check{}),
									"resource_statement": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											names.AttrResources: knownvalue.Null(),
											"resource_types":    knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("AWS::EC2::NetworkInterface")}),
										}),
									}),
								}),
							}),
							names.AttrDestination: knownvalue.ListExact([]knownvalue.Check{}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exclude_paths"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetworkInsightsAccessScope_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_access_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAccessScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAccessScopeConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkInsightsAccessScopeExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceNetworkInsightsAccessScope, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkInsightsAccessScope_matchPaths_resources(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_access_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAccessScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAccessScopeConfig_matchPathsResources(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkInsightsAccessScopeExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("match_paths").AtSliceIndex(0).AtMapKey(names.AttrSource).AtSliceIndex(0).AtMapKey("resource_statement").AtSliceIndex(0).AtMapKey(names.AttrResources), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("vpc-123456"),
						knownvalue.StringExact("vpc-654321"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("match_paths").AtSliceIndex(0).AtMapKey(names.AttrSource).AtSliceIndex(0).AtMapKey("resource_statement").AtSliceIndex(0).AtMapKey("resource_types"), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetworkInsightsAccessScope_matchPaths_packetHeaderStatement(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_access_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAccessScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAccessScopeConfig_matchPathsPacketHeaderStatement(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkInsightsAccessScopeExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("match_paths").AtSliceIndex(0).AtMapKey(names.AttrSource).AtSliceIndex(0).AtMapKey("packet_header_statement").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"source_addresses":         knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("10.0.0.0/16")}),
						"destination_addresses":    knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("192.168.0.0/24")}),
						"source_ports":             knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("443")}),
						"destination_ports":        knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("80")}),
						"source_prefix_lists":      knownvalue.Null(),
						"destination_prefix_lists": knownvalue.Null(),
						"protocols":                knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("tcp")}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("match_paths").AtSliceIndex(0).AtMapKey(names.AttrSource).AtSliceIndex(0).AtMapKey("resource_statement"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetworkInsightsAccessScope_matchPaths_destination(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_access_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAccessScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAccessScopeConfig_matchPathsDestination(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkInsightsAccessScopeExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("match_paths").AtSliceIndex(0).AtMapKey(names.AttrSource).AtSliceIndex(0).AtMapKey("resource_statement").AtSliceIndex(0).AtMapKey("resource_types"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("AWS::EC2::NetworkInterface"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("match_paths").AtSliceIndex(0).AtMapKey(names.AttrDestination).AtSliceIndex(0).AtMapKey("resource_statement").AtSliceIndex(0).AtMapKey("resource_types"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("AWS::EC2::InternetGateway"),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetworkInsightsAccessScope_excludePaths_throughResources(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_access_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAccessScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAccessScopeConfig_excludePathsThroughResources(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkInsightsAccessScopeExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exclude_paths").AtSliceIndex(0).AtMapKey("through_resources"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"resource_statement": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrResources: knownvalue.Null(),
									"resource_types":    knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("AWS::EC2::NatGateway")}),
								}),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetworkInsightsAccessScope_excludePaths(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_access_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAccessScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAccessScopeConfig_excludePaths(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkInsightsAccessScopeExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exclude_paths").AtSliceIndex(0).AtMapKey(names.AttrSource).AtSliceIndex(0).AtMapKey("resource_statement").AtSliceIndex(0).AtMapKey("resource_types"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("AWS::EC2::InternetGateway"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exclude_paths").AtSliceIndex(0).AtMapKey("through_resources"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetworkInsightsAccessScope_multiplePaths(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_access_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsAccessScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsAccessScopeConfig_multiplePaths(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkInsightsAccessScopeExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("match_paths").AtSliceIndex(0).AtMapKey(names.AttrSource).AtSliceIndex(0).AtMapKey("resource_statement").AtSliceIndex(0).AtMapKey("resource_types"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("AWS::EC2::NetworkInterface"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("match_paths").AtSliceIndex(1).AtMapKey(names.AttrSource).AtSliceIndex(0).AtMapKey("resource_statement").AtSliceIndex(0).AtMapKey("resource_types"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("AWS::EC2::InternetGateway"),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckNetworkInsightsAccessScopeExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		_, err := tfec2.FindNetworkInsightsAccessScopeByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckNetworkInsightsAccessScopeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_network_insights_access_scope" {
				continue
			}

			_, err := tfec2.FindNetworkInsightsAccessScopeByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Network Insights Access Scope %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVPCNetworkInsightsAccessScopeConfig_basic() string {
	return `
resource "aws_ec2_network_insights_access_scope" "test" {
  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
  }
}
`
}

func testAccVPCNetworkInsightsAccessScopeConfig_matchPathsResources() string {
	return `
resource "aws_ec2_network_insights_access_scope" "test" {
  match_paths {
    source {
      resource_statement {
        resources = ["vpc-123456", "vpc-654321"]
      }
    }
  }
}
`
}

func testAccVPCNetworkInsightsAccessScopeConfig_matchPathsPacketHeaderStatement() string {
	return `
resource "aws_ec2_network_insights_access_scope" "test" {
  match_paths {
    source {
      packet_header_statement {
        source_addresses      = ["10.0.0.0/16"]
        destination_addresses = ["192.168.0.0/24"]
        source_ports          = ["443"]
        destination_ports     = ["80"]
        protocols             = ["tcp"]
      }
    }
  }
}
`
}

func testAccVPCNetworkInsightsAccessScopeConfig_matchPathsDestination() string {
	return `
resource "aws_ec2_network_insights_access_scope" "test" {
  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
    destination {
      resource_statement {
        resource_types = ["AWS::EC2::InternetGateway"]
      }
    }
  }
}
`
}

func testAccVPCNetworkInsightsAccessScopeConfig_excludePathsThroughResources() string {
	return `
resource "aws_ec2_network_insights_access_scope" "test" {
  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
  }
  exclude_paths {
    through_resources {
      resource_statement {
        resource_types = ["AWS::EC2::NatGateway"]
      }
    }
  }
}
`
}

func testAccVPCNetworkInsightsAccessScopeConfig_excludePaths() string {
	return `
resource "aws_ec2_network_insights_access_scope" "test" {
  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
  }
  exclude_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::InternetGateway"]
      }
    }
  }
}
`
}

func testAccVPCNetworkInsightsAccessScopeConfig_multiplePaths() string {
	return `
resource "aws_ec2_network_insights_access_scope" "test" {
  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
  }
  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::InternetGateway"]
      }
    }
  }
}
`
}
