package organizations_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
)

func testAccOrganization_basic(t *testing.T) {
	var organization organizations.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "accounts.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "accounts.0.arn", resourceName, "master_account_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "accounts.0.email", resourceName, "master_account_email"),
					resource.TestCheckResourceAttrPair(resourceName, "accounts.0.id", resourceName, "master_account_id"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "organizations", regexp.MustCompile(`organization/o-.+`)),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "feature_set", organizations.OrganizationFeatureSetAll),
					acctest.MatchResourceAttrGlobalARN(resourceName, "master_account_arn", "organizations", regexp.MustCompile(`account/o-.+/.+`)),
					resource.TestMatchResourceAttr(resourceName, "master_account_email", regexp.MustCompile(`.+@.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttr(resourceName, "non_master_accounts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "roots.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "roots.0.id", regexp.MustCompile(`r-[a-z0-9]{4,32}`)),
					resource.TestCheckResourceAttrSet(resourceName, "roots.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "roots.0.arn"),
					resource.TestCheckResourceAttr(resourceName, "roots.0.policy_types.#", "0"),
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

func testAccOrganization_serviceAccessPrincipals(t *testing.T) {
	var organization organizations.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_serviceAccessPrincipals1("config.amazonaws.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "config.amazonaws.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfig_serviceAccessPrincipals2("config.amazonaws.com", "ds.amazonaws.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "config.amazonaws.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "ds.amazonaws.com"),
				),
			},
			{
				Config: testAccOrganizationConfig_serviceAccessPrincipals1("fms.amazonaws.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "fms.amazonaws.com"),
				),
			},
		},
	})
}

func testAccOrganization_EnabledPolicyTypes(t *testing.T) {
	var organization organizations.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(organizations.PolicyTypeServiceControlPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", organizations.PolicyTypeServiceControlPolicy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "0"),
				),
			},
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(organizations.PolicyTypeAiservicesOptOutPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", organizations.PolicyTypeAiservicesOptOutPolicy),
				),
			},
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(organizations.PolicyTypeServiceControlPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", organizations.PolicyTypeServiceControlPolicy),
				),
			},
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(organizations.PolicyTypeBackupPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", organizations.PolicyTypeBackupPolicy),
				),
			},
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(organizations.PolicyTypeTagPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", organizations.PolicyTypeTagPolicy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "0"),
				),
			},
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(organizations.PolicyTypeTagPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
				),
			},
		},
	})
}

func testAccOrganization_FeatureSet(t *testing.T) {
	var organization organizations.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_featureSet(organizations.OrganizationFeatureSetConsolidatedBilling),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "feature_set", organizations.OrganizationFeatureSetConsolidatedBilling),
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

func testAccOrganization_FeatureSetForcesNew(t *testing.T) {
	var beforeValue, afterValue organizations.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_featureSet(organizations.OrganizationFeatureSetAll),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &beforeValue),
					resource.TestCheckResourceAttr(resourceName, "feature_set", organizations.OrganizationFeatureSetAll),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfig_featureSet(organizations.OrganizationFeatureSetConsolidatedBilling),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &afterValue),
					resource.TestCheckResourceAttr(resourceName, "feature_set", organizations.OrganizationFeatureSetConsolidatedBilling),
					testAccOrganizationRecreated(&beforeValue, &afterValue),
				),
			},
		},
	})
}

func testAccOrganization_FeatureSetUpdate(t *testing.T) {
	var beforeValue, afterValue organizations.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_featureSet(organizations.OrganizationFeatureSetConsolidatedBilling),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &beforeValue),
					resource.TestCheckResourceAttr(resourceName, "feature_set", organizations.OrganizationFeatureSetConsolidatedBilling),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccOrganizationConfig_featureSet(organizations.OrganizationFeatureSetAll),
				ExpectNonEmptyPlan: true, // See note below on this perpetual difference
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(resourceName, &afterValue),
					// The below check cannot be performed here because the user must confirm the change
					// via Console. Until then, the FeatureSet will not actually be toggled to ALL
					// and will continue to show as CONSOLIDATED_BILLING when calling DescribeOrganization
					// resource.TestCheckResourceAttr(resourceName, "feature_set", organizations.OrganizationFeatureSetAll),
					testAccOrganizationNotRecreated(&beforeValue, &afterValue),
				),
			},
		},
	})
}

func testAccCheckOrganizationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_organization" {
			continue
		}

		params := &organizations.DescribeOrganizationInput{}

		resp, err := conn.DescribeOrganization(params)

		if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAWSOrganizationsNotInUseException) {
			return nil
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.Organization != nil {
			return fmt.Errorf("Bad: Organization still exists: %q", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckOrganizationExists(n string, org *organizations.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Organization ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn
		params := &organizations.DescribeOrganizationInput{}

		resp, err := conn.DescribeOrganization(params)

		if err != nil {
			return err
		}

		if resp == nil || resp.Organization == nil {
			return fmt.Errorf("Organization %q does not exist", rs.Primary.ID)
		}

		*org = *resp.Organization

		return nil
	}
}

const testAccOrganizationConfig_basic = "resource \"aws_organizations_organization\" \"test\" {}"

func testAccOrganizationConfig_serviceAccessPrincipals1(principal1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = [%q]
}
`, principal1)
}

