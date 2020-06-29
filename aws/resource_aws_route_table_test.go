package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func init() {
	resource.AddTestSweepers("aws_route_table", &resource.Sweeper{
		Name: "aws_route_table",
		F:    testSweepRouteTables,
	})
}

func testSweepRouteTables(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	input := &ec2.DescribeRouteTablesInput{}
	err = conn.DescribeRouteTablesPages(input, func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
		for _, routeTable := range page.RouteTables {
			isMainRouteTableAssociation := false

			for _, routeTableAssociation := range routeTable.Associations {
				if aws.BoolValue(routeTableAssociation.Main) {
					isMainRouteTableAssociation = true
					break
				}

				input := &ec2.DisassociateRouteTableInput{
					AssociationId: routeTableAssociation.RouteTableAssociationId,
				}

				log.Printf("[DEBUG] Deleting Route Table Association: %s", input)
				_, err := conn.DisassociateRouteTable(input)
				if err != nil {
					log.Printf("[ERROR] Error deleting Route Table Association (%s): %s", aws.StringValue(routeTableAssociation.RouteTableAssociationId), err)
				}
			}

			if isMainRouteTableAssociation {
				log.Printf("[DEBUG] Skipping Main Route Table: %s", aws.StringValue(routeTable.RouteTableId))
				continue
			}

			input := &ec2.DeleteRouteTableInput{
				RouteTableId: routeTable.RouteTableId,
			}

			log.Printf("[DEBUG] Deleting Route Table: %s", input)
			_, err := conn.DeleteRouteTable(input)
			if err != nil {
				log.Printf("[ERROR] Error deleting Route Table (%s): %s", aws.StringValue(routeTable.RouteTableId), err)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Route Table sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing Route Tables: %s", err)
	}

	return nil
}

func TestAccAWSRouteTable_basic(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 1),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSRouteTable_disappears(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRouteTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRouteTable_IPv4_To_InternetGateway(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr1 := "10.2.0.0/16"
	destinationCidr2 := "10.3.0.0/16"
	destinationCidr3 := "10.4.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4InternetGateway(rName, destinationCidr1, destinationCidr2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr1,
						"ipv6_cidr_block": "",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr2,
						"ipv6_cidr_block": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSRouteTableConfigIpv4InternetGateway(rName, destinationCidr2, destinationCidr3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr2,
						"ipv6_cidr_block": "",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr3,
						"ipv6_cidr_block": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv4_To_Instance(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4Instance(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr,
						"ipv6_cidr_block": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv6_To_EgressOnlyInternetGateway(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv6EgressOnlyInternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      "",
						"ipv6_cidr_block": destinationCidr,
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			/*
				TODO Fix this.
				{
					// Verify that expanded form of the destination CIDR causes no diff.
					Config: testAccAWSRouteTableConfigIpv6EgressOnlyInternetGateway(rName, "::0/0"),
					PlanOnly: true,
				},
			*/
		},
	})
}

func TestAccAWSRouteTable_tags(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
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
				Config: testAccAWSRouteTableConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSRouteTableConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

// For GH-13545, Fixes panic on an empty route config block
func TestAccAWSRouteTable_panicEmptyRoute(t *testing.T) {
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSRouteTableConfigPanicEmptyRoute(rName),
				ExpectError: regexp.MustCompile("The request must contain exactly one of: destinationCidrBlock, destinationIpv6CidrBlock, destinationPrefixListId"),
			},
		},
	})
}

func TestAccAWSRouteTable_Route_ConfigMode(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr1 := "10.2.0.0/16"
	destinationCidr2 := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4InternetGateway(rName, destinationCidr1, destinationCidr2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr1,
						"ipv6_cidr_block": "",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr2,
						"ipv6_cidr_block": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRouteTableConfigRouteConfigModeNoBlocks(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr1,
						"ipv6_cidr_block": "",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr2,
						"ipv6_cidr_block": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRouteTableConfigRouteConfigModeZeroed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 1),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv4_To_TransitGateway(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4TransitGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr,
						"ipv6_cidr_block": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv4_To_VpcPeeringConnection(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4VpcPeeringConnection(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr,
						"ipv6_cidr_block": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_vgwRoutePropagation(t *testing.T) {
	var routeTable ec2.RouteTable
	var vpnGateway1, vpnGateway2 ec2.VpnGateway
	resourceName := "aws_route_table.test"
	vgwResourceName1 := "aws_vpn_gateway.test1"
	vgwResourceName2 := "aws_vpn_gateway.test2"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigVgwRoutePropagation(rName, vgwResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 1),
					testAccCheckVpnGatewayExists(vgwResourceName1, &vpnGateway1),
					testAccCheckVpnGatewayExists(vgwResourceName2, &vpnGateway2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "1"),
					testAccCheckAWSRouteTablePropagatingVgw(resourceName, &vpnGateway1),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSRouteTableConfigVgwRoutePropagation(rName, vgwResourceName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 1),
					testAccCheckVpnGatewayExists(vgwResourceName1, &vpnGateway1),
					testAccCheckVpnGatewayExists(vgwResourceName2, &vpnGateway2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "1"),
					testAccCheckAWSRouteTablePropagatingVgw(resourceName, &vpnGateway2),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_ConditionalCidrBlock(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigConditionalIpv4Ipv6(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      "0.0.0.0/0",
						"ipv6_cidr_block": "",
					}),
				),
			},
			{
				Config: testAccAWSRouteTableConfigConditionalIpv4Ipv6(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      "",
						"ipv6_cidr_block": "::/0",
					}),
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

func TestAccAWSRouteTable_IPv4_To_NatGateway(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4NatGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      destinationCidr,
						"ipv6_cidr_block": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv6_To_NetworkInterface(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv6NetworkInterface(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      "",
						"ipv6_cidr_block": destinationCidr,
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

/*
TODO Fix this. When associating RT with VPCE, need to wait until available.
func TestAccAWSRouteTable_VpcMultipleCidrs_VpcEndpointAssociation(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigVpcMultipleCidrsPlusVpcEndpointAssociation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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
*/

func testAccCheckRouteTableExists(n string, v *ec2.RouteTable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
			RouteTableIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.RouteTables) == 0 {
			return fmt.Errorf("RouteTable not found")
		}

		*v = *resp.RouteTables[0]

		return nil
	}
}

func testAccCheckRouteTableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route_table" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
			RouteTableIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			if len(resp.RouteTables) > 0 {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		if !isAWSErr(err, "InvalidRouteTableID.NotFound", "") {
			return err
		}
	}

	return nil
}

func testAccCheckAWSRouteTableNumberOfRoutes(routeTable *ec2.RouteTable, n int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len := len(routeTable.Routes); len != n {
			return fmt.Errorf("Route Table has incorrect number of routes (Expected=%d, Actual=%d)\n", n, len)
		}

		return nil
	}
}

func testAccCheckAWSRouteTablePropagatingVgw(resourceName string, vpnGateway *ec2.VpnGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return tfawsresource.TestCheckTypeSetElemAttr(resourceName, "propagating_vgws.*", aws.StringValue(vpnGateway.VpnGatewayId))(s)
	}
}

func testAccAWSRouteTableConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccAWSRouteTableConfigIpv4InternetGateway(rName, destinationCidr1, destinationCidr2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = %[2]q
    gateway_id = aws_internet_gateway.test.id
  }

  route {
    cidr_block = %[3]q
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr1, destinationCidr2)
}

func testAccAWSRouteTableConfigIpv6EgressOnlyInternetGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    ipv6_cidr_block        = %[2]q
    egress_only_gateway_id = aws_egress_only_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr)
}

