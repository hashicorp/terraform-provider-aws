package route53_test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
)

func TestAccRoute53DelegationSet_basic(t *testing.T) {
	refName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_delegation_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDelegationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSetConfig_basic(refName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSetExists(resourceName),
					acctest.MatchResourceAttrGlobalARNNoAccount(resourceName, "arn", "route53", regexp.MustCompile("delegationset/.+")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reference_name"},
			},
		},
	})
}

func TestAccRoute53DelegationSet_withZones(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	refName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_delegation_set.test"
	primaryZoneResourceName := "aws_route53_zone.primary"
	secondaryZoneResourceName := "aws_route53_zone.secondary"

	domain := acctest.RandomDomainName()
	zoneName1 := fmt.Sprintf("primary.%s", domain)
	zoneName2 := fmt.Sprintf("secondary.%s", domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDelegationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSetConfig_zones(refName, zoneName1, zoneName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSetExists(resourceName),
					testAccCheckZoneExists(primaryZoneResourceName, &zone),
					testAccCheckZoneExists(secondaryZoneResourceName, &zone),
					testAccCheckNameServersMatch(resourceName, primaryZoneResourceName),
					testAccCheckNameServersMatch(resourceName, secondaryZoneResourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reference_name"},
			},
		},
	})
}

func TestAccRoute53DelegationSet_disappears(t *testing.T) {
	refName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_delegation_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDelegationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSetConfig_basic(refName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSetExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53.ResourceDelegationSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDelegationSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_delegation_set" {
			continue
		}

		_, err := conn.GetReusableDelegationSet(&route53.GetReusableDelegationSetInput{Id: aws.String(rs.Primary.ID)})
		if err == nil {
			return fmt.Errorf("Delegation set still exists")
		}
	}
	return nil
}

func testAccCheckDelegationSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No delegation set ID is set")
		}

		out, err := conn.GetReusableDelegationSet(&route53.GetReusableDelegationSetInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("Delegation set does not exist: %#v", rs.Primary.ID)
		}

		setID := tfroute53.CleanDelegationSetID(*out.DelegationSet.Id)
		if setID != rs.Primary.ID {
			return fmt.Errorf("Delegation set ID does not match:\nExpected: %#v\nReturned: %#v", rs.Primary.ID, setID)
		}

		return nil
	}
}

func testAccCheckNameServersMatch(delegationSetName, zoneName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn

		delegationSetLocal, ok := s.RootModule().Resources[delegationSetName]
		if !ok {
			return fmt.Errorf("Not found: %s", delegationSetName)
		}
		delegationSet, err := conn.GetReusableDelegationSet(&route53.GetReusableDelegationSetInput{
			Id: aws.String(delegationSetLocal.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Delegation set does not exist: %#v", delegationSetLocal.Primary.ID)
		}

		hostedZoneLocal, ok := s.RootModule().Resources[zoneName]
		if !ok {
			return fmt.Errorf("Not found: %s", zoneName)
		}
		hostedZone, err := conn.GetHostedZone(&route53.GetHostedZoneInput{
			Id: aws.String(hostedZoneLocal.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Delegation set does not exist: %#v", hostedZoneLocal.Primary.ID)
		}

		if !reflect.DeepEqual(delegationSet.DelegationSet.NameServers, hostedZone.DelegationSet.NameServers) {
			return fmt.Errorf("Name servers do not match:\nDelegation Set: %#v\nHosted Zone:%#v",
				delegationSet.DelegationSet.NameServers, hostedZone.DelegationSet.NameServers)
		}

		return nil
	}
}

func testAccDelegationSetConfig_basic(refName string) string {
	return fmt.Sprintf(`
resource "aws_route53_delegation_set" "test" {
  reference_name = %[1]q
}
`, refName)
}

func testAccDelegationSetConfig_zones(refName, zoneName1, zoneName2 string) string {
	return fmt.Sprintf(`
resource "aws_route53_delegation_set" "test" {
  reference_name = %[1]q
}

resource "aws_route53_zone" "primary" {
  name              = %[2]q
  delegation_set_id = aws_route53_delegation_set.test.id
}

resource "aws_route53_zone" "secondary" {
  name              = %[3]q
  delegation_set_id = aws_route53_delegation_set.test.id
}
`, refName, zoneName1, zoneName2)
}
