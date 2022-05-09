package route53_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/route53"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccPreCheckTrafficPolicy(t *testing.T) {
	acctest.PreCheckPartitionHasService(route53.EndpointsID, t)

	if got, want := acctest.Partition(), endpoints.AwsUsGovPartitionID; got == want {
		t.Skipf("Route 53 Traffic Policies are not supported in %s partition", got)
	}
}

func TestAccRoute53TrafficPolicyInstance_basic(t *testing.T) {
	var v route53.TrafficPolicyInstance
	resourceName := "aws_route53_traffic_policy_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	zoneName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTrafficPolicy(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficPolicyInstanceDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyInstanceConfig(rName, zoneName, 3600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s.%s", rName, zoneName)),
					resource.TestCheckResourceAttr(resourceName, "ttl", "3600"),
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

func TestAccRoute53TrafficPolicyInstance_disappears(t *testing.T) {
	var v route53.TrafficPolicyInstance
	resourceName := "aws_route53_traffic_policy_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	zoneName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTrafficPolicy(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficPolicyInstanceDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyInstanceConfig(rName, zoneName, 360),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyInstanceExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53.ResourceTrafficPolicyInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53TrafficPolicyInstance_update(t *testing.T) {
	var v route53.TrafficPolicyInstance
	resourceName := "aws_route53_traffic_policy_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	zoneName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTrafficPolicy(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficPolicyInstanceDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyInstanceConfig(rName, zoneName, 3600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ttl", "3600"),
				),
			},
			{
				Config: testAccTrafficPolicyInstanceConfig(rName, zoneName, 7200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ttl", "7200"),
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

func testAccCheckTrafficPolicyInstanceExists(n string, v *route53.TrafficPolicyInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Traffic Policy Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn

		output, err := tfroute53.FindTrafficPolicyInstanceByID(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTrafficPolicyInstanceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_traffic_policy_instance" {
			continue
		}

		_, err := tfroute53.FindTrafficPolicyInstanceByID(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Route53 Traffic Policy Instance %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccTrafficPolicyInstanceConfig(rName, zoneName string, ttl int) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[2]q
}

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

resource "aws_route53_traffic_policy_instance" "test" {
  hosted_zone_id         = aws_route53_zone.test.zone_id
  name                   = "%[1]s.%[2]s"
  traffic_policy_id      = aws_route53_traffic_policy.test.id
  traffic_policy_version = aws_route53_traffic_policy.test.version
  ttl                    = %[3]d
}
`, rName, zoneName, ttl)
}