func testAccAWSRouteTableConfigIpv4Instance(rName, destinationCidr string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
data "aws_availability_zones" "current" {
  # Exclude usw2-az4 (us-west-2d) as it has limited instance types.
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.current.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block  = %[2]q
    instance_id = aws_instance.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr))
}

func testAccAWSRouteTableConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSRouteTableConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSRouteTableConfigIpv4VpcPeeringConnection(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.target.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block                = %[2]q
    vpc_peering_connection_id = aws_vpc_peering_connection.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr)
}

func testAccAWSRouteTableConfigVgwRoutePropagation(rName, vgwResourceName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test1" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test2" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = aws_vpc.test.id
  vpn_gateway_id = %[2]s.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  propagating_vgws = [aws_vpn_gateway_attachment.test.vpn_gateway_id]

  tags = {
    Name = %[1]q
  }
}
`, rName, vgwResourceName)
}

// For GH-13545
func testAccAWSRouteTableConfigPanicEmptyRoute(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {}

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSRouteTableConfigRouteConfigModeNoBlocks(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSRouteTableConfigRouteConfigModeZeroed(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route = []

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSRouteTableConfigIpv4TransitGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # IncorrectState: Transit Gateway is not available in availability zone us-west-2d
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block         = %[2]q
    transit_gateway_id = aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr)
}

func testAccAWSRouteTableConfigConditionalIpv4Ipv6(rName string, ipv6Route bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = %[1]q
  }
}


locals {
  ipv6             = %[2]t
  destination      = "0.0.0.0/0"
  destination_ipv6 = "::/0"
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  route {
    cidr_block      = "${local.ipv6 ? "" : local.destination}"
    ipv6_cidr_block = "${local.ipv6 ? local.destination_ipv6 : ""}"
    gateway_id      = "${aws_internet_gateway.test.id}"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv6Route)
}

func testAccAWSRouteTableConfigIpv4NatGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block     = %[2]q
    nat_gateway_id = aws_nat_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr)
}

func testAccAWSRouteTableConfigIpv6NetworkInterface(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    ipv6_cidr_block      = %[2]q
    network_interface_id = aws_network_interface.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr)
}

/*
func testAccAWSRouteTableConfigVpcMultipleCidrsPlusVpcEndpointAssociation(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "172.2.0.0/16"
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_route_table_association" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.test.id
  route_table_id  = aws_route_table.test.id
}
`, rName)
}
*/
