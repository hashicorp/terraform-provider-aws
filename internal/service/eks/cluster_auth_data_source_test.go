// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEKSClusterAuthDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceResourceName := "data.aws_eks_cluster_auth.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterAuthDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceResourceName, names.AttrName, "foobar"),
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "token"),
					testAccCheckClusterAuthToken(dataSourceResourceName),
				),
			},
		},
	})
}

func testAccCheckClusterAuthToken(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		name := rs.Primary.Attributes[names.AttrName]
		tok := rs.Primary.Attributes["token"]
		verifier := tfeks.NewVerifier(name)
		identity, err := verifier.Verify(tok)
		if err != nil {
			return fmt.Errorf("Error verifying token for cluster %q: %v", name, err)
		}
		if identity.ARN == "" {
			return fmt.Errorf("Unexpected blank ARN for token identity")
		}

		return nil
	}
}

const testAccClusterAuthDataSourceConfig_basic = `
data "aws_eks_cluster_auth" "test" {
  name = "foobar"
}
`
