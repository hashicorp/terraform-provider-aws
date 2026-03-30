// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSReservedInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := "RUN_RDS_RESERVED_INSTANCE_TESTS"
	vifId := os.Getenv(key)
	if vifId != acctest.CtTrue {
		t.Skipf("Environment variable %s is not set to true", key)
	}

	var reservation types.ReservedDBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_reserved_instance.test"
	dataSourceName := "data.aws_rds_reserved_instance_offering.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccReservedInstanceConfig_basic(rName, "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccReservedInstanceExists(ctx, t, resourceName, &reservation),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rds", regexache.MustCompile(`ri:.+`)),
					resource.TestCheckResourceAttrPair(dataSourceName, "currency_code", resourceName, "currency_code"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_instance_class", resourceName, "db_instance_class"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDuration, resourceName, names.AttrDuration),
					resource.TestCheckResourceAttrPair(dataSourceName, "fixed_price", resourceName, "fixed_price"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceCount, "1"),
					resource.TestCheckResourceAttrSet(resourceName, "lease_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az", resourceName, "multi_az"),
					resource.TestCheckResourceAttrPair(dataSourceName, "offering_id", resourceName, "offering_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "offering_type", resourceName, "offering_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "product_description", resourceName, "product_description"),
					resource.TestCheckResourceAttrSet(resourceName, "recurring_charges"),
					resource.TestCheckResourceAttr(resourceName, "reservation_id", rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStartTime),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttrSet(resourceName, "usage_price"),
				),
			},
		},
	})
}

func testAccReservedInstanceExists(ctx context.Context, t *testing.T, n string, v *types.ReservedDBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		output, err := tfrds.FindReservedDBInstanceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccReservedInstanceConfig_basic(rName string, instanceCount string) string {
	return fmt.Sprintf(`
data "aws_rds_reserved_instance_offering" "test" {
  db_instance_class   = "db.t2.micro"
  duration            = 31536000
  multi_az            = false
  offering_type       = "All Upfront"
  product_description = "mysql"
}

resource "aws_rds_reserved_instance" "test" {
  offering_id    = data.aws_rds_reserved_instance_offering.test.offering_id
  reservation_id = %[1]q
  instance_count = %[2]s
}
`, rName, instanceCount)
}
