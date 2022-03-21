package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSRoute53TrafficPolicy_basic(t *testing.T) {
	policyName := fmt.Sprintf("policy-%s", acctest.RandString(8))
	resourceName := "aws_route53_traffic_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53TrafficPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53TrafficPolicyConfig(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
				),
			},
		},
	})
}

func testAccRoute53TrafficPolicyConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_route53_traffic_policy" "test" {
	name     = "%s"
	comment  = "comment"
	document = "{\"AWSPolicyFormatVersion\":\"2015-10-01\",\"RecordType\":\"A\",\"Endpoints\":{\"endpoint-start-NkPh\":{\"Type\":\"value\",\"Value\":\"10.0.0.1\"}},\"StartEndpoint\":\"endpoint-start-NkPh\"}"
}
`, name)
}

func testAccCheckRoute53TrafficPolicyDestroy(s *terraform.State) error {
	return testAccCheckRoute53TrafficPolicyDestroyWithProvider(s, testAccProvider)
}

func testAccCheckRoute53TrafficPolicyDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).r53conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_traffic_policy" {
			continue
		}
		tp, err := getTrafficPolicyById(rs.Primary.ID, conn)
		if err != nil {
			return fmt.Errorf("Error during check if traffic policy still exists, %#v", err)
		}
		if tp != nil {
			return fmt.Errorf("Traffic Policy still exists")
		}
	}
	return nil
}
