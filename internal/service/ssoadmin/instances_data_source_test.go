// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccPreCheckInstances(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn(ctx)

	var instances []*ssoadmin.InstanceMetadata
	err := conn.ListInstancesPagesWithContext(ctx, &ssoadmin.ListInstancesInput{}, func(page *ssoadmin.ListInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		instances = append(instances, page.Instances...)

		return !lastPage
	})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if len(instances) == 0 {
		t.Skip("skipping acceptance testing: No SSO Instance found.")
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func TestAccSSOAdminInstancesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ssoadmin_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "identity_store_ids.#", "1"),
					acctest.MatchResourceAttrGlobalARNNoAccount(dataSourceName, "arns.0", "sso", regexache.MustCompile("instance/(sso)?ins-[0-9A-Za-z.-]{16}")),
					resource.TestMatchResourceAttr(dataSourceName, "identity_store_ids.0", regexache.MustCompile("^[0-9A-Za-z-]*")),
				),
			},
		},
	})
}

const testAccInstancesDataSourceConfig_basic = `data "aws_ssoadmin_instances" "test" {}`
