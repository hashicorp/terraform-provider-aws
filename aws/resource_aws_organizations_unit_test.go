package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccAwsOrganizationsUnit_importBasic(t *testing.T) {
	resourceName := "aws_organizations_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsUnitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsUnitConfig("foo"),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsOrganizationsUnit_basic(t *testing.T) {
	var unit organizations.OrganizationalUnit

	rInt := acctest.RandInt()
	name := fmt.Sprintf("tf_outest_%d", rInt)
	resourceName := "aws_organizations_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsUnitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsUnitConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsUnitExists(resourceName, &unit),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
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

func testAccAwsOrganizationsUnitUpdate(t *testing.T) {
	var unit organizations.OrganizationalUnit

	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("tf_outest_%d", rInt)
	name2 := fmt.Sprintf("tf_outest_%d", rInt+1)
	resourceName := "aws_organizations_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsUnitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsUnitConfig(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsUnitExists(resourceName, &unit),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
				),
			},
			{
				Config: testAccAwsOrganizationsUnitConfig(name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsUnitExists(resourceName, &unit),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
				),
			},
		},
	})
}

func testAccCheckAwsOrganizationsUnitDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_unit" {
			continue
		}

		exists, err := existsOrganization(conn)
		if err != nil {
			return fmt.Errorf("failed to check for the existance of an AWS Organization: %v", err)
		}

		if !exists {
			continue
		}

		params := &organizations.DescribeOrganizationalUnitInput{
			OrganizationalUnitId: &rs.Primary.ID,
		}

		resp, err := conn.DescribeOrganizationalUnit(params)

		if err != nil {
			if isAWSErr(err, organizations.ErrCodeOrganizationalUnitNotFoundException, "") {
				return nil
			}
			return err
		}

		if resp != nil && resp.OrganizationalUnit != nil {
			return fmt.Errorf("Bad: Organizational Unit still exists: %q", rs.Primary.ID)
		}
	}

	return nil

}

func existsOrganization(client *organizations.Organizations) (ok bool, err error) {
	_, err = client.DescribeOrganization(&organizations.DescribeOrganizationInput{})
	if err != nil {
		if isAWSErr(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
			err = nil
		}
		return
	}
	ok = true
	return
}

func testAccCheckAwsOrganizationsUnitExists(n string, ou *organizations.OrganizationalUnit) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).organizationsconn
		params := &organizations.DescribeOrganizationalUnitInput{
			OrganizationalUnitId: &rs.Primary.ID,
		}

		resp, err := conn.DescribeOrganizationalUnit(params)

		if err != nil {
			return err
		}

		if resp != nil || resp.OrganizationalUnit == nil {
			return fmt.Errorf("Organizational Unit %q does not exist", rs.Primary.ID)
		}

		ou = resp.OrganizationalUnit

		return nil
	}
}

func testAccAwsOrganizationsUnitConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "org" {
}

resource "aws_organizations_unit" "test" {
  parent_id = "${aws_organizations_organization.org.roots.0.id}"
  name = "%s"
}
`, name)
}
