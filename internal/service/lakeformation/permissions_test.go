// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPermissions_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionCreateDatabase)),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccPermissions_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcBasic(rName, "\"event\", \"timestamp\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflakeformation.ResourcePermissions(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPermissions_database(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	dbName := "aws_glue_catalog_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_database(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "database.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionAlter)),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", string(awstypes.PermissionCreateTable)),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", string(awstypes.PermissionDrop)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", string(awstypes.PermissionCreateTable)),
				),
			},
		},
	})
}

func testAccPermissions_databaseIAMAllowed(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dbName := "aws_glue_catalog_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_databaseIAMAllowed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, tflakeformation.IAMAllowedPrincipals),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "database.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionAll)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccPermissions_databaseMultiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	resourceName2 := "aws_lakeformation_permissions.test2"
	roleName := "aws_iam_role.test"
	roleName2 := "aws_iam_role.test2"
	dbName := "aws_glue_catalog_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_databaseMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "database.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionAlter)),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", string(awstypes.PermissionCreateTable)),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", string(awstypes.PermissionDrop)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", string(awstypes.PermissionCreateTable)),
					testAccCheckPermissionsExists(ctx, resourceName2),
					resource.TestCheckResourceAttrPair(resourceName2, names.AttrPrincipal, roleName2, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName2, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName2, names.AttrPrincipal, roleName2, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName2, "database.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName2, "database.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName2, "permissions.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName2, "permissions.0", string(awstypes.PermissionAlter)),
					resource.TestCheckResourceAttr(resourceName2, "permissions.1", string(awstypes.PermissionDrop)),
					resource.TestCheckResourceAttr(resourceName2, "permissions_with_grant_option.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccPermissions_dataCellsFilter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_dataCellsFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionDescribe)),
					resource.TestCheckResourceAttr(resourceName, "data_cells_filter.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccPermissions_dataLocation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	bucketName := "aws_s3_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_dataLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionDataLocationAccess)),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "data_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_location.0.arn", bucketName, names.AttrARN),
				),
			},
		},
	})
}

func testAccPermissions_lfTag(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tagName := "aws_lakeformation_lf_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_lfTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lf_tag.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag.0.key", tagName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag.0.values", tagName, names.AttrValues),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "ASSOCIATE"),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", "DESCRIBE"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", "ASSOCIATE"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.1", "DESCRIBE"),
				),
			},
		},
	})
}

func testAccPermissions_lfTagPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tagName := "aws_lakeformation_lf_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_lfTagPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.0.resource_type", "DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.0.expression.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag_policy.0.expression.0.key", tagName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag_policy.0.expression.0.values", tagName, names.AttrValues),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionAlter)),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", string(awstypes.PermissionCreateTable)),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", string(awstypes.PermissionDrop)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", string(awstypes.PermissionCreateTable)),
				),
			},
		},
	})
}

func testAccPermissions_lfTagPolicyMultiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	// tagName := "aws_lakeformation_lf_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_lfTagPolicyMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.0.resource_type", "DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "lf_tag_policy.0.expression.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag_policy.0.expression.*", map[string]string{
						names.AttrKey: rName + "-0",
						"values.#":    acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag_policy.0.expression.*", map[string]string{
						names.AttrKey: rName + "-1",
						"values.#":    acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag_policy.0.expression.*", map[string]string{
						names.AttrKey: rName + "-2",
						"values.#":    acctest.Ct2,
					}),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionAlter)),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", string(awstypes.PermissionCreateTable)),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", string(awstypes.PermissionDrop)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", string(awstypes.PermissionCreateTable)),
				),
			},
		},
	})
}

func testAccPermissions_tableBasic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(roleName, names.AttrARN, resourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttr(resourceName, "table.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionAlter)),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", string(awstypes.PermissionDelete)),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", string(awstypes.PermissionDescribe)),
				),
			},
		},
	})
}

func testAccPermissions_tableIAMAllowed(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dbName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableIAMAllowed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, tflakeformation.IAMAllowedPrincipals),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "table.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", dbName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", dbName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionAll)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccPermissions_tableImplicit(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableImplicit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table.#", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	resourceName2 := "aws_lakeformation_permissions.test2"
	roleName := "aws_iam_role.test"
	roleName2 := "aws_iam_role.test2"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableMultipleRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(roleName, names.AttrARN, resourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttr(resourceName, "table.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionAlter)),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", string(awstypes.PermissionDelete)),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", string(awstypes.PermissionDescribe)),
					testAccCheckPermissionsExists(ctx, resourceName2),
					resource.TestCheckResourceAttrPair(roleName2, names.AttrARN, resourceName2, names.AttrPrincipal),
					resource.TestCheckResourceAttr(resourceName2, "table.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName2, "table.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName2, "table.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName2, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName2, "permissions.0", string(awstypes.PermissionSelect)),
				),
			},
		},
	})
}

