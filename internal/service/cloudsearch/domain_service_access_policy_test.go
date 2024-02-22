// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudsearch_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudsearch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudsearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudSearchDomainServiceAccessPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudsearch_domain_service_access_policy.test"
	rName := acctest.ResourcePrefix + "-" + sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudSearchEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainServiceAccessPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainServiceAccessPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainServiceAccessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_policy"),
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

func TestAccCloudSearchDomainServiceAccessPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudsearch_domain_service_access_policy.test"
	rName := acctest.ResourcePrefix + "-" + sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudSearchEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainServiceAccessPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainServiceAccessPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainServiceAccessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainServiceAccessPolicyConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDomainServiceAccessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_policy"),
				),
			},
		},
	})
}

func testAccDomainServiceAccessPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudSearchClient(ctx)

		_, err := tfcloudsearch.FindAccessPolicyByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckDomainServiceAccessPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudsearch_domain_service_access_policy" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).CloudSearchClient(ctx)

			_, err := tfcloudsearch.FindAccessPolicyByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudSearch Domain Service Access Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDomainServiceAccessPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q
}

resource "aws_cloudsearch_domain_service_access_policy" "test" {
  domain_name = aws_cloudsearch_domain.test.id

  access_policy = <<POLICY
{
  "Version":"2012-10-17",
  "Statement":[{
    "Sid":"search_and_document",
    "Effect":"Allow",
    "Principal":"*",
    "Action":[
      "cloudsearch:search",
      "cloudsearch:document"
    ],
    "Condition":{"IpAddress":{"aws:SourceIp":"192.0.2.0/32"}}
  }]
}
POLICY
}
`, rName)
}

func testAccDomainServiceAccessPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
  name = %[1]q
}

resource "aws_cloudsearch_domain_service_access_policy" "test" {
  domain_name = aws_cloudsearch_domain.test.id

  access_policy = <<POLICY
{
  "Version":"2012-10-17",
  "Statement":[{
    "Sid":"all",
    "Effect":"Allow",
    "Action":"cloudsearch:*",
    "Principal":"*"
  }]
}
POLICY
}
`, rName)
}
