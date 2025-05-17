// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMServiceLinkedRoleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_service_linked_role.test"
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "airflow.amazonaws.com"
	name := "AWSServiceRoleForAmazonMWAA"
	path := fmt.Sprintf("/aws-service-role/%s/", awsServiceName)
	arnResource := fmt.Sprintf("role%s%s", path, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{

			{
				PreConfig: func() {
					// Remove existing if possible
					client := acctest.Provider.Meta().(*conns.AWSClient)
					arn := arn.ARN{
						Partition: client.Partition(ctx),
						Service:   "iam",
						Region:    client.Region(ctx),
						AccountID: client.AccountID(ctx),
						Resource:  arnResource,
					}.String()
					r := tfiam.ResourceServiceLinkedRole()
					d := r.Data(nil)
					d.SetId(arn)
					err := acctest.DeleteResource(ctx, r, d, client)

					if err != nil {
						t.Fatalf("deleting service-linked role %s: %s", name, err)
					}
				},
				Config: testAccServiceLinkedRoleDataSourceConfig_basic(awsServiceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "aws_service_name", resourceName, "aws_service_name"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "create_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "unique_id", resourceName, "unique_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPath, resourceName, names.AttrPath),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccIAMServiceLinkedRoleDataSource_customSuffix(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_service_linked_role.test"
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "autoscaling.amazonaws.com"
	rCustomSufix := sdkacctest.RandString(10)
	rDescription := "This is a service linked role"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{

			{
				Config: testAccServiceLinkedRoleDataSourceConfig_customSuffix(awsServiceName, rCustomSufix, rDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "aws_service_name", resourceName, "aws_service_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "custom_suffix", resourceName, "custom_suffix"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "create_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "unique_id", resourceName, "unique_id"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccIAMServiceLinkedRoleDataSource_createIfMissing(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_service_linked_role.test"
	awsServiceName := "autoscaling.amazonaws.com"
	name := "AWSServiceRoleForAutoScaling"
	customSuffix := "ServiceLinkedRoleDataSource"
	path := fmt.Sprintf("/aws-service-role/%s/", awsServiceName)
	arnResource := fmt.Sprintf("role%s%s_%s", path, name, customSuffix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{

			{
				PreConfig: func() {
					// Remove existing if possible
					client := acctest.Provider.Meta().(*conns.AWSClient)
					arn := arn.ARN{
						Partition: client.Partition(ctx),
						Service:   "iam",
						Region:    client.Region(ctx),
						AccountID: client.AccountID(ctx),
						Resource:  arnResource,
					}.String()
					r := tfiam.ResourceServiceLinkedRole()
					d := r.Data(nil)
					d.SetId(arn)
					err := acctest.DeleteResource(ctx, r, d, client)

					if err != nil {
						t.Fatalf("deleting service-linked role %s: %s", name, err)
					}
				},
				Config:      testAccServiceLinkedRoleDataSourceConfig_createIfMissing(awsServiceName, customSuffix, false),
				ExpectError: regexache.MustCompile("Role was not found"),
			},
			{
				Config: testAccServiceLinkedRoleDataSourceConfig_createIfMissing(awsServiceName, customSuffix, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "aws_service_name", awsServiceName),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "create_date"),
					resource.TestCheckResourceAttrSet(dataSourceName, "unique_id"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func testAccServiceLinkedRoleDataSourceConfig_basic(awsServiceName string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = %[1]q
  description      = "This is a service linked role"
}

data "aws_iam_service_linked_role" "test" {
  aws_service_name = aws_iam_service_linked_role.test.aws_service_name
}
`, awsServiceName)
}

func testAccServiceLinkedRoleDataSourceConfig_customSuffix(awsServiceName string, customSuffix string, description string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = %[1]q
  custom_suffix    = %[2]q
  description      = %[3]q
}

data "aws_iam_service_linked_role" "test" {
  aws_service_name = aws_iam_service_linked_role.test.aws_service_name
  custom_suffix    = aws_iam_service_linked_role.test.custom_suffix
}
`, awsServiceName, customSuffix, description)
}

func testAccServiceLinkedRoleDataSourceConfig_createIfMissing(awsServiceName string, customSufix string, createIfMissing bool) string {
	return fmt.Sprintf(`

data "aws_iam_service_linked_role" "test" {
  aws_service_name  = %[1]q
  create_if_missing = %[2]t
  custom_suffix     = %[3]q
}
`, awsServiceName, createIfMissing, customSufix)
}