func testAccPermissions_tableSelectOnly(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableSelectOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(roleName, names.AttrARN, resourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttr(resourceName, "table.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionSelect)),
				),
			},
		},
	})
}

func testAccPermissions_tableSelectPlus(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableSelectPlus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	databaseResourceName := "aws_glue_catalog_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableWildcardNoSelect(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "table.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", databaseResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table.0.wildcard", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccPermissions_tableWildcardSelectOnly(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableWildcardSelectOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionSelect)),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccPermissions_tableWildcardSelectPlus(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_tableWildcardSelectPlus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "7"),
				),
			},
		},
	})
}

func testAccPermissions_twcBasic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcBasic(rName, "\"event\", \"timestamp\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionSelect)),
				),
			},
			{
				Config: testAccPermissionsConfig_twcBasic(rName, "\"timestamp\", \"event\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionSelect)),
				),
			},
			{
				Config: testAccPermissionsConfig_twcBasic(rName, "\"timestamp\", \"event\", \"transactionamount\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.2", "transactionamount"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionSelect)),
				),
			},
			{
				Config: testAccPermissionsConfig_twcBasic(rName, "\"event\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionSelect)),
				),
			},
		},
	})
}

func testAccPermissions_twcImplicit(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcImplicit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcWildcardExcludedColumns(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccPermissions_twcWildcardSelectOnly(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcWildcardSelectOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.wildcard", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", string(awstypes.PermissionSelect)),
				),
			},
		},
	})
}

func testAccPermissions_twcWildcardSelectPlus(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsConfig_twcWildcardSelectPlus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckPermissionsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_permissions" {
				continue
			}

			permCount, err := permissionCountForResource(ctx, conn, rs)

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

func testAccCheckPermissionsExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("acceptance test: resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		permCount, err := permissionCountForResource(ctx, conn, rs)

		if err != nil {
			return fmt.Errorf("acceptance test: error listing Lake Formation permissions (%s): %w", rs.Primary.ID, err)
		}

		if permCount == 0 {
			return fmt.Errorf("acceptance test: Lake Formation permissions (%s) do not exist or could not be found", rs.Primary.ID)
		}

		return nil
	}
}

