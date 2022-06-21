package cloudsearch_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudsearch"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudsearch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudsearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCloudSearchDomainServiceAccessPolicy_basic(t *testing.T) {
	resourceName := "aws_cloudsearch_domain_service_access_policy.test"
	rName := acctest.ResourcePrefix + "-" + sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccDomainServiceAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainServiceAccessPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainServiceAccessPolicyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_policy"),
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

func TestAccCloudSearchDomainServiceAccessPolicy_update(t *testing.T) {
	resourceName := "aws_cloudsearch_domain_service_access_policy.test"
	rName := acctest.ResourcePrefix + "-" + sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudsearch.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccDomainServiceAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainServiceAccessPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainServiceAccessPolicyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainServiceAccessPolicyConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainServiceAccessPolicyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_policy"),
				),
			},
		},
	})
}

func testAccDomainServiceAccessPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudSearch Domain Service Access Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudSearchConn

		_, err := tfcloudsearch.FindAccessPolicyByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccDomainServiceAccessPolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudsearch_domain_service_access_policy" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudSearchConn

		_, err := tfcloudsearch.FindAccessPolicyByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudSearch Domain Service Access Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccDomainServiceAccessPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q
}

resource "aws_cloudsearch_domain_service_access_policy" "test" {
  domain_name = aws_cloudsearch_domain.test.id

  access_policy = <<POLICY
{
  "Version":"2012-10-17",
  "Statement":[{
    "Sid":"search_and_document",
    "Effect":"Allow",
    "Principal":"*",
    "Action":[
      "cloudsearch:search",
      "cloudsearch:document"
    ],
    "Condition":{"IpAddress":{"aws:SourceIp":"192.0.2.0/32"}}
  }]
}
POLICY
}
`, rName)
}

func testAccDomainServiceAccessPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q
}

resource "aws_cloudsearch_domain_service_access_policy" "test" {
  domain_name = aws_cloudsearch_domain.test.id

  access_policy = <<POLICY
{
  "Version":"2012-10-17",
  "Statement":[{
    "Sid":"all",
    "Effect":"Allow",
    "Action":"cloudsearch:*",
    "Principal":"*"
  }]
}
POLICY
}
`, rName)
}
