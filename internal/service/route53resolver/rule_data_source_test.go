package route53resolver_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(route53resolver.EndpointsID, testAccErrorCheckSkipRoute53)
}

func TestAccRoute53ResolverRuleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.RandomDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_rule.test"
	ds1ResourceName := "data.aws_route53_resolver_rule.by_resolver_rule_id"
	ds2ResourceName := "data.aws_route53_resolver_rule.by_domain_name"
	ds3ResourceName := "data.aws_route53_resolver_rule.by_name_and_rule_type"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDataSourceConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
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

func TestAccRoute53ResolverRuleDataSource_resolverEndpointIdWithTags(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.RandomDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_rule.test"
	ds1ResourceName := "data.aws_route53_resolver_rule.by_resolver_endpoint_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDataSourceConfig_resolverEndpointIDTags(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
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

func TestAccRoute53ResolverRuleDataSource_sharedByMe(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.RandomDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_rule.test"
	ds1ResourceName := "data.aws_route53_resolver_rule.by_resolver_endpoint_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDataSourceConfig_sharedByMe(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttr(ds1ResourceName, "share_status", "SHARED_BY_ME"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttr(ds1ResourceName, "tags.%", "2"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.Key1", resourceName, "tags.Key1"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.Key2", resourceName, "tags.Key2"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRuleDataSource_sharedWithMe(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.RandomDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_rule.test"
	ds1ResourceName := "data.aws_route53_resolver_rule.by_resolver_endpoint_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDataSourceConfig_sharedWithMe(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttr(ds1ResourceName, "share_status", "SHARED_WITH_ME"),
					// Tags cannot be retrieved for rules shared with us.
					resource.TestCheckResourceAttr(ds1ResourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccRuleDataSourceConfig_basic(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[2]q
  rule_type   = "SYSTEM"
  name        = %[1]q
}

data "aws_route53_resolver_rule" "by_resolver_rule_id" {
  resolver_rule_id = aws_route53_resolver_rule.test.id
}

data "aws_route53_resolver_rule" "by_domain_name" {
  domain_name = aws_route53_resolver_rule.test.domain_name
}

data "aws_route53_resolver_rule" "by_name_and_rule_type" {
  name      = aws_route53_resolver_rule.test.name
  rule_type = aws_route53_resolver_rule.test.rule_type
}
`, rName, domainName)
}

func testAccRuleDataSourceConfig_resolverEndpointIDTags(rName, domainName string) string {
	return acctest.ConfigCompose(testAccRuleConfig_resolverEndpointBase(rName), fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[2]q
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.test[1].id

  target_ip {
    ip = "192.0.2.7"
  }

  tags = {
    "Key1" = "Value1"
    "Key2" = "Value2"
  }
}

data "aws_route53_resolver_rule" "by_resolver_endpoint_id" {
  resolver_endpoint_id = aws_route53_resolver_rule.test.resolver_endpoint_id
}
`, rName, domainName))
}

func testAccRuleDataSourceConfig_sharedByMe(rName, domainName string) string {
	return acctest.ConfigCompose(testAccRuleConfig_resolverEndpointBase(rName), acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[2]q
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.test[1].id

  target_ip {
    ip = "192.0.2.7"
  }

  tags = {
    "Key1" = "Value1"
    "Key2" = "Value2"
  }
}

resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = true
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_route53_resolver_rule.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_organizations_organization" "test" {}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_route53_resolver_rule" "by_resolver_endpoint_id" {
  resolver_endpoint_id = aws_route53_resolver_rule.test.resolver_endpoint_id

  depends_on = [aws_ram_resource_association.test, aws_ram_principal_association.test]
}
`, rName, domainName))
}

func testAccRuleDataSourceConfig_sharedWithMe(rName, domainName string) string {
	return acctest.ConfigCompose(testAccRuleConfig_resolverEndpointBase(rName), acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[2]q
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.test[1].id

  target_ip {
    ip = "192.0.2.7"
  }

  tags = {
    "Key1" = "Value1"
    "Key2" = "Value2"
  }
}

resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = true
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_route53_resolver_rule.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_organizations_organization" "test" {}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_route53_resolver_rule" "by_resolver_endpoint_id" {
  provider = "awsalternate"

  resolver_endpoint_id = aws_route53_resolver_rule.test.resolver_endpoint_id

  depends_on = [aws_ram_resource_association.test, aws_ram_principal_association.test]
}
`, rName, domainName))
}

// testAccErrorCheckSkipRoute53 skips Route53 tests that have error messages indicating unsupported features
func testAccErrorCheckSkipRoute53(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Operations related to PublicDNS",
		"Regional control plane current does not support",
		"NoSuchHostedZone: The specified hosted zone",
	)
}
