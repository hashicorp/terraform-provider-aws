// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftDataShareConsumerAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_data_share_consumer_association.test"
	regionDataSourceName := "data.aws_region.current"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, redshift.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataShareConsumerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataShareConsumerAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataShareConsumerAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "consumer_region", regionDataSourceName, names.AttrName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "data_share_arn", "redshift", regexache.MustCompile(`datashare:+.`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "producer_arn", "redshift-serverless", regexache.MustCompile(`namespace/+.`)),
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

func TestAccRedshiftDataShareConsumerAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_data_share_consumer_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, redshift.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataShareConsumerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataShareConsumerAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataShareConsumerAssociationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceDataShareConsumerAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftDataShareConsumerAssociation_associateEntireAccount(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_data_share_consumer_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, redshift.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataShareConsumerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataShareConsumerAssociationConfig_associateEntireAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataShareConsumerAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "associate_entire_account", acctest.CtTrue),
					acctest.MatchResourceAttrRegionalARN(resourceName, "data_share_arn", "redshift", regexache.MustCompile(`datashare:+.`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "producer_arn", "redshift-serverless", regexache.MustCompile(`namespace/+.`)),
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

func testAccCheckDataShareConsumerAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_data_share_consumer_association" {
				continue
			}

			_, err := tfredshift.FindDataShareConsumerAssociationByID(ctx, conn, rs.Primary.ID)
			if tfawserr.ErrMessageContains(err, redshift.ErrCodeInvalidDataShareFault, "because the ARN doesn't exist.") ||
				tfawserr.ErrMessageContains(err, redshift.ErrCodeInvalidDataShareFault, "either doesn't exist or isn't associated with this data consumer") {
				return nil
			}
			if err != nil {
				return create.Error(names.Redshift, create.ErrActionCheckingDestroyed, tfredshift.ResNameDataShareConsumerAssociation, rs.Primary.ID, err)
			}

			return create.Error(names.Redshift, create.ErrActionCheckingDestroyed, tfredshift.ResNameDataShareConsumerAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDataShareConsumerAssociationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Redshift, create.ErrActionCheckingExistence, tfredshift.ResNameDataShareConsumerAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Redshift, create.ErrActionCheckingExistence, tfredshift.ResNameDataShareConsumerAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)
		_, err := tfredshift.FindDataShareConsumerAssociationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Redshift, create.ErrActionCheckingExistence, tfredshift.ResNameDataShareConsumerAssociation, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccDataShareConsumerAssociationConfigBase(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
  db_name        = "test"
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_redshiftdata_statement" "test_create" {
  workgroup_name = aws_redshiftserverless_workgroup.test.workgroup_name
  database       = aws_redshiftserverless_namespace.test.db_name
  sql            = "CREATE DATASHARE tfacctest;"
}
`, rName),
		// Split this resource into a string literal so the terraform `format` function
		// interpolates properly
		`
resource "aws_redshiftdata_statement" "test_grant_usage" {
  depends_on     = [aws_redshiftdata_statement.test_create]
  workgroup_name = aws_redshiftserverless_workgroup.test.workgroup_name
  database       = aws_redshiftserverless_namespace.test.db_name
  sql            = format("GRANT USAGE ON DATASHARE tfacctest TO ACCOUNT '%s';", data.aws_caller_identity.current.account_id)
}

locals {
  # Data share ARN is not returned from the GRANT USAGE statement, so must be
  # composed manually.
  # Ref: https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonredshift.html#amazonredshift-resources-for-iam-policies
  data_share_arn = format("arn:%s:redshift:%s:%s:datashare:%s/%s",
    data.aws_partition.current.id,
    data.aws_region.current.name,
    data.aws_caller_identity.current.account_id,
    aws_redshiftserverless_namespace.test.namespace_id,
    "tfacctest",
  )
}

resource "aws_redshift_data_share_authorization" "test" {
  depends_on = [aws_redshiftdata_statement.test_grant_usage]

  data_share_arn      = local.data_share_arn
  consumer_identifier = data.aws_caller_identity.current.account_id
}
`)
}

func testAccDataShareConsumerAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDataShareConsumerAssociationConfigBase(rName),
		`
resource "aws_redshift_data_share_consumer_association" "test" {
  depends_on = [aws_redshift_data_share_authorization.test]

  data_share_arn  = local.data_share_arn
  consumer_region = data.aws_region.current.name
}
`)
}

func testAccDataShareConsumerAssociationConfig_associateEntireAccount(rName string) string {
	return acctest.ConfigCompose(
		testAccDataShareConsumerAssociationConfigBase(rName),
		`
resource "aws_redshift_data_share_consumer_association" "test" {
  depends_on = [aws_redshift_data_share_authorization.test]

  data_share_arn           = local.data_share_arn
  associate_entire_account = true
}
`)
}
