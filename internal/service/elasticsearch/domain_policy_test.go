// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch_test

import (
	"fmt"
	"testing"

	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElasticsearchDomainPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domain elasticsearch.ElasticsearchDomainStatus
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
            "Resource": "${aws_elasticsearch_domain.example.arn}"
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainPolicyConfig_basic(ri, policy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_elasticsearch_domain.example", &domain),
					resource.TestCheckResourceAttr("aws_elasticsearch_domain.example", "elasticsearch_version", "2.3"),
					func(s *terraform.State) error {
						awsClient := acctest.Provider.Meta().(*conns.AWSClient)
						expectedArn, err := buildDomainARN(name, awsClient.Partition, awsClient.AccountID, awsClient.Region)
						if err != nil {
							return err
						}
						expectedPolicy := fmt.Sprintf(expectedPolicyTpl, expectedArn)

						return testAccCheckPolicyMatch("aws_elasticsearch_domain_policy.main", "access_policies", expectedPolicy)(s)
					},
				),
			},
		},
	})
}

func buildDomainARN(name, partition, accId, region string) (string, error) {
	if partition == "" {
		return "", fmt.Errorf("Unable to construct ES Domain ARN because of missing AWS partition")
	}
	if accId == "" {
		return "", fmt.Errorf("Unable to construct ES Domain ARN because of missing AWS Account ID")
	}
	// arn:aws:es:us-west-2:187416307283:domain/example-name
	return fmt.Sprintf("arn:%s:es:%s:%s:domain/%s", partition, region, accId, name), nil
}

func testAccDomainPolicyConfig_basic(randInt int, policy string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "example" {
  domain_name           = "tf-test-%d"
  elasticsearch_version = "2.3"

  cluster_config {
    instance_type = "t2.small.elasticsearch" # supported in both aws and aws-us-gov
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_elasticsearch_domain_policy" "main" {
  domain_name = aws_elasticsearch_domain.example.domain_name

  access_policies = <<POLICIES
%s
POLICIES
}
`, randInt, policy)
}
