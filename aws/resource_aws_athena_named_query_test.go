package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAthenaNamedQuery(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaNamedQueryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaNamedQueryConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaNamedQueryExists("aws_athena_named_query.foo"),
				),
			},
		},
	})
}

func testAccCheckAWSAthenaNamedQueryDestroy(s *terraform.State) error {
	return nil
}

func testAccCheckAWSAthenaNamedQueryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

const testAccAthenaNamedQueryConfig = `
`
