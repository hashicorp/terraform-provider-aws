package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
)

func TestAccAWSGlueCatalogTable_importBasic(t *testing.T) {
	resourceName := "aws_glue_catalog_table.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCatalogTable_full(rInt, "A test table from terraform"),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

