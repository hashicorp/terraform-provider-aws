// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffms "github.com/hashicorp/terraform-provider-aws/internal/service/fms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, names.AttrARN, "fms", "policy/.+"),
					resource.TestCheckResourceAttr(resourceName, "delete_unused_fm_managed_resources", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func testAccPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffms.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPolicy_cloudFrontDistribution(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_cloudFrontDistribution(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.#", acctest.Ct1),
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

func testAccPolicy_includeMap(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_include(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.#", acctest.Ct1),
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

func testAccPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.#", acctest.Ct1),
				),
			},
			{
				Config: testAccPolicyConfig_updated(rName2, rName1),
			},
		},
	})
}

func testAccPolicy_policyOption(t *testing.T) {
	acctest.Skip(t, "PolicyOption not returned from AWS API")

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_policyOption(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, names.AttrARN, "fms", "policy/.+"),
					resource.TestCheckResourceAttr(resourceName, "delete_unused_fm_managed_resources", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.0.policy_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.0.policy_option.0.network_firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.0.policy_option.0.network_firewall_policy.0.firewall_deployment_model", "CENTRALIZED"),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.0.policy_option.0.third_party_firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.0.policy_option.0.third_party_firewall_policy.0.firewall_deployment_model", "DISTRIBUTED"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func testAccPolicy_resourceTags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_resourceTags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.key1", acctest.CtValue1),
				),
			},
			{
				Config: testAccPolicyConfig_resourceTags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.key1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.key2", acctest.CtValue2),
				),
			},
		},
	})
}

func testAccPolicy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccPolicyConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccPolicy_alb(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_alb(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceType, "ResourceTypeList"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_list.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_type_list.*", "AWS::ElasticLoadBalancingV2::LoadBalancer"),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.0.type", "WAFV2"),
					acctest.CheckResourceAttrJMES(resourceName, "security_service_policy_data.0.managed_service_data", names.AttrType, "WAFV2"),
					acctest.CheckResourceAttrJMES(resourceName, "security_service_policy_data.0.managed_service_data", "defaultAction.type", "ALLOW"),
				),
			},
		},
	})
}

func testAccPolicy_securityGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_securityGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceType, "AWS::EC2::SecurityGroup"),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_service_policy_data.0.type", "SECURITY_GROUPS_CONTENT_AUDIT"),
					acctest.CheckResourceAttrJMES(resourceName, "security_service_policy_data.0.managed_service_data", names.AttrType, "SECURITY_GROUPS_CONTENT_AUDIT"),
					acctest.CheckResourceAttrJMES(resourceName, "security_service_policy_data.0.managed_service_data", "securityGroupAction.type", "ALLOW"),
				),
			},
		},
	})
}

func testAccPolicy_rscSet(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_rscSet(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "resource_set_ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccCheckPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fms_policy" {
				continue
			}

			_, err := tffms.FindPolicyByID(ctx, conn, rs.Primary.ID)

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
}

func testAccCheckPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FMSClient(ctx)

		_, err := tffms.FindPolicyByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPolicyConfig_basic(policyName, ruleGroupName string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
  description           = "test description"
  remediation_enabled   = false
  resource_set_ids      = [aws_fms_resource_set.test.id]
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

resource "aws_fms_resource_set" "test" {
  depends_on = [aws_fms_admin_account.test]
  resource_set {
    name               = %[1]q
    resource_type_list = ["AWS::NetworkFirewall::Firewall"]
  }
}
`, policyName, ruleGroupName))
}

func testAccPolicyConfig_policyOption(policyName, ruleGroupName string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
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

    policy_option {
      network_firewall_policy {
        firewall_deployment_model = "CENTRALIZED"
      }

      third_party_firewall_policy {
        firewall_deployment_model = "DISTRIBUTED"
      }
    }
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = %[2]q
}
`, policyName, ruleGroupName))
}

func testAccPolicyConfig_cloudFrontDistribution(rName string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
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
          "sts:ExternalId" = data.aws_caller_identity.current.account_id
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
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = "aws-waf-logs-%[1]s"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName))
}

func testAccPolicyConfig_updated(policyName, ruleGroupName string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
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

func testAccPolicyConfig_include(rName string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
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
  name        = %[1]q
}
`, rName))
}

func testAccPolicyConfig_resourceTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  resource_tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = %[1]q
}
`, rName, tagKey1, tagValue1))
}

func testAccPolicyConfig_resourceTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  resource_tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = %[1]q
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccPolicyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = %[1]q
}
`, rName, tagKey1, tagValue1))
}

func testAccPolicyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = %[1]q
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccPolicyConfig_alb(rName string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
locals {
  msd = {
    type = "WAFV2"
    preProcessRuleGroups = [
      {
        overrideAction = {
          type = "NONE"
        }
        managedRuleGroupIdentifier = {
          vendorName           = "AWS"
          managedRuleGroupName = "AWSManagedRulesAmazonIpReputationList"
          version              = null
        }
        ruleGroupArn           = null
        ruleGroupType          = "ManagedRuleGroup"
        excludeRules           = null
        sampledRequestsEnabled = true
      },
      {
        overrideAction = {
          type = "NONE"
        },
        managedRuleGroupIdentifier = {
          vendorName           = "AWS"
          managedRuleGroupName = "AWSManagedRulesKnownBadInputsRuleSet"
          version              = "Version_1.1"
        },
        ruleGroupArn           = null
        ruleGroupType          = "ManagedRuleGroup"
        excludeRules           = null
        sampledRequestsEnabled = true
      }
    ]
    postProcessRuleGroups                   = []
    loggingConfiguration                    = null
    sampledRequestsEnabledForDefaultActions = true
    defaultAction = {
      type = "ALLOW"
    }
    overrideCustomerWebACLAssociation = true
  }
}

resource "aws_fms_policy" "test" {
  name = %[1]q
  resource_tags = {
    "disable" = ""
  }
  exclude_resource_tags       = true
  remediation_enabled         = true
  resource_type_list          = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]
  delete_all_policy_resources = false

  security_service_policy_data {
    type                 = "WAFV2"
    managed_service_data = jsonencode(local.msd)
  }

  depends_on = [aws_fms_admin_account.test]
}
`, rName))
}

func testAccPolicyConfig_securityGroup(rName string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = ""
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = ""
  }
}

resource "aws_fms_policy" "test" {
  name                        = %[1]q
  delete_all_policy_resources = false
  exclude_resource_tags       = false
  remediation_enabled         = false
  resource_type               = "AWS::EC2::SecurityGroup"

  security_service_policy_data {
    type = "SECURITY_GROUPS_CONTENT_AUDIT"

    managed_service_data = jsonencode({
      type = "SECURITY_GROUPS_CONTENT_AUDIT",

      securityGroupAction = {
        type = "ALLOW"
      },

      securityGroups = [
        {
          id = aws_security_group.test.id
        }
      ],
    })
  }

  depends_on = [aws_fms_admin_account.test]
}
`, rName))
}

func testAccPolicyConfig_rscSet(policyName, ruleGroupName string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_basic, fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
  description           = "test description"
  remediation_enabled   = false
  resource_set_ids      = [aws_fms_resource_set.test.id]
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

resource "aws_fms_resource_set" "test" {
  depends_on = [aws_fms_admin_account.test]
  resource_set {
    name               = %[1]q
    resource_type_list = ["AWS::NetworkFirewall::Firewall"]
  }
}
`, policyName, ruleGroupName))
}
