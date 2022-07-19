package route53_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRoute53TrafficPolicy_basic(t *testing.T) {
	var v route53.TrafficPolicy
	resourceName := "aws_route53_traffic_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTrafficPolicy(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficPolicyDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrafficPolicyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "comment", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "A"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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

func TestAccRoute53TrafficPolicy_disappears(t *testing.T) {
	var v route53.TrafficPolicy
	resourceName := "aws_route53_traffic_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTrafficPolicy(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficPolicyDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53.ResourceTrafficPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53TrafficPolicy_update(t *testing.T) {
	var v route53.TrafficPolicy
	resourceName := "aws_route53_traffic_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	comment := `comment`
	commentUpdated := `comment updated`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTrafficPolicy(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficPolicyDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig_complete(rName, comment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "comment", comment),
				),
			},
			{
				Config: testAccTrafficPolicyConfig_complete(rName, commentUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyExists(resourceName, &v),
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

func testAccCheckTrafficPolicyExists(n string, v *route53.TrafficPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Traffic Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn

		output, err := tfroute53.FindTrafficPolicyByID(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTrafficPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_traffic_policy" {
			continue
		}

		_, err := tfroute53.FindTrafficPolicyByID(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Route53 Traffic Policy %s still exists", rs.Primary.ID)
	}
	return nil
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

func testAccTrafficPolicyConfig_basic(rName string) string {
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
`, rName)
}

func testAccTrafficPolicyConfig_complete(rName, comment string) string {
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
`, rName, comment)
}
