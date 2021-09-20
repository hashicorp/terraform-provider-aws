package aws

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/aws/aws-sdk-go/service/ec2"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/directconnect/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/directconnect/lister"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_dx_gateway_association", &resource.Sweeper{
		Name: "aws_dx_gateway_association",
		F:    testSweepDirectConnectGatewayAssociations,
		Dependencies: []string{
			"aws_dx_gateway_association_proposal",
		},
	})
}

func testSweepDirectConnectGatewayAssociations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).dxconn
	input := &directconnect.DescribeDirectConnectGatewaysInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*testSweepResource, 0)

	err = lister.DescribeDirectConnectGatewaysPages(conn, input, func(page *directconnect.DescribeDirectConnectGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gateway := range page.DirectConnectGateways {
			directConnectGatewayID := aws.StringValue(gateway.DirectConnectGatewayId)

			input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				DirectConnectGatewayId: aws.String(directConnectGatewayID),
			}

			err := lister.DescribeDirectConnectGatewayAssociationsPages(conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, association := range page.DirectConnectGatewayAssociations {
					gatewayID := aws.StringValue(association.AssociatedGateway.Id)

					if aws.StringValue(association.AssociatedGateway.Region) != region {
						log.Printf("[INFO] Skipping Direct Connect Gateway (%s) Association (%s) in different home region: %s", directConnectGatewayID, gatewayID, aws.StringValue(association.AssociatedGateway.Region))
						continue
					}

					if state := aws.StringValue(association.AssociationState); state != directconnect.GatewayAssociationStateAssociated {
						log.Printf("[INFO] Skipping Direct Connect Gateway (%s) Association in non-available (%s) state: %s", directConnectGatewayID, state, gatewayID)
						continue
					}

					r := resourceAwsDxGatewayAssociation()
					d := r.Data(nil)
					d.SetId(tfdirectconnect.GatewayAssociationCreateResourceID(directConnectGatewayID, gatewayID))

					sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Direct Connect Gateway Associations (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Print(fmt.Errorf("[WARN] Skipping Direct Connect Gateway Association sweep for %s: %w", region, err))
		return sweeperErrs // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Direct Connect Gateways (%s): %w", region, err))
	}

	// Handle cross-account EC2 Transit Gateway associations.
	// Direct Connect does not provide an easy lookup method for
	// these within the service itself so they can only be found
	// via AssociatedGatewayId of the EC2 Transit Gateway since the
	// DirectConnectGatewayId lives in the other account.
	ec2conn := client.(*AWSClient).ec2conn

	err = ec2conn.DescribeTransitGatewaysPages(&ec2.DescribeTransitGatewaysInput{}, func(page *ec2.DescribeTransitGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, transitGateway := range page.TransitGateways {
			if aws.StringValue(transitGateway.State) == ec2.TransitGatewayStateDeleted {
				continue
			}

			transitGatewayID := aws.StringValue(transitGateway.TransitGatewayId)

			input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				AssociatedGatewayId: aws.String(transitGatewayID),
			}

			err := lister.DescribeDirectConnectGatewayAssociationsPages(conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, association := range page.DirectConnectGatewayAssociations {
					directConnectGatewayID := aws.StringValue(association.DirectConnectGatewayId)

					if state := aws.StringValue(association.AssociationState); state != directconnect.GatewayAssociationStateAssociated {
						log.Printf("[INFO] Skipping Direct Connect Gateway (%s) Association in non-available (%s) state: %s", directConnectGatewayID, state, transitGatewayID)
						continue
					}

					r := resourceAwsDxGatewayAssociation()
					d := r.Data(nil)
					d.SetId(tfdirectconnect.GatewayAssociationCreateResourceID(directConnectGatewayID, transitGatewayID))

					sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Direct Connect Gateway Associations (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 Transit Gateways (%s): %w", region, err))
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Direct Connect Gateway Associations (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

// V0 state upgrade testing must be done via acceptance testing due to API call
func TestAccAwsDxGatewayAssociation_V0StateUpgrade(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_basicVpnGatewaySingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName, &ga, &gap),
					testAccCheckAwsDxGatewayAssociationStateUpgradeV0(resourceName),
				),
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_basicVpnGatewaySingleAccount(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_basicVpnGatewaySingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/28"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "virtualPrivateGateway"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "dx_gateway_owner_account_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsDxGatewayAssociationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_basicVpnGatewayCrossAccount(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_basicVpnGatewayCrossAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/28"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "virtualPrivateGateway"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					// dx_gateway_owner_account_id is the "awsalternate" provider's account ID.
					// acctest.CheckResourceAttrAccountID(resourceName, "dx_gateway_owner_account_id"),
				),
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_basicTransitGatewaySingleAccount(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameTgw := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_basicTransitGatewaySingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/30"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/30"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameTgw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "transitGateway"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "dx_gateway_owner_account_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsDxGatewayAssociationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_basicTransitGatewayCrossAccount(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameTgw := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_basicTransitGatewayCrossAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/30"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/30"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameTgw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "transitGateway"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					// dx_gateway_owner_account_id is the "awsalternate" provider's account ID.
					// acctest.CheckResourceAttrAccountID(resourceName, "dx_gateway_owner_account_id"),
				),
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_multiVpnGatewaysSingleAccount(t *testing.T) {
	resourceName1 := "aws_dx_gateway_association.test.0"
	resourceName2 := "aws_dx_gateway_association.test.1"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_multiVpnGatewaysSingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName1, &ga, &gap),
					testAccCheckAwsDxGatewayAssociationExists(resourceName2, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName1, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName1, "allowed_prefixes.*", "10.255.255.0/28"),
					resource.TestCheckResourceAttrSet(resourceName1, "dx_gateway_association_id"),
					resource.TestCheckResourceAttr(resourceName2, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName2, "allowed_prefixes.*", "10.255.255.16/28"),
					resource.TestCheckResourceAttrSet(resourceName2, "dx_gateway_association_id"),
				),
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_allowedPrefixesVpnGatewaySingleAccount(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_allowedPrefixesVpnGatewaySingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/30"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/30"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsDxGatewayAssociationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDxGatewayAssociationConfig_allowedPrefixesVpnGatewaySingleAccountUpdated(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/29"),
				),
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_allowedPrefixesVpnGatewayCrossAccount(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_allowedPrefixesVpnGatewayCrossAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/29"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
				),
				// Accepting the proposal with overridden prefixes changes the returned RequestedAllowedPrefixesToDirectConnectGateway value (allowed_prefixes attribute).
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccDxGatewayAssociationConfig_allowedPrefixesVpnGatewayCrossAccountUpdated(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/30"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/30"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
				),
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_recreateProposal(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dx_gateway_association.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga1, ga2 directconnect.GatewayAssociation
	var gap1, gap2 directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_basicVpnGatewayCrossAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName, &ga1, &gap1),
				),
			},
			{
				Config: testAccDxGatewayAssociationConfig_basicVpnGatewayCrossAccountUpdatedProposal(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName, &ga2, &gap2),
					testAccCheckAwsDxGatewayAssociationNotRecreated(&ga1, &ga2),
					testAccCheckAwsDxGatewayAssociationProposalRecreated(&gap1, &gap2),
				),
			},
		},
	})
}

func testAccAwsDxGatewayAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["dx_gateway_id"], rs.Primary.Attributes["associated_gateway_id"]), nil
	}
}

func testAccCheckAwsDxGatewayAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_gateway_association" {
			continue
		}

		_, err := finder.GatewayAssociationByID(conn, rs.Primary.Attributes["dx_gateway_association_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Direct Connect Gateway Association %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckAwsDxGatewayAssociationExists(name string, ga *directconnect.GatewayAssociation, gap *directconnect.GatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.Attributes["dx_gateway_association_id"] == "" {
			return fmt.Errorf("No Direct Connect Gateway Association ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).dxconn

		output, err := finder.GatewayAssociationByID(conn, rs.Primary.Attributes["dx_gateway_association_id"])

		if err != nil {
			return err
		}

		if proposalID := rs.Primary.Attributes["proposal_id"]; proposalID != "" {
			output, err := finder.GatewayAssociationProposalByID(conn, proposalID)

			if err != nil {
				return err
			}

			*gap = *output
		}

		*ga = *output

		return nil
	}
}

func testAccCheckAwsDxGatewayAssociationNotRecreated(old, new *directconnect.GatewayAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if old, new := aws.StringValue(old.AssociationId), aws.StringValue(new.AssociationId); old != new {
			return fmt.Errorf("Direct Connect Gateway Association (%s) recreated (%s)", old, new)
		}

		return nil
	}
}

// Perform check in acceptance testing as this StateUpgrader requires an API call
func testAccCheckAwsDxGatewayAssociationStateUpgradeV0(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		rawState := map[string]interface{}{
			"dx_gateway_id":  rs.Primary.Attributes["dx_gateway_id"],
			"vpn_gateway_id": rs.Primary.Attributes["associated_gateway_id"], // vpn_gateway_id was removed in 3.0, but older state still has it
		}

		updatedRawState, err := resourceAwsDxGatewayAssociationStateUpgradeV0(context.Background(), rawState, testAccProvider.Meta())

		if err != nil {
			return err
		}

		if got, want := updatedRawState["dx_gateway_association_id"], rs.Primary.Attributes["dx_gateway_association_id"]; got != want {
			return fmt.Errorf("Invalid dx_gateway_association_id attribute in migrated state. Expected %s, got %s", want, got)
		}

		return nil
	}
}

