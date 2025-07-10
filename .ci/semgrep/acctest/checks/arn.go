package main

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func test1(t *testing.T) {
	ctx := acctest.Context(t)

	const resourceName = "aws_prometheus_scraper.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScraperDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScraperConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// ruleid: arn-resourceattrset
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					// ok: arn-resourceattrset
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "aps", regexache.MustCompile(`scraper/\w+$`)),
					// todoruleid: arn-resourceattrset
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					// todoruleid: arn-resourceattrset
					resource.TestCheckResourceAttrSet(resourceName, "some_other_arn"),
				),
			},
		},
	})

}
