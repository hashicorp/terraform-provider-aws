package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSEFSFileSystem_importBasic(t *testing.T) {
	resourceName := "aws_efs_file_system.foo-with-tags"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfigWithTags(rInt),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reference_name", "creation_token"},
			},
		},
	})
}
