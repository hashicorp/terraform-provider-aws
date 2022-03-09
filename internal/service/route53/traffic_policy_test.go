package route53_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
)

func TestAccTrafficPolicy_basic(t *testing.T) {
	var output route53.TrafficPolicySummary
	resourceName := "aws_route53_traffic_policy.test"
	rName := fmt.Sprintf("tf-route53-traffic-policy-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficPolicyDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccTrafficPolicyImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficPolicy_disappears(t *testing.T) {
	var output route53.TrafficPolicySummary
	resourceName := "aws_route53_traffic_policy.test"
	rName := fmt.Sprintf("tf-route53-traffic-policy-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficPolicyDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyExists(resourceName, &output),
					testAccCheckTrafficPolicyDisappears(&output),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTrafficPolicy_complete(t *testing.T) {
	var output route53.TrafficPolicySummary
	resourceName := "aws_route53_traffic_policy.test"
	rName := fmt.Sprintf("tf-route53-traffic-policy-%s", sdkacctest.RandString(5))
	comment := `comment`
	commentUpdated := `comment updated`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficPolicyDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfigComplete(rName, comment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "comment", comment),
				),
			},
			{
				Config: testAccTrafficPolicyConfigComplete(rName, commentUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "comment", commentUpdated),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccTrafficPolicyImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTrafficPolicyExists(resourceName string, trafficPolicy *route53.TrafficPolicySummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn

		resp, err := tfroute53.FindTrafficPolicyById(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("problem checking for traffic policy existence: %w", err)
		}

		if resp == nil {
			return fmt.Errorf("traffic policy %q does not exist", rs.Primary.ID)
		}

		*trafficPolicy = *resp

		return nil
	}
}

func testAccCheckTrafficPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_traffic_policy" {
			continue
		}

		resp, err := tfroute53.FindTrafficPolicyById(context.Background(), conn, rs.Primary.ID)
		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicy) || resp == nil {
			continue
		}

		if err != nil {
			return fmt.Errorf("error during check if traffic policy still exists, %#v", err)
		}
		if resp != nil {
			return fmt.Errorf("traffic Policy still exists")
		}
	}
	return nil
}

func testAccCheckTrafficPolicyDisappears(trafficPolicy *route53.TrafficPolicySummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn

		input := &route53.DeleteTrafficPolicyInput{
			Id:      trafficPolicy.Id,
			Version: trafficPolicy.LatestVersion,
		}

		_, err := conn.DeleteTrafficPolicyWithContext(context.Background(), input)

		return err
	}
}

func testAccTrafficPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["id"], rs.Primary.Attributes["version"]), nil
	}
}

func testAccTrafficPolicyConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_route53_traffic_policy" "test" {
  name     = %[1]q
  document = <<-EOT
{
    "AWSPolicyFormatVersion":"2015-10-01",
    "RecordType":"A",
    "Endpoints":{
        "endpoint-start-NkPh":{
            "Type":"value",
            "Value":"10.0.0.1"
        }
    },
    "StartEndpoint":"endpoint-start-NkPh"
}
EOT
}
`, name)
}

func testAccTrafficPolicyConfigComplete(name, comment string) string {
	return fmt.Sprintf(`
resource "aws_route53_traffic_policy" "test" {
  name     = %[1]q
  comment  = %[2]q
  document = <<-EOT
{
    "AWSPolicyFormatVersion":"2015-10-01",
    "RecordType":"A",
    "Endpoints":{
        "endpoint-start-NkPh":{
            "Type":"value",
            "Value":"10.0.0.1"
        }
    },
    "StartEndpoint":"endpoint-start-NkPh"
}
EOT
}
`, name, comment)
}
