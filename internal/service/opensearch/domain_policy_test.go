package opensearch_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opensearchservice"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccOpenSearchDomainPolicy_basic(t *testing.T) {
	var domain opensearchservice.DomainStatus
	ri := sdkacctest.RandInt()
	policy := `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "es:*",
            "Principal": "*",
            "Effect": "Allow",
            "Condition": {
                "IpAddress": {"aws:SourceIp": "127.0.0.1/32"}
            },
            "Resource": "${aws_opensearch_domain.test.arn}"
        }
    ]
}`
	expectedPolicyTpl := `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "es:*",
            "Principal": "*",
            "Effect": "Allow",
            "Condition": {
                "IpAddress": {"aws:SourceIp": "127.0.0.1/32"}
            },
            "Resource": "%s"
        }
    ]
}`
	name := fmt.Sprintf("tf-test-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainPolicyConfig_basic(ri, policy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists("aws_opensearch_domain.test", &domain),
					resource.TestCheckResourceAttr("aws_opensearch_domain.test", "engine_version", "OpenSearch_1.1"),
					func(s *terraform.State) error {
						awsClient := acctest.Provider.Meta().(*conns.AWSClient)
						expectedArn, err := buildDomainARN(name, awsClient.Partition, awsClient.AccountID, awsClient.Region)
						if err != nil {
							return err
						}
						expectedPolicy := fmt.Sprintf(expectedPolicyTpl, expectedArn)

						return testAccCheckPolicyMatch("aws_opensearch_domain_policy.test", "access_policies", expectedPolicy)(s)
					},
				),
			},
		},
	})
}

func testAccCheckPolicyMatch(resource, attr, expectedPolicy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		given, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("Attribute %q not found for %q", attr, resource)
		}

		areEquivalent, err := awspolicy.PoliciesAreEquivalent(given, expectedPolicy)
		if err != nil {
			return fmt.Errorf("Comparing AWS Policies failed: %s", err)
		}

		if !areEquivalent {
			return fmt.Errorf("AWS policies differ.\nGiven: %s\nExpected: %s", given, expectedPolicy)
		}

		return nil
	}
}

func buildDomainARN(name, partition, accId, region string) (string, error) {
	if partition == "" {
		return "", fmt.Errorf("Unable to construct OpenSearch Domain ARN because of missing AWS partition")
	}
	if accId == "" {
		return "", fmt.Errorf("Unable to construct OpenSearch Domain ARN because of missing AWS Account ID")
	}
	// arn:aws:es:us-west-2:187416307283:domain/example-name
	return fmt.Sprintf("arn:%s:es:%s:%s:domain/%s", partition, region, accId, name), nil
}

func testAccDomainPolicyConfig_basic(randInt int, policy string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = "tf-test-%d"
  engine_version = "OpenSearch_1.1"

  cluster_config {
    instance_type = "t2.small.search" # supported in both aws and aws-us-gov
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_opensearch_domain_policy" "test" {
  domain_name = aws_opensearch_domain.test.domain_name

  access_policies = <<POLICIES
%s
POLICIES
}
`, randInt, policy)
}
