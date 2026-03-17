// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	"github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAppSyncSourceAPIAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var sourceapiassociation types.SourceApiAssociation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_source_api_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSourceAPIAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSourceAPIAssociationConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceAPIAssociationExists(ctx, t, resourceName, &sourceapiassociation),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+/sourceApiAssociations/.+`)),
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

func testAccAppSyncSourceAPIAssociation_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var sourceapiassociation types.SourceApiAssociation
	var sourceapiassociationUpdated types.SourceApiAssociation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_source_api_association.test"
	updateDesc := rName + "Update"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSourceAPIAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSourceAPIAssociationConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceAPIAssociationExists(ctx, t, resourceName, &sourceapiassociation),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+/sourceApiAssociations/.+`)),
				),
			},
			{
				Config: testAccSourceAPIAssociationConfig_basic(rName, updateDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceAPIAssociationExists(ctx, t, resourceName, &sourceapiassociationUpdated),
					testAccCheckSourceAPIAssociationNotRecreated(&sourceapiassociation, &sourceapiassociationUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updateDesc),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+/sourceApiAssociations/.+`)),
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

func testAccAppSyncSourceAPIAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var sourceapiassociation types.SourceApiAssociation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_source_api_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSourceAPIAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSourceAPIAssociationConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceAPIAssociationExists(ctx, t, resourceName, &sourceapiassociation),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfappsync.ResourceSourceAPIAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSourceAPIAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_source_api_association" {
				continue
			}

			_, err := tfappsync.FindSourceAPIAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAssociationID], rs.Primary.Attributes["merged_api_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Appsync Source API Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSourceAPIAssociationExists(ctx context.Context, t *testing.T, n string, v *types.SourceApiAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

		output, err := tfappsync.FindSourceAPIAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAssociationID], rs.Primary.Attributes["merged_api_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

	input := &appsync.ListGraphqlApisInput{}
	_, err := conn.ListGraphqlApis(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckSourceAPIAssociationNotRecreated(before, after *types.SourceApiAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.AssociationId), aws.ToString(after.AssociationId); before != after {
			return errors.New("recreated")
		}

		return nil
	}
}

func testAccSourceAPIAssociationConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = data.aws_iam_policy_document.test.json
  name_prefix        = %[1]q
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["appsync.amazonaws.com"]
      type        = "Service"
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }

    condition {
      test     = "ArnLike"
      values   = ["arn:${data.aws_partition.current.partition}:appsync:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}::apis/*"]
      variable = "aws:SourceArn"
    }
  }
}

resource "aws_appsync_graphql_api" "merged" {
  authentication_type           = "API_KEY"
  name                          = %[1]q
  api_type                      = "MERGED"
  merged_api_execution_role_arn = aws_iam_role.test.arn
}

resource "aws_appsync_graphql_api" "source" {
  authentication_type = "API_KEY"
  name                = %[1]q
  schema              = <<EOF
schema {
    query: Query
}
type Query {
  test: Int
}
EOF
}

resource "aws_appsync_source_api_association" "test" {
  description   = %[2]q
  merged_api_id = aws_appsync_graphql_api.merged.id
  source_api_id = aws_appsync_graphql_api.source.id
}
`, rName, description)
}
