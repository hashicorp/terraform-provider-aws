package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSEcrRepository_importBasic(t *testing.T) {
	resourceName := "aws_ecr_repository.default"
	randString := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSEcrRepository(randString),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
