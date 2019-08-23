package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsRoute53ResolverRule_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-testacc-r53-resolver-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "aws_route53_resolver_rule.example"
	ds1ResourceName := "data.aws_route53_resolver_rule.by_resolver_rule_id"
	ds2ResourceName := "data.aws_route53_resolver_rule.by_domain_name"
	ds3ResourceName := "data.aws_route53_resolver_rule.by_name_and_rule_type"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ResolverRule_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.%", resourceName, "tags.%"),

					resource.TestCheckResourceAttrPair(ds2ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "tags.%", resourceName, "tags.%"),

					resource.TestCheckResourceAttrPair(ds3ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRoute53ResolverRule_ResolverEndpointIdWithTags(t *testing.T) {
	rName := fmt.Sprintf("tf-testacc-r53-resolver-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "aws_route53_resolver_rule.example"
	ds1ResourceName := "data.aws_route53_resolver_rule.by_resolver_endpoint_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ResolverRule_resolverEndpointIdWithTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttr(ds1ResourceName, "tags.%", "2"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.Key1", resourceName, "tags.Key1"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.Key2", resourceName, "tags.Key2"),
				),
			},
		},
	})
}

func testAccDataSourceAwsRoute53ResolverRule_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "example" {
  domain_name = "%[1]s.example.com"
  rule_type   = "SYSTEM"
  name        = %[1]q
}

data "aws_route53_resolver_rule" "by_resolver_rule_id" {
  resolver_rule_id = "${aws_route53_resolver_rule.example.id}"
}

data "aws_route53_resolver_rule" "by_domain_name" {
  domain_name = "${aws_route53_resolver_rule.example.domain_name}"
}

data "aws_route53_resolver_rule" "by_name_and_rule_type" {
  name      = "${aws_route53_resolver_rule.example.name}"
  rule_type = "${aws_route53_resolver_rule.example.rule_type}"
}
`, rName)
}

func testAccDataSourceAwsRoute53ResolverRule_resolverEndpointIdWithTags(rName string) string {
	return testAccRoute53ResolverRuleConfig_resolverEndpoint(rName) + fmt.Sprintf(`
resource "aws_route53_resolver_rule" "example" {
  domain_name = "%[1]s.example.com"
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = "${aws_route53_resolver_endpoint.bar.id}"

  target_ip {
    ip = "192.0.2.7"
  }

  tags = {
    "Key1" = "Value1"
    "Key2" = "Value2"
  }
}

data "aws_route53_resolver_rule" "by_resolver_endpoint_id" {
  resolver_endpoint_id = "${aws_route53_resolver_rule.example.resolver_endpoint_id}"
}
`, rName)
}
