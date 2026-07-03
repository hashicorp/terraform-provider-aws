// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53profiles_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ProfilesProfileDataSource_byName(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53profiles_profile.test"
	resourceName := "aws_route53profiles_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileDataSourceConfig_byName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "profile_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(dataSourceName, "share_status"),
				),
			},
		},
	})
}

func TestAccRoute53ProfilesProfileDataSource_byID(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53profiles_profile.test"
	resourceName := "aws_route53profiles_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileDataSourceConfig_byID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "profile_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(dataSourceName, "share_status"),
				),
			},
		},
	})
}

func testAccProfileDataSourceConfig_byName(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

data "aws_route53profiles_profile" "test" {
  name = aws_route53profiles_profile.test.name
}
`, rName)
}

func testAccProfileDataSourceConfig_byID(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

data "aws_route53profiles_profile" "test" {
  profile_id = aws_route53profiles_profile.test.id
}
`, rName)
}
