package aws

import (
	"testing"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"fmt"
)

func TestAccAWSGlueCatalogTable_full(t *testing.T) {
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogTable_basic(rInt),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists("aws_glue_catalog_table.test"),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_table.test",
						"name",
						fmt.Sprintf("my_test_catalog_table_%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_table.test",
						"description",
						"",
					),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_table.test",
						"owner",
						"",
					),
				),
			},
		},
	})
}