func testAccOrganizationConfig_serviceAccessPrincipals2(principal1, principal2 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = [%q, %q]
}
`, principal1, principal2)
}

func testAccOrganizationConfig_enabledPolicyTypes1(policyType1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  enabled_policy_types = [%[1]q]
}
`, policyType1)
}

func testAccOrganizationConfig_featureSet(featureSet string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  feature_set = %q
}
`, featureSet)
}

func TestFlattenRoots(t *testing.T) {
	roots := []*organizations.Root{
		{
			Name: aws.String("Root1"),
			Arn:  aws.String("arn:1"),
			Id:   aws.String("r-1"),
			PolicyTypes: []*organizations.PolicyTypeSummary{
				{
					Status: aws.String("ENABLED"),
					Type:   aws.String("SERVICE_CONTROL_POLICY"),
				},
				{
					Status: aws.String("DISABLED"),
					Type:   aws.String("SERVICE_CONTROL_POLICY"),
				},
			},
		},
	}
	result := tforganizations.FlattenRoots(roots)

	if len(result) != len(roots) {
		t.Fatalf("expected result to have %d elements, got %d", len(roots), len(result))
	}

	for i, r := range roots {
		if aws.StringValue(r.Name) != result[i]["name"] {
			t.Fatalf(`expected result[%d]["name"] to equal %q, got %q`, i, aws.StringValue(r.Name), result[i]["name"])
		}
		if aws.StringValue(r.Arn) != result[i]["arn"] {
			t.Fatalf(`expected result[%d]["arn"] to equal %q, got %q`, i, aws.StringValue(r.Arn), result[i]["arn"])
		}
		if aws.StringValue(r.Id) != result[i]["id"] {
			t.Fatalf(`expected result[%d]["id"] to equal %q, got %q`, i, aws.StringValue(r.Id), result[i]["id"])
		}
		if result[i]["policy_types"] == nil {
			continue
		}
		if types, ok := result[i]["policy_types"].([]map[string]interface{}); ok {
			testFlattenRootPolicyTypes(t, i, types, r.PolicyTypes)
			continue
		}
		t.Fatalf(`result[%d]["policy_types"] could not be converted to []map[string]interface{}`, i)
	}
}

func testFlattenRootPolicyTypes(t *testing.T, index int, result []map[string]interface{}, types []*organizations.PolicyTypeSummary) {
	if len(result) != len(types) {
		t.Fatalf(`expected result[%d]["policy_types"] to have %d elements, got %d`, index, len(types), len(result))
	}
	for i, v := range types {
		if aws.StringValue(v.Status) != result[i]["status"] {
			t.Fatalf(`expected result[%d]["policy_types"][%d]["status"] to equal %q, got %q`, index, i, aws.StringValue(v.Status), result[i]["status"])
		}
		if aws.StringValue(v.Type) != result[i]["type"] {
			t.Fatalf(`expected result[%d]["policy_types"][%d]["type"] to equal %q, got %q`, index, i, aws.StringValue(v.Type), result[i]["type"])
		}
	}
}

func testAccOrganizationRecreated(before, after *organizations.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(before.Id) == aws.StringValue(after.Id) {
			return fmt.Errorf("Organization (%s) not recreated", aws.StringValue(before.Id))
		}
		return nil
	}
}

func testAccOrganizationNotRecreated(before, after *organizations.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(before.Id) != aws.StringValue(after.Id) {
			return fmt.Errorf("Organization (%s) recreated", aws.StringValue(before.Id))
		}
		return nil
	}
}