func testAccDxGatewayAssociationConfigBase_vpnGatewaySingleAccount(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_vpc" "test" {
  cidr_block = "10.255.255.0/28"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = aws_vpc.test.id
  vpn_gateway_id = aws_vpn_gateway.test.id
}
`, rName, rBgpAsn)
}

func testAccDxGatewayAssociationConfigBase_vpnGatewayCrossAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
# Creator
data "aws_caller_identity" "creator" {}

resource "aws_vpc" "test" {
  cidr_block = "10.255.255.0/28"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = aws_vpc.test.id
  vpn_gateway_id = aws_vpn_gateway.test.id
}

# Accepter
resource "aws_dx_gateway" "test" {
  provider = "awsalternate"

  amazon_side_asn = %[2]d
  name            = %[1]q
}
`, rName, rBgpAsn))
}

func testAccDxGatewayAssociationConfig_basicVpnGatewaySingleAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccDxGatewayAssociationConfigBase_vpnGatewaySingleAccount(rName, rBgpAsn),
		`
resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_vpn_gateway_attachment.test.vpn_gateway_id
}
`)
}

func testAccDxGatewayAssociationConfig_basicVpnGatewayCrossAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccDxGatewayAssociationConfigBase_vpnGatewayCrossAccount(rName, rBgpAsn),
		`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway_attachment.test.vpn_gateway_id
}

# Accepter
resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.creator.account_id
}
`)
}

func testAccDxGatewayAssociationConfig_basicVpnGatewayCrossAccountUpdatedProposal(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccDxGatewayAssociationConfigBase_vpnGatewayCrossAccount(rName, rBgpAsn),
		`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway_attachment.test.vpn_gateway_id
}

resource "aws_dx_gateway_association_proposal" "test2" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway_attachment.test.vpn_gateway_id
}

# Accepter
resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test2.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.creator.account_id
}
`)
}

func testAccDxGatewayAssociationConfig_basicTransitGatewaySingleAccount(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_ec2_transit_gateway.test.id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
`, rName, rBgpAsn)
}

func testAccDxGatewayAssociationConfig_basicTransitGatewayCrossAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
# Creator
data "aws_caller_identity" "creator" {}

# Accepter
resource "aws_dx_gateway" "test" {
  provider = "awsalternate"

  amazon_side_asn = %[2]d
  name            = %[1]q
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_ec2_transit_gateway.test.id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}

# Accepter
resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.creator.account_id
}
`, rName, rBgpAsn))
}

func testAccDxGatewayAssociationConfig_multiVpnGatewaysSingleAccount(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_vpc" "test" {
  count = 2

  cidr_block = cidrsubnet("10.255.255.0/26", 2, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  count = 2

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  count = 2

  vpc_id         = aws_vpc.test[count.index].id
  vpn_gateway_id = aws_vpn_gateway.test[count.index].id
}

resource "aws_dx_gateway_association" "test" {
  count = 2

  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_vpn_gateway_attachment.test[count.index].vpn_gateway_id
}
`, rName, rBgpAsn)
}

func testAccDxGatewayAssociationConfig_allowedPrefixesVpnGatewaySingleAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccDxGatewayAssociationConfigBase_vpnGatewaySingleAccount(rName, rBgpAsn),
		`
resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_vpn_gateway_attachment.test.vpn_gateway_id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
`)
}

func testAccDxGatewayAssociationConfig_allowedPrefixesVpnGatewaySingleAccountUpdated(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccDxGatewayAssociationConfigBase_vpnGatewaySingleAccount(rName, rBgpAsn),
		`
resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_vpn_gateway_attachment.test.vpn_gateway_id

  allowed_prefixes = [
    "10.255.255.8/29",
  ]
}
`)
}

func testAccDxGatewayAssociationConfig_allowedPrefixesVpnGatewayCrossAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccDxGatewayAssociationConfigBase_vpnGatewayCrossAccount(rName, rBgpAsn),
		`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway_attachment.test.vpn_gateway_id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}

# Accepter
resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.creator.account_id

  allowed_prefixes = [
    "10.255.255.8/29",
  ]
}
`)
}

func testAccDxGatewayAssociationConfig_allowedPrefixesVpnGatewayCrossAccountUpdated(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccDxGatewayAssociationConfigBase_vpnGatewayCrossAccount(rName, rBgpAsn),
		`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway_attachment.test.vpn_gateway_id
}

# Accepter
resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.creator.account_id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
`)
}
