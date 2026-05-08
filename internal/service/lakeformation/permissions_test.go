// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPermissions_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "data_cells_filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "data_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "database.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lf_tag.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionCreateDatabase)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "0"),
				),
			},
		},
	})
}

func testAccPermissions_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcBasic(rName, "\"event\", \"timestamp\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflakeformation.ResourcePermissions(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPermissions_database(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	dbName := "aws_glue_catalog_database.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_database(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionAlter)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionCreateTable)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDrop)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", string(awstypes.PermissionCreateTable)),
				),
			},
		},
	})
}

func testAccPermissions_databaseIAMAllowed(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dbName := "aws_glue_catalog_database.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_databaseIAMAllowed(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, tflakeformation.IAMAllowedPrincipals),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionAll)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccPermissions_databaseIAMPrincipals(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dbName := "aws_glue_catalog_database.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_databaseIAMPrincipals(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					testAccCheckIAMPrincipalsGrantPrincipal(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionAll)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccPermissions_databaseMultiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	resourceName2 := "aws_lakeformation_permissions.test2"
	roleName := "aws_iam_role.test"
	roleName2 := "aws_iam_role.test2"
	dbName := "aws_glue_catalog_database.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_databaseMultiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionAlter)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionCreateTable)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDrop)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", string(awstypes.PermissionCreateTable)),
					testAccCheckPermissionsExists(ctx, t, resourceName2),
					resource.TestCheckResourceAttrPair(resourceName2, names.AttrPrincipal, roleName2, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName2, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName2, names.AttrPrincipal, roleName2, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName2, "database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName2, "database.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName2, "permissions.#", "2"),
					resource.TestCheckResourceAttr(resourceName2, "permissions.0", string(awstypes.PermissionAlter)),
					resource.TestCheckResourceAttr(resourceName2, "permissions.1", string(awstypes.PermissionDrop)),
					resource.TestCheckResourceAttr(resourceName2, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccPermissions_dataCellsFilter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_dataCellsFilter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDescribe)),
					resource.TestCheckResourceAttr(resourceName, "data_cells_filter.#", "1"),
				),
			},
		},
	})
}

func testAccPermissions_dataLocation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	bucketName := "aws_s3_bucket.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_dataLocation(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDataLocationAccess)),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "data_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "data_location.0.arn", bucketName, names.AttrARN),
				),
			},
		},
	})
}

func testAccPermissions_lfTag(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tagName := "aws_lakeformation_lf_tag.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_lfTag(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lf_tag.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag.0.key", tagName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag.0.values", tagName, names.AttrValues),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", "ASSOCIATE"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", "DESCRIBE"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", "ASSOCIATE"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.1", "DESCRIBE"),
				),
			},
		},
	})
}

func testAccPermissions_lfTagPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tagName := "aws_lakeformation_lf_tag.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_lfTagPolicy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.0.resource_type", "DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.0.expression.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag_policy.0.expression.0.key", tagName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag_policy.0.expression.0.values", tagName, names.AttrValues),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionAlter)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionCreateTable)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDrop)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", string(awstypes.PermissionCreateTable)),
				),
			},
		},
	})
}

func testAccPermissions_lfTagPolicyMultiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	// tagName := "aws_lakeformation_lf_tag.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_lfTagPolicyMultiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.0.resource_type", "DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.0.expression.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag_policy.0.expression.*", map[string]string{
						names.AttrKey: rName + "-0",
						"values.#":    "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag_policy.0.expression.*", map[string]string{
						names.AttrKey: rName + "-1",
						"values.#":    "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag_policy.0.expression.*", map[string]string{
						names.AttrKey: rName + "-2",
						"values.#":    "2",
					}),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionAlter)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionCreateTable)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDrop)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", string(awstypes.PermissionCreateTable)),
				),
			},
		},
	})
}

func testAccPermissions_tableBasic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(roleName, names.AttrARN, resourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionAlter)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDelete)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDescribe)),
				),
			},
		},
	})
}

