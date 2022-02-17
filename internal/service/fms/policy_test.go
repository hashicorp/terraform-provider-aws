package fms_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/fms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffms "github.com/hashicorp/terraform-provider-aws/internal/service/fms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccPolicy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckFmsAdmin(t)
			// acctest.PreCheckOrganizationsAccount(t)
			acctest.PreCheckOrganizationsEnabled(t)
			acctest.PreCheckOrganizationManagementAccount(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, fms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName),
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, "arn", "fms", "policy/.+"),
					resource.TestCheckResourceAttr(resourceName, "delete_unused_fm_managed_resources", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_update_token", "delete_all_policy_resources"},
			},
		},
	})
}

func testAccPolicy_cloudFrontDistribution(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", sdkacctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckFmsAdmin(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, fms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig_cloudfrontDistribution(fmsPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists("aws_fms_policy.test"),
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount("aws_fms_policy.test", "arn", "fms", "policy/.+"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "security_service_policy_data.#", "1"),
				),
			},
			{
				ResourceName:            "aws_fms_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_update_token", "delete_all_policy_resources"},
			},
		},
	})
}

func testAccPolicy_includeMap(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", sdkacctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", sdkacctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckFmsAdmin(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, fms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig_include(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists("aws_fms_policy.test"),
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount("aws_fms_policy.test", "arn", "fms", "policy/.+"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "security_service_policy_data.#", "1"),
				),
			},
			{
				ResourceName:            "aws_fms_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_update_token", "delete_all_policy_resources"},
			},
		},
	})
}

func testAccPolicy_update(t *testing.T) {
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckFmsAdmin(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, fms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig(rName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", rName1),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "security_service_policy_data.#", "1"),
				),
			},
			{
				Config: testAccFmsPolicyConfig_updated(rName2, rName1),
			},
		},
	})
}

func testAccPolicy_tags(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", sdkacctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", sdkacctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckFmsAdmin(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, fms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig_tags(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists("aws_fms_policy.test"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.%", "2"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.Usage", "original"),
				),
			},
			{
				Config: testAccFmsPolicyConfig_tagsChanged(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists("aws_fms_policy.test"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.%", "1"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.Usage", "changed"),
				),
			},
		},
	})
}

func testAccCheckPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FMSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fms_policy" {
			continue
		}

		_, err := tffms.FindPolicyByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("FMS Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No FMS Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FMSConn

		_, err := tffms.FindPolicyByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccFmsPolicyConfigOrgMgmtAccountBase() string {
	return acctest.ConfigCompose(testAccFmsAdminRegionProviderConfig(), `
data "aws_caller_identity" "current" {}

resource "aws_fms_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`)
}

func testAccFmsPolicyConfigBase() string {
	return acctest.ConfigCompose(testAccFmsAdminRegionProviderConfig(), `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["fms.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_fms_admin_account" "test" {
  account_id = aws_organizations_organization.test.master_account_id
}
`)
}

func testAccFmsPolicyConfig(policyName, ruleGroupName string) string {
	// return acctest.ConfigCompose(testAccFmsPolicyConfigBase(), fmt.Sprintf(`
	return acctest.ConfigCompose(testAccFmsPolicyConfigOrgMgmtAccountBase(), fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  exclude_map {
    account = [data.aws_caller_identity.current.account_id]
  }

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = %[2]q
}
`, policyName, ruleGroupName))
}

func testAccFmsPolicyConfig_cloudfrontDistribution(name string) string {
	return acctest.ConfigCompose(
		testAccFmsPolicyConfigBase(),
		fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = false
  resource_type         = "AWS::CloudFront::Distribution"

  security_service_policy_data {
    type                 = "WAFV2"
    managed_service_data = "{\"type\":\"WAFV2\",\"preProcessRuleGroups\":[{\"ruleGroupArn\":null,\"overrideAction\":{\"type\":\"NONE\"},\"managedRuleGroupIdentifier\":{\"version\":null,\"vendorName\":\"AWS\",\"managedRuleGroupName\":\"AWSManagedRulesAmazonIpReputationList\"},\"ruleGroupType\":\"ManagedRuleGroup\",\"excludeRules\":[]}],\"postProcessRuleGroups\":[],\"defaultAction\":{\"type\":\"ALLOW\"},\"overrideCustomerWebACLAssociation\":false,\"loggingConfiguration\":{\"logDestinationConfigs\":[\"${aws_kinesis_firehose_delivery_stream.test.arn}\"],\"redactedFields\":[{\"redactedFieldType\":\"SingleHeader\",\"redactedFieldValue\":\"Cookies\"}]}}"
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Condition = {
        StringEquals = {
          "sts:ExternalId" = "${data.aws_caller_identity.current.account_id}"
        }
      }
      Effect = "Allow"
      Principal = {
        Service = "firehose.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = %[1]q

  inline_policy {
    name = "test"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [{
        Action = [
          "s3:AbortMultipartUpload",
          "s3:GetBucketLocation",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
          "s3:PutObject",
        ]
        Effect = "Allow"
        Resource = [
          aws_s3_bucket.test.arn,
          "${aws_s3_bucket.test.arn}/*"
        ]
        Sid = ""
      }]
    })
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = "private"
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = "aws-waf-logs-%[1]s"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, name))
}

func testAccFmsPolicyConfig_updated(policyName, ruleGroupName string) string {
	return acctest.ConfigCompose(testAccFmsPolicyConfigBase(), fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
  remediation_enabled   = true
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  exclude_map {
    account = [data.aws_caller_identity.current.account_id]
  }

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"ALLOW\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  lifecycle {
    create_before_destroy = false
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = %[2]q
}
`, policyName, ruleGroupName))
}

func testAccFmsPolicyConfig_include(name string, group string) string {
	return acctest.ConfigCompose(
		testAccFmsPolicyConfigBase(),
		fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  include_map {
    account = [data.aws_caller_identity.current.account_id]
  }

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group))
}

func testAccFmsPolicyConfig_tags(name string, group string) string {
	return acctest.ConfigCompose(
		testAccFmsPolicyConfigBase(),
		fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  resource_tags = {
    Environment = "Testing"
    Usage       = "original"
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group))
}

func testAccFmsPolicyConfig_tagsChanged(name string, group string) string {
	return acctest.ConfigCompose(
		testAccFmsPolicyConfigBase(),
		fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  resource_tags = {
    Usage = "changed"
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group))
}