func permissionCountForResource(ctx context.Context, conn *lakeformation.Client, rs *terraform.ResourceState) (int, error) {
	input := &lakeformation.ListPermissionsInput{
		Principal: &awstypes.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(rs.Primary.Attributes[names.AttrPrincipal]),
		},
		Resource: &awstypes.Resource{},
	}

	noResource := true

	if v, ok := rs.Primary.Attributes[names.AttrCatalogID]; ok && v != "" {
		input.CatalogId = aws.String(v)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["catalog_resource"]; ok && v != "" && v == acctest.CtTrue {
		input.Resource.Catalog = tflakeformation.ExpandCatalogResource()

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["data_location.#"]; ok && v != "" && v != acctest.Ct0 {
		tfMap := map[string]interface{}{}

		if v := rs.Primary.Attributes["data_location.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := rs.Primary.Attributes["data_location.0.arn"]; v != "" {
			tfMap[names.AttrARN] = v
		}

		input.Resource.DataLocation = tflakeformation.ExpandDataLocationResource(tfMap)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["database.#"]; ok && v != "" && v != acctest.Ct0 {
		tfMap := map[string]interface{}{}

		if v := rs.Primary.Attributes["database.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := rs.Primary.Attributes["database.0.name"]; v != "" {
			tfMap[names.AttrName] = v
		}

		input.Resource.Database = tflakeformation.ExpandDatabaseResource(tfMap)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["lf_tag.#"]; ok && v != "" && v != acctest.Ct0 {
		tfMap := map[string]interface{}{}

		if v := rs.Primary.Attributes["lf_tag.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := rs.Primary.Attributes["lf_tag.0.key"]; v != "" {
			tfMap[names.AttrKey] = v
		}

		if count, err := strconv.Atoi(rs.Primary.Attributes["lf_tag.0.values.#"]); err == nil {
			var tagValues []string
			for i := 0; i < count; i++ {
				tagValues = append(tagValues, rs.Primary.Attributes[fmt.Sprintf("lf_tag.0.values.%d", i)])
			}
			tfMap[names.AttrValues] = flex.FlattenStringSet(aws.StringSlice(tagValues))
		}

		input.Resource.LFTag = tflakeformation.ExpandLFTagKeyResource(tfMap)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["lf_tag_policy.#"]; ok && v != "" && v != acctest.Ct0 {
		tfMap := map[string]interface{}{}

		if v := rs.Primary.Attributes["lf_tag_policy.0.catalog_id"]; v != "" {
			tfMap[names.AttrCatalogID] = v
		}

		if v := rs.Primary.Attributes["lf_tag_policy.0.resource_type"]; v != "" {
			tfMap[names.AttrResourceType] = v
		}

		if expressionCount, err := strconv.Atoi(rs.Primary.Attributes["lf_tag_policy.0.expression.#"]); err == nil {
			expressionSlice := make([]interface{}, expressionCount)
			for i := 0; i < expressionCount; i++ {
				expression := make(map[string]interface{})

				if v := rs.Primary.Attributes[fmt.Sprintf("lf_tag_policy.0.expression.%d.key", i)]; v != "" {
					expression[names.AttrKey] = v
				}

				if expressionValueCount, err := strconv.Atoi(rs.Primary.Attributes[fmt.Sprintf("lf_tag_policy.0.expression.%d.values.#", i)]); err == nil {
					valueSlice := make([]string, expressionValueCount)
					for j := 0; j < expressionValueCount; j++ {
						valueSlice[j] = rs.Primary.Attributes[fmt.Sprintf("lf_tag_policy.0.expression.%d.values.%d", i, j)]
					}
					expression[names.AttrValues] = flex.FlattenStringSet(aws.StringSlice(valueSlice))
				}
				expressionSlice[i] = expression
			}
			// The exact details of the set hash function don't matter, elements just have distinct values.
			tfMap[names.AttrExpression] = schema.NewSet(func(_ interface{}) int { return sdkacctest.RandInt() }, expressionSlice)
		}

		input.Resource.LFTagPolicy = tflakeformation.ExpandLFTagPolicyResource(tfMap)

		noResource = false
	}

	tableType := ""

	if v, ok := rs.Primary.Attributes["table.#"]; ok && v != "" && v != acctest.Ct0 {
		tableType = tflakeformation.TableTypeTable

		tfMap := map[string]interface{}{}

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

	if v, ok := rs.Primary.Attributes["table_with_columns.#"]; ok && v != "" && v != acctest.Ct0 {
		tableType = tflakeformation.TableTypeTableWithColumns

		tfMap := map[string]interface{}{}

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

	if v, ok := rs.Primary.Attributes["data_cells_filter.#"]; ok && v != "" && v != acctest.Ct0 {
		tfMap := map[string]interface{}{}

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

		input.Resource.DataCellsFilter = tflakeformation.ExpandDataCellsFilter([]interface{}{tfMap})
		noResource = false
	}

	if noResource {
		// if after read, there is no resource, it has been deleted
		return 0, nil
	}

	log.Printf("[DEBUG] Reading Lake Formation permissions: %v", input)
	var allPermissions []awstypes.PrincipalResourcePermissions

	err := retry.RetryContext(ctx, tflakeformation.IAMPropagationTimeout, func() *retry.RetryError {
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
				return retry.RetryableError(err)
			}

			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("acceptance test: error listing Lake Formation Permissions getting permission count: %w", err))
			}

			for _, permission := range page.PrincipalResourcePermissions {
				if reflect.ValueOf(permission).IsZero() {
					continue
				}

				allPermissions = append(allPermissions, permission)
			}
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.ListPermissions(ctx, input)
	}

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return 0, nil
	}

	// used to find error in retryable error
	if tfawserr.ErrMessageContains(err, "AccessDeniedException", "Resource does not exist") {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("acceptance test: error listing Lake Formation permissions after retry %v: %w", input, err)
	}

	columnNames := make([]string, 0)
	excludedColumnNames := make([]string, 0)
	columnWildcard := false

	if tableType == tflakeformation.TableTypeTableWithColumns {
		if v := rs.Primary.Attributes["table_with_columns.0.wildcard"]; v != "" && v == acctest.CtTrue {
			columnWildcard = true
		}

		colCount := 0

		if v := rs.Primary.Attributes["table_with_columns.0.column_names.#"]; v != "" {
			colCount, err = strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.column_names.#"])

			if err != nil {
				return 0, fmt.Errorf("acceptance test: could not convert string (%s) Atoi for column_names: %w", rs.Primary.Attributes["table_with_columns.0.column_names.#"], err)
			}
		}

		for i := 0; i < colCount; i++ {
			columnNames = append(columnNames, rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.column_names.%d", i)])
		}

		colCount = 0

		if v := rs.Primary.Attributes["table_with_columns.0.excluded_column_names.#"]; v != "" {
			colCount, err = strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.excluded_column_names.#"])

			if err != nil {
				return 0, fmt.Errorf("acceptance test: could not convert string (%s) Atoi for excluded_column_names: %w", rs.Primary.Attributes["table_with_columns.0.excluded_column_names.#"], err)
			}
		}

		for i := 0; i < colCount; i++ {
			excludedColumnNames = append(excludedColumnNames, rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.excluded_column_names.%d", i)])
		}
	}

	// clean permissions = filter out permissions that do not pertain to this specific resource
	cleanPermissions := tflakeformation.FilterPermissions(input, tableType, columnNames, excludedColumnNames, columnWildcard, allPermissions)

	return len(cleanPermissions), nil
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
  permissions = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]
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
