package ecr_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccECRRepositoriesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rNames []string
	for i := 1; i < 6; i++ {
		rNames = append(rNames, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
	}
	dataSourceName := "data.aws_ecr_repositories.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(ecr.ServiceID))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(ecr.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoriesDataSourceConfig_basic(rNames),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "5"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "names.*", rNames[0]),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "names.*", rNames[1]),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "names.*", rNames[2]),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "names.*", rNames[3]),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "names.*", rNames[4]),
				),
			},
		},
	})
}

func TestAccECRRepositoriesDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecr_repositories.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(ecr.ServiceID))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(ecr.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoriesDataSourceConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "0"),
				),
			},
		},
	})
}

func testAccRepositoriesDataSourceConfig_basic(rNames []string) string {
	rNameJson, _ := json.Marshal(rNames)
	rNameString := string(rNameJson)
	return fmt.Sprintf(`
locals {
	repo_list = %[1]s
}

resource "aws_ecr_repository" "test" {
	count = length(local.repo_list)
	name  = local.repo_list[count.index]
}
  
  data "aws_ecr_repositories" "test" {}
`, rNameString)
}

func testAccRepositoriesDataSourceConfig_empty() string {
	return fmt.Sprint(`
data "aws_ecr_repositories" "test" {}
`)
}
