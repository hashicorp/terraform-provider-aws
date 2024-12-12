// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRRepositoriesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rNames []string
	for i := 0; i < 5; i++ {
		rNames = append(rNames, sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
	}
	dataSourceName := "data.aws_ecr_repositories.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECREndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(ecr.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoriesDataSourceConfig_basic(rNames),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "names.#", 5),
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

data "aws_ecr_repositories" "test" {
  depends_on = [aws_ecr_repository.test]
}
`, rNameString)
}
