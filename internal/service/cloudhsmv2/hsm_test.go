// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudhsmv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/cloudhsmv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudhsmv2 "github.com/hashicorp/terraform-provider-aws/internal/service/cloudhsmv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccHSM_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudhsm_v2_hsm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudHSMV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHSMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHSMConfig_subnetID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHSMExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, "aws_subnet.test.0", names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_id", "aws_cloudhsm_v2_cluster.test", names.AttrID),
					resource.TestMatchResourceAttr(resourceName, "hsm_eni_id", regexache.MustCompile(`^eni-.+`)),
					resource.TestMatchResourceAttr(resourceName, "hsm_id", regexache.MustCompile(`^hsm-.+`)),
					resource.TestCheckResourceAttr(resourceName, "hsm_state", string(types.HsmStateActive)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrIPAddress),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, "aws_subnet.test.0", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccHSM_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudhsm_v2_hsm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudHSMV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHSMConfig_subnetID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_cloudhsm_v2_cluster.test"),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudhsmv2.ResourceHSM(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccHSM_AvailabilityZone(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudhsm_v2_hsm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudHSMV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHSMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHSMConfig_availabilityZone(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHSMExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, "aws_subnet.test.0", names.AttrAvailabilityZone),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccHSM_IPAddress(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudhsm_v2_hsm.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudHSMV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHSMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHSMConfig_ipAddress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHSMExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddress, "10.0.0.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckHSMDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudHSMV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudhsm_v2_hsm" {
				continue
			}

			_, err := tfcloudhsmv2.FindHSMByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["hsm_eni_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudHSMv2 HSM %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckHSMExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudHSMV2Client(ctx)

		_, err := tfcloudhsmv2.FindHSMByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["hsm_eni_id"])

		return err
	}
}

func testAccHSMConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccHSMConfig_availabilityZone(rName string) string {
	return acctest.ConfigCompose(testAccHSMConfig_base(rName), `
resource "aws_cloudhsm_v2_hsm" "test" {
  availability_zone = aws_subnet.test[0].availability_zone
  cluster_id        = aws_cloudhsm_v2_cluster.test.cluster_id
}
`)
}

func testAccHSMConfig_ipAddress(rName string) string {
	return acctest.ConfigCompose(testAccHSMConfig_base(rName), `
resource "aws_cloudhsm_v2_hsm" "test" {
  cluster_id = aws_cloudhsm_v2_cluster.test.cluster_id
  ip_address = cidrhost(aws_subnet.test[0].cidr_block, 5)
  subnet_id  = aws_subnet.test[0].id
}
`)
}

func testAccHSMConfig_subnetID(rName string) string {
	return acctest.ConfigCompose(testAccHSMConfig_base(rName), `
resource "aws_cloudhsm_v2_hsm" "test" {
  cluster_id = aws_cloudhsm_v2_cluster.test.cluster_id
  subnet_id  = aws_subnet.test[0].id
}
`)
}