func testAccPermissions_tableIAMAllowed(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dbName := "aws_glue_catalog_table.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableIAMAllowed(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, tflakeformation.IAMAllowedPrincipals),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", dbName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionAll)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccPermissions_tableIAMPrincipals(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dbName := "aws_glue_catalog_table.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableIAMPrincipals(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					testAccCheckIAMPrincipalsGrantPrincipal(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", dbName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionAll)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccPermissions_tableImplicit(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableImplicit(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "7"),
				),
			},
		},
	})
}

func testAccPermissions_tableMultipleRoles(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	resourceName2 := "aws_lakeformation_permissions.test2"
	roleName := "aws_iam_role.test"
	roleName2 := "aws_iam_role.test2"
	tableName := "aws_glue_catalog_table.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableMultipleRoles(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(roleName, names.AttrARN, resourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionAlter)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDelete)),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDescribe)),
					testAccCheckPermissionsExists(ctx, t, resourceName2),
					resource.TestCheckResourceAttrPair(roleName2, names.AttrARN, resourceName2, names.AttrPrincipal),
					resource.TestCheckResourceAttr(resourceName2, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName2, "table.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName2, "table.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName2, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName2, "permissions.0", string(awstypes.PermissionSelect)),
				),
			},
		},
	})
}

func testAccPermissions_tableSelectOnly(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableSelectOnly(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(roleName, names.AttrARN, resourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionSelect)),
				),
			},
		},
	})
}

func testAccPermissions_tableSelectPlus(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableSelectPlus(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "7"),
				),
			},
		},
	})
}

func testAccPermissions_tableWildcardNoSelect(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	databaseResourceName := "aws_glue_catalog_database.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableWildcardNoSelect(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", databaseResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table.0.wildcard", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccPermissions_tableWildcardSelectOnly(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableWildcardSelectOnly(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionSelect)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccPermissions_tableWildcardSelectPlus(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableWildcardSelectPlus(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "7"),
				),
			},
		},
	})
}

func testAccPermissions_table_nonIAMPrincipals(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dbName := "aws_glue_catalog_table.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_table_nonIAMPrincipals(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, "aws_identitystore_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", dbName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDescribe)),
				),
			},
		},
	})
}

func testAccPermissions_twcBasic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcBasic(rName, "\"event\", \"timestamp\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionSelect)),
				),
			},
			{
				Config: testAccPermissionsConfig_twcBasic(rName, "\"timestamp\", \"event\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionSelect)),
				),
			},
			{
				Config: testAccPermissionsConfig_twcBasic(rName, "\"timestamp\", \"event\", \"transactionamount\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.2", "transactionamount"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionSelect)),
				),
			},
			{
				Config: testAccPermissionsConfig_twcBasic(rName, "\"event\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionSelect)),
				),
			},
		},
	})
}

func testAccPermissions_twcImplicit(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcImplicit(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.wildcard", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "7"),
				),
			},
		},
	})
}

func testAccPermissions_twcWildcardExcludedColumns(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcWildcardExcludedColumns(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccPermissions_twcWildcardSelectOnly(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcWildcardSelectOnly(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.wildcard", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionSelect)),
				),
			},
		},
	})
}

func testAccPermissions_twcWildcardSelectPlus(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcWildcardSelectPlus(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionSelect)),
				),
			},
		},
	})
}

func testAccPermissions_catalogResource_nonIAMPrincipals(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_catalogResource_nonIAMPrincipals(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, "aws_identitystore_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*", string(awstypes.PermissionDescribe)),
				),
			},
		},
	})
}

func testAccCheckIAMPrincipalsGrantPrincipal(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("acceptance test: resource not found: %s", resourceName)
		}

		if v, ok := rs.Primary.Attributes[names.AttrPrincipal]; ok && v != "" {
			expectedPrincipalValue := acctest.AccountID(ctx) + ":IAMPrincipals"
			if v == expectedPrincipalValue {
				return nil
			} else {
				return fmt.Errorf("acceptance test: unexpected principal value for (%s). Is %s, should be %s", rs.Primary.ID, v, expectedPrincipalValue)
			}
		}

		return fmt.Errorf("acceptance test: error finding IAMPrincipals grant (%s)", rs.Primary.ID)
	}
}

func testAccCheckPermissionsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_permissions" {
				continue
			}

			permCount, err := permissionCountForResource(ctx, t, conn, rs)

			if err != nil {
				return fmt.Errorf("acceptance test: error listing Lake Formation permissions (%s): %w", rs.Primary.ID, err)
			}

			if permCount != 0 {
				return fmt.Errorf("acceptance test: Lake Formation permissions (%s) still exist: %d", rs.Primary.ID, permCount)
			}

			return nil
		}

		return nil
	}
}

func testAccCheckPermissionsExists(ctx context.Context, t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("acceptance test: resource not found: %s", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)

		permCount, err := permissionCountForResource(ctx, t, conn, rs)

		if err != nil {
			return fmt.Errorf("acceptance test: error listing Lake Formation permissions (%s): %w", rs.Primary.ID, err)
		}

		if permCount == 0 {
			return fmt.Errorf("acceptance test: Lake Formation permissions (%s) do not exist or could not be found", rs.Primary.ID)
		}

		return nil
	}
}

func permissionCountForResource(ctx context.Context, t *testing.T, conn *lakeformation.Client, rs *terraform.ResourceState) (int, error) {
	input := &lakeformation.ListPermissionsInput{
		Resource: &awstypes.Resource{},
	}

	principalIdentifier := rs.Primary.Attributes[names.AttrPrincipal]
	if tflakeformation.IncludePrincipalIdentifierInList(principalIdentifier) {
		principal := awstypes.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(principalIdentifier),
		}
		input.Principal = &principal
	}

	noResource := true

	if v, ok := rs.Primary.Attributes[names.AttrCatalogID]; ok && v != "" {
		input.CatalogId = aws.String(v)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["catalog_resource"]; ok && v == acctest.CtTrue {
		input.Resource.Catalog = tflakeformation.ExpandCatalogResource()

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["data_cells_filter.#"]; ok && v != "" && v != "0" {
		tfMap := map[string]any{}

		if v := rs.Primary.Attributes["data_cells_filter.0.database_name"]; v != "" {
			tfMap[names.AttrDatabaseName] = v
		}

		if v := rs.Primary.Attributes["data_cells_filter.0.name"]; v != "" {
			tfMap[names.AttrName] = v
		}

		if v := rs.Primary.Attributes["data_cells_filter.0.table_catalog_id"]; v != "" {
			tfMap["table_catalog_id"] = v
		}

		if v := rs.Primary.Attributes["data_cells_filter.0.table_name"]; v != "" {
			tfMap[names.AttrTableName] = v
		}

		input.Resource.DataCellsFilter = tflakeformation.ExpandDataCellsFilter([]any{tfMap})
		noResource = false
	}

	if v, ok := rs.Primary.Attributes["data_location.#"]; ok && v != "" && v != "0" {
		tfMap := map[string]any{}

		if v := rs.Primary.Attributes["data_location.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := rs.Primary.Attributes["data_location.0.arn"]; v != "" {
			tfMap[names.AttrARN] = v
		}

		input.Resource.DataLocation = tflakeformation.ExpandDataLocationResource(tfMap)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["database.#"]; ok && v != "" && v != "0" {
		tfMap := map[string]any{}

		if v := rs.Primary.Attributes["database.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := rs.Primary.Attributes["database.0.name"]; v != "" {
			tfMap[names.AttrName] = v
		}

		input.Resource.Database = tflakeformation.ExpandDatabaseResource(tfMap)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["lf_tag.#"]; ok && v != "" && v != "0" {
		tfMap := map[string]any{}

		if v := rs.Primary.Attributes["lf_tag.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := rs.Primary.Attributes["lf_tag.0.key"]; v != "" {
			tfMap[names.AttrKey] = v
		}

		if count, err := strconv.Atoi(rs.Primary.Attributes["lf_tag.0.values.#"]); err == nil {
			var tagValues []string
			for i := range count {
				tagValues = append(tagValues, rs.Primary.Attributes[fmt.Sprintf("lf_tag.0.values.%d", i)])
			}
			tfMap[names.AttrValues] = flex.FlattenStringSet(aws.StringSlice(tagValues))
		}

		input.Resource.LFTag = tflakeformation.ExpandLFTagKeyResource(tfMap)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["lf_tag_policy.#"]; ok && v != "" && v != "0" {
		tfMap := map[string]any{}

		if v := rs.Primary.Attributes["lf_tag_policy.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := rs.Primary.Attributes["lf_tag_policy.0.resource_type"]; v != "" {
			tfMap[names.AttrResourceType] = v
		}

		if expressionCount, err := strconv.Atoi(rs.Primary.Attributes["lf_tag_policy.0.expression.#"]); err == nil {
			expressionSlice := make([]any, expressionCount)
			for i := range expressionCount {
				expression := make(map[string]any)

				if v := rs.Primary.Attributes[fmt.Sprintf("lf_tag_policy.0.expression.%d.key", i)]; v != "" {
					expression[names.AttrKey] = v
				}

				if expressionValueCount, err := strconv.Atoi(rs.Primary.Attributes[fmt.Sprintf("lf_tag_policy.0.expression.%d.values.#", i)]); err == nil {
					valueSlice := make([]string, expressionValueCount)
					for j := range expressionValueCount {
						valueSlice[j] = rs.Primary.Attributes[fmt.Sprintf("lf_tag_policy.0.expression.%d.values.%d", i, j)]
					}
					expression[names.AttrValues] = flex.FlattenStringSet(aws.StringSlice(valueSlice))
				}
				expressionSlice[i] = expression
			}
			// The exact details of the set hash function don't matter, elements just have distinct values.
			tfMap[names.AttrExpression] = schema.NewSet(func(_ any) int { return acctest.RandInt(t) }, expressionSlice)
		}

		input.Resource.LFTagPolicy = tflakeformation.ExpandLFTagPolicyResource(tfMap)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["table.#"]; ok && v != "" && v != "0" {
		tfMap := map[string]any{}

		if v := rs.Primary.Attributes["table.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := rs.Primary.Attributes["table.0.database_name"]; v != "" {
			tfMap[names.AttrDatabaseName] = v
		}

		if v := rs.Primary.Attributes["table.0.name"]; v != "" && v != tflakeformation.TableNameAllTables {
			tfMap[names.AttrName] = v
		}

		if v := rs.Primary.Attributes["table.0.wildcard"]; v != "" && v == acctest.CtTrue {
			tfMap["wildcard"] = true
		}

		input.Resource.Table = tflakeformation.ExpandTableResource(tfMap)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["table_with_columns.#"]; ok && v != "" && v != "0" {
		tfMap := map[string]any{}

		if v := rs.Primary.Attributes["table_with_columns.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := rs.Primary.Attributes["table_with_columns.0.database_name"]; v != "" {
			tfMap[names.AttrDatabaseName] = v
		}

		if v := rs.Primary.Attributes["table_with_columns.0.name"]; v != "" {
			tfMap[names.AttrName] = v
		}

		input.Resource.Table = tflakeformation.ExpandTableWithColumnsResourceAsTable(tfMap)

		noResource = false
	}

	if noResource {
		// if after read, there is no resource, it has been deleted
		return 0, nil
	}

	filter, err := permissionsFilter(rs.Primary.Attributes)
	if err != nil {
		return 0, fmt.Errorf("acceptance test: error creating permissions filter for (%s): %w", rs.Primary.ID, err)
	}

	var permissions []awstypes.PrincipalResourcePermissions

	err = tfresource.Retry(ctx, tflakeformation.IAMPropagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		pages := lakeformation.NewListPermissionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Resource does not exist") {
				return nil
			}

			if errs.IsA[*awstypes.EntityNotFoundException](err) {
				return nil
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Invalid principal") {
				return tfresource.RetryableError(err)
			}

			if err != nil {
				return tfresource.NonRetryableError(fmt.Errorf("acceptance test: error listing Lake Formation Permissions getting permission count: %w", err))
			}

			for _, permission := range page.PrincipalResourcePermissions {
				if filter(permission) {
					permissions = append(permissions, permission)
				}
			}
		}

		return nil
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return 0, nil
	}

	// used to find error in retryable error
	if tfawserr.ErrMessageContains(err, "AccessDeniedException", "Resource does not exist") {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("acceptance test: error listing Lake Formation permissions: %w", err)
	}

	return len(permissions), nil
}

func permissionsFilter(attributes map[string]string) (tflakeformation.PermissionsFilter, error) {
	principalIdentifier := attributes[names.AttrPrincipal]

	if v, ok := attributes["catalog_resource"]; ok && v == acctest.CtTrue {
		return tflakeformation.FilterCatalogPermissions(principalIdentifier), nil
	}
	if v, ok := attributes["data_cells_filter.#"]; ok && v != "" && v != "0" {
		return tflakeformation.FilterDataCellsFilter(principalIdentifier), nil
	}
	if v, ok := attributes["data_location.#"]; ok && v != "" && v != "0" {
		return tflakeformation.FilterDataLocationPermissions(principalIdentifier), nil
	}
	if v, ok := attributes["database.#"]; ok && v != "" && v != "0" {
		return tflakeformation.FilterDatabasePermissions(principalIdentifier), nil
	}
	if v, ok := attributes["lf_tag.#"]; ok && v != "" && v != "0" {
		return tflakeformation.FilterLFTagPermissions(principalIdentifier), nil
	}
	if v, ok := attributes["lf_tag_policy.#"]; ok && v != "" && v != "0" {
		return tflakeformation.FilterLFTagPolicyPermissions(principalIdentifier), nil
	}
	if v, ok := attributes["table.#"]; ok && v != "" && v != "0" {
		tfMap := map[string]any{}

		if v := attributes["table.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := attributes["table.0.database_name"]; v != "" {
			tfMap[names.AttrDatabaseName] = v
		}

		if v := attributes["table.0.name"]; v != "" && v != tflakeformation.TableNameAllTables {
			tfMap[names.AttrName] = v
		}

		if v := attributes["table.0.wildcard"]; v != "" && v == acctest.CtTrue {
			tfMap["wildcard"] = true
		}
		return tflakeformation.FilterTablePermissions(principalIdentifier, tflakeformation.ExpandTableResource(tfMap)), nil
	}
	if v, ok := attributes["table_with_columns.#"]; ok && v != "" && v != "0" {
		tfMap := map[string]any{}

		if v := attributes["table_with_columns.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := attributes["table_with_columns.0.database_name"]; v != "" {
			tfMap[names.AttrDatabaseName] = v
		}

		if v := attributes["table_with_columns.0.name"]; v != "" {
			tfMap[names.AttrName] = v
		}

		var columnNames []string
		if v := attributes["table_with_columns.0.column_names.#"]; v != "" {
			colCount, err := strconv.Atoi(attributes["table_with_columns.0.column_names.#"])
			if err != nil {
				return nil, fmt.Errorf("acceptance test: could not convert string (%s) Atoi for column_names: %w", attributes["table_with_columns.0.column_names.#"], err)
			}
			for i := range colCount {
				columnNames = append(columnNames, attributes[fmt.Sprintf("table_with_columns.0.column_names.%d", i)])
			}
		}

		var excludedColumnNames []string
		if v := attributes["table_with_columns.0.excluded_column_names.#"]; v != "" {
			colCount, err := strconv.Atoi(attributes["table_with_columns.0.excluded_column_names.#"])
			if err != nil {
				return nil, fmt.Errorf("acceptance test: could not convert string (%s) Atoi for excluded_column_names: %w", attributes["table_with_columns.0.excluded_column_names.#"], err)
			}
			for i := range colCount {
				excludedColumnNames = append(excludedColumnNames, attributes[fmt.Sprintf("table_with_columns.0.excluded_column_names.%d", i)])
			}
		}

		var columnWildcard bool
		if v := attributes["table_with_columns.0.wildcard"]; v == acctest.CtTrue {
			columnWildcard = true
		}

		return tflakeformation.FilterTableWithColumnsPermissions(principalIdentifier, tflakeformation.ExpandTableWithColumnsResourceAsTable(tfMap), columnNames, excludedColumnNames, columnWildcard), nil
	}
	return nil, nil
}

func testAccPermissionsConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal        = aws_iam_role.test.arn
  permissions      = ["CREATE_DATABASE"]
  catalog_resource = true

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_dataCellsFilter(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name    = "my_column_12"
      type    = "date"
      comment = "my_column1_comment2"
    }

    columns {
      name    = "my_column_22"
      type    = "timestamp"
      comment = "my_column2_comment2"
    }

    columns {
      name    = "my_column_23"
      type    = "string"
      comment = "my_column23_comment2"
    }
  }
}

resource "aws_lakeformation_data_cells_filter" "test" {
  table_data {
    database_name    = aws_glue_catalog_database.test.name
    name             = %[1]q
    table_catalog_id = data.aws_caller_identity.current.account_id
    table_name       = aws_glue_catalog_table.test.name

    column_names = ["my_column_22"]

    row_filter {
      filter_expression = "my_column_23='testing'"
    }
  }

  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_permissions" "test" {
  principal   = aws_iam_role.test.arn
  permissions = ["DESCRIBE"]

  data_cells_filter {
    database_name    = aws_lakeformation_data_cells_filter.test.table_data[0].database_name
    name             = aws_lakeformation_data_cells_filter.test.table_data[0].name
    table_catalog_id = aws_lakeformation_data_cells_filter.test.table_data[0].table_catalog_id
    table_name       = aws_lakeformation_data_cells_filter.test.table_data[0].table_name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_database(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALTER", "CREATE_TABLE", "DROP"]
  permissions_with_grant_option = ["CREATE_TABLE"]
  principal                     = aws_iam_role.test.arn

  database {
    name = aws_glue_catalog_database.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_databaseIAMAllowed(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "transactionamount"
      type = "double"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = "IAM_ALLOWED_PRINCIPALS"

  database {
    name = aws_glue_catalog_database.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_databaseIAMPrincipals(rName string) string {
	return fmt.Sprintf(`

data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "transactionamount"
      type = "double"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = "${data.aws_caller_identity.current.account_id}:IAMPrincipals"

  database {
    name = aws_glue_catalog_database.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_databaseMultiple(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test2" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALTER", "CREATE_TABLE", "DROP"]
  permissions_with_grant_option = ["CREATE_TABLE"]
  principal                     = aws_iam_role.test.arn

  database {
    name = aws_glue_catalog_database.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_permissions" "test2" {
  permissions = ["ALTER", "DROP"]
  principal   = aws_iam_role.test2.arn

  database {
    name = aws_glue_catalog_database.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_dataLocation(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
      }, {
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_lakeformation_resource" "test" {
  arn      = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal   = aws_iam_role.test.arn
  permissions = ["DATA_LOCATION_ACCESS"]

  data_location {
    arn = aws_s3_bucket.test.arn
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_lfTag(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "test" {
  key    = %[1]q
  values = ["value1", "value2"]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ASSOCIATE", "DESCRIBE"]
  permissions_with_grant_option = ["ASSOCIATE", "DESCRIBE"]
  principal                     = aws_iam_role.test.arn

  lf_tag {
    key    = aws_lakeformation_lf_tag.test.key
    values = aws_lakeformation_lf_tag.test.values
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_lfTagPolicy(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "test" {
  key    = %[1]q
  values = ["value1", "value2"]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALTER", "CREATE_TABLE", "DROP"]
  permissions_with_grant_option = ["CREATE_TABLE"]
  principal                     = aws_iam_role.test.arn

  lf_tag_policy {
    resource_type = "DATABASE"

    expression {
      key    = aws_lakeformation_lf_tag.test.key
      values = aws_lakeformation_lf_tag.test.values
    }
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [
    aws_lakeformation_data_lake_settings.test,
    aws_lakeformation_lf_tag.test,
  ]
}
`, rName)
}

func testAccPermissionsConfig_lfTagPolicyMultiple(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "test" {
  count = 3

  key    = "%[1]s-${count.index}"
  values = ["value1", "value2"]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALTER", "CREATE_TABLE", "DROP"]
  permissions_with_grant_option = ["CREATE_TABLE"]
  principal                     = aws_iam_role.test.arn

  lf_tag_policy {
    resource_type = "DATABASE"

    expression {
      key    = aws_lakeformation_lf_tag.test[2].key
      values = aws_lakeformation_lf_tag.test[2].values
    }

    expression {
      key    = aws_lakeformation_lf_tag.test[0].key
      values = aws_lakeformation_lf_tag.test[0].values
    }

    expression {
      key    = aws_lakeformation_lf_tag.test[1].key
      values = aws_lakeformation_lf_tag.test[1].values
    }
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [
    aws_lakeformation_data_lake_settings.test,
    aws_lakeformation_lf_tag.test,
  ]
}
`, rName)
}

func testAccPermissionsConfig_tableBasic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALTER", "DELETE", "DESCRIBE"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_tableIAMAllowed(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "transactionamount"
      type = "double"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = "IAM_ALLOWED_PRINCIPALS"

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}
`, rName)
}

func testAccPermissionsConfig_tableIAMPrincipals(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "transactionamount"
      type = "double"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = "${data.aws_caller_identity.current.account_id}:IAMPrincipals"

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}
`, rName)
}

func testAccPermissionsConfig_tableImplicit(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  # this will give the principal implicit permissions
  admins = [aws_iam_role.test.arn, data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal                     = aws_iam_role.test.arn
  permissions                   = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]
  permissions_with_grant_option = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_tableMultipleRoles(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test2" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALTER", "DELETE", "DESCRIBE"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_permissions" "test2" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test2.arn

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_tableSelectOnly(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_tableSelectPlus(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal = aws_iam_role.test.arn

  permissions                   = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]
  permissions_with_grant_option = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_tableWildcardNoSelect(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT"]
  permissions_with_grant_option = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT"]
  principal                     = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_tableWildcardSelectOnly(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal = aws_iam_role.test.arn

  permissions = ["SELECT"]

  table {
    database_name = aws_glue_catalog_table.test.database_name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_tableWildcardSelectPlus(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal = aws_iam_role.test.arn

  permissions                   = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]
  permissions_with_grant_option = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]

  table {
    database_name = aws_glue_catalog_table.test.database_name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_twcBasic(rName string, columns string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "transactionamount"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test.arn

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    column_names  = [%[2]s]
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, columns)
}

func testAccPermissionsConfig_twcImplicit(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  # this will give the principal implicit permissions
  admins = [aws_iam_role.test.arn, data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal                     = aws_iam_role.test.arn
  permissions                   = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]
  permissions_with_grant_option = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_twcWildcardExcludedColumns(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test.arn

  table_with_columns {
    database_name         = aws_glue_catalog_table.test.database_name
    name                  = aws_glue_catalog_table.test.name
    wildcard              = true
    excluded_column_names = ["value"]
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_twcWildcardSelectOnly(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test.arn

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_twcWildcardSelectPlus(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "ALL"]
  principal   = aws_iam_role.test.arn

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccPermissionsConfig_catalogResource_nonIAMPrincipals(rName string) string {
	return fmt.Sprintf(`
resource "aws_lakeformation_permissions" "test" {
  principal        = aws_identitystore_group.test.arn
  permissions      = ["DESCRIBE"]
  catalog_resource = true

  # for consistency, ensure that admins are setup before testing
  depends_on = [
    aws_lakeformation_data_lake_settings.test,
    aws_lakeformation_identity_center_configuration.test,
  ]
}

resource "aws_identitystore_group" "test" {
  identity_store_id = local.identity_store_id
  display_name      = %[1]q
}

resource "aws_lakeformation_identity_center_configuration" "test" {
  instance_arn = local.identity_center_instance_arn
}

locals {
  identity_center_instance_arn = data.aws_ssoadmin_instances.test.arns[0]
  identity_store_id            = data.aws_ssoadmin_instances.test.identity_store_ids[0]
}

data "aws_ssoadmin_instances" "test" {}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}
`, rName)
}

func testAccPermissionsConfig_table_nonIAMPrincipals(rName string) string {
	return fmt.Sprintf(`
resource "aws_lakeformation_permissions" "test" {
  principal   = aws_identitystore_group.test.arn
  permissions = ["DESCRIBE"]

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [
    aws_lakeformation_data_lake_settings.test,
    aws_lakeformation_identity_center_configuration.test,
  ]
}

resource "aws_identitystore_group" "test" {
  identity_store_id = local.identity_store_id
  display_name      = %[1]q
}

resource "aws_lakeformation_identity_center_configuration" "test" {
  instance_arn = local.identity_center_instance_arn
}

locals {
  identity_center_instance_arn = data.aws_ssoadmin_instances.test.arns[0]
  identity_store_id            = data.aws_ssoadmin_instances.test.identity_store_ids[0]
}

data "aws_ssoadmin_instances" "test" {}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "transactionamount"
      type = "double"
    }
  }
}
`, rName)
}
