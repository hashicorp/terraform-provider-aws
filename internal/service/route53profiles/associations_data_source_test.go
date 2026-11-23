// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53profiles_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ProfilesAssociationsDataSource_resourceID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53profiles_associations.test"
	resourceName1 := "aws_route53profiles_association.test1"
	resourceName2 := "aws_route53profiles_association.test2"
	bothFiltersName := "data.aws_route53profiles_associations.both"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53Profiles)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationsDataSourceConfig_resourceID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "associations.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "associations.*.id", resourceName1, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "associations.*.id", resourceName2, names.AttrID),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "associations.*", map[string]string{
						names.AttrName:     rName + "-1",
						names.AttrPriority: "100",
						names.AttrStatus:   string(awstypes.ProfileStatusComplete),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "associations.*", map[string]string{
						names.AttrName:     rName + "-2",
						names.AttrPriority: "200",
						names.AttrStatus:   string(awstypes.ProfileStatusComplete),
					}),
					resource.TestCheckResourceAttr(bothFiltersName, "associations.#", "1"),
					resource.TestCheckResourceAttrPair(bothFiltersName, "associations.0.id", resourceName2, names.AttrID),
				),
			},
		},
	})
}

func TestAccRoute53ProfilesAssociationsDataSource_profileID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53profiles_associations.test"
	resourceName := "aws_route53profiles_association.test"
	noMatchName := "data.aws_route53profiles_associations.no_match"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53Profiles)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationsDataSourceConfig_profileID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "associations.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.0.id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.0.name", resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.0.owner_id", resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.0.priority", resourceName, names.AttrPriority),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.0.profile_id", resourceName, "profile_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.0.resource_id", resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.0.status", resourceName, names.AttrStatus),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(noMatchName, tfjsonpath.New("associations"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccAssociationsDataSourceConfig_resourceID(rName string) string {
	return acctest.ConfigCompose(testAccAssociationConfig_multipleProfiles(rName, 100, 200), `
data "aws_route53profiles_associations" "test" {
  resource_id = aws_vpc.test.id

  depends_on = [
    aws_route53profiles_association.test1,
    aws_route53profiles_association.test2,
  ]
}

data "aws_route53profiles_associations" "both" {
  profile_id  = aws_route53profiles_association.test2.profile_id
  resource_id = aws_route53profiles_association.test2.resource_id
}
`)
}

func testAccAssociationsDataSourceConfig_profileID(rName string) string {
	return acctest.ConfigCompose(testAccAssociationConfig_basic(rName), `
data "aws_route53profiles_associations" "test" {
  profile_id = aws_route53profiles_association.test.profile_id
}

data "aws_route53profiles_associations" "no_match" {
  profile_id  = aws_route53profiles_association.test.profile_id
  resource_id = "vpc-00000000000000000"
}
`)
}
