// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mpa_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mpa"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmpa "github.com/hashicorp/terraform-provider-aws/internal/service/mpa"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMPAIdentitySource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var identitysource mpa.GetIdentitySourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mpa_identity_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MPA)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MPA),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentitySourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentitySourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, t, resourceName, &identitysource),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mpa", regexache.MustCompile(`identity-source/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
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

func TestAccMPAIdentitySource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var identitysource mpa.GetIdentitySourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mpa_identity_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MPA)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MPA),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentitySourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentitySourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, t, resourceName, &identitysource),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfmpa.ResourceIdentitySource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIdentitySourceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MPAClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mpa_identity_source" {
				continue
			}

			_, err := tfmpa.FindIdentitySourceByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("MPA Identity Source %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIdentitySourceExists(ctx context.Context, t *testing.T, name string, identitysource *mpa.GetIdentitySourceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).MPAClient(ctx)

		resp, err := tfmpa.FindIdentitySourceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*identitysource = *resp

		return nil
	}
}

func testAccIdentitySourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}
data "aws_region" "current" {}

resource "aws_mpa_identity_source" "test" {
  name = %[1]q

  identity_source_parameters {
    iam_identity_center {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      region       = data.aws_region.current.name
    }
  }
}
`, rName)
}
