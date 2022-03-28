package opensearch_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opensearchservice"
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
            "Action": "opensearch:*",
            "Principal": "*",
            "Effect": "Allow",
            "Condition": {
                "IpAddress": {"aws:SourceIp": "127.0.0.1/32"}
            },
            "Resource": aws_opensearch_domain.example.arn
        }
    ]
}`
	expectedPolicyTpl := `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "opensearch:*",
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainPolicyConfig(ri, policy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists("aws_opensearch_domain.example", &domain),
					resource.TestCheckResourceAttr("aws_opensearch_domain.example", "engine_version", "OpenSearch_1.1"),
					func(s *terraform.State) error {
						awsClient := acctest.Provider.Meta().(*conns.AWSClient)
						expectedArn, err := buildESDomainArn(name, awsClient.Partition, awsClient.AccountID, awsClient.Region)
						if err != nil {
							return err
						}
						expectedPolicy := fmt.Sprintf(expectedPolicyTpl, expectedArn)

						return testAccCheckPolicyMatch("aws_opensearch_domain_policy.main", "access_policies", expectedPolicy)(s)
					},
				),
			},
		},
	})
}

func buildESDomainArn(name, partition, accId, region string) (string, error) {
	if partition == "" {
		return "", fmt.Errorf("Unable to construct ES Domain ARN because of missing AWS partition")
	}
	if accId == "" {
		return "", fmt.Errorf("Unable to construct ES Domain ARN because of missing AWS Account ID")
	}
	// arn:aws:opensearch:us-west-2:187416307283:domain/example-name
	return fmt.Sprintf("arn:%s:opensearch:%s:%s:domain/%s", partition, region, accId, name), nil
}

func testAccDomainPolicyConfig(randInt int, policy string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "example" {
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

resource "aws_opensearch_domain_policy" "main" {
  domain_name = aws_opensearch_domain.example.domain_name

  access_policies = <<POLICIES
%s
POLICIES
}
`, randInt, policy)
}
