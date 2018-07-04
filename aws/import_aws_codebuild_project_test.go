package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSCodeBuildProject_importBasic(t *testing.T) {
	resourceName := "aws_codebuild_project.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(rName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
