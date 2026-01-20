// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lakeformation_permissions", name="Permissions")
func ResourcePermissions() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePermissionsCreate,
		ReadWithoutTimeout:   resourcePermissionsRead,
		DeleteWithoutTimeout: resourcePermissionsDelete,

		Schema: map[string]*schema.Schema{
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"catalog_resource": {
				Type:     schema.TypeBool,
				Default:  false,
				ForceNew: true,
				Optional: true,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					names.AttrDatabase,
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
					"data_cells_filter",
				},
			},
			"data_cells_filter": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDatabaseName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"table_catalog_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrTableName: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"data_location": {
				Type:     schema.TypeList,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Optional: true,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					names.AttrDatabase,
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
					"data_cells_filter",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrCatalogID: {
							Type:     schema.TypeString,
							Computed: true,
							ForceNew: true,
							Optional: true,
						},
					},
				},
			},
			names.AttrDatabase: {
				Type:     schema.TypeList,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Optional: true,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					names.AttrDatabase,
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
					"data_cells_filter",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:     schema.TypeString,
							Computed: true,
							ForceNew: true,
							Optional: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
					},
				},
			},
			"lf_tag": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					names.AttrDatabase,
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
					"data_cells_filter",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
							Computed: true,
						},
						names.AttrKey: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						names.AttrValues: {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateLFTagValues(),
							},
						},
					},
				},
			},
			"lf_tag_policy": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					names.AttrDatabase,
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
					"data_cells_filter",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						names.AttrExpression: {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKey: {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									names.AttrValues: {
										Type:     schema.TypeSet,
										Required: true,
										ForceNew: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validateLFTagValues(),
										},
									},
								},
							},
						},
						names.AttrResourceType: {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ResourceType](),
						},
					},
				},
			},
			names.AttrPermissions: {
				Type:     schema.TypeSet,
				ForceNew: true,
				MinItems: 1,
				Required: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.Permission](),
				},
			},
			"permissions_with_grant_option": {
				Type:     schema.TypeSet,
				Computed: true,
				ForceNew: true,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.Permission](),
				},
			},
			names.AttrPrincipal: {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validPrincipal,
			},
			"table": {
				Type:     schema.TypeList,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Optional: true,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					names.AttrDatabase,
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
					"data_cells_filter",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:     schema.TypeString,
							Computed: true,
							ForceNew: true,
							Optional: true,
						},
						names.AttrDatabaseName: {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
							ForceNew: true,
							Optional: true,
							AtLeastOneOf: []string{
								"table.0.name",
								"table.0.wildcard",
							},
						},
						"wildcard": {
							Type:     schema.TypeBool,
							Default:  false,
							ForceNew: true,
							Optional: true,
							AtLeastOneOf: []string{
								"table.0.name",
								"table.0.wildcard",
							},
						},
					},
				},
			},
			"table_with_columns": {
				Type:     schema.TypeList,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Optional: true,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					names.AttrDatabase,
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
					"data_cells_filter",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:     schema.TypeString,
							Computed: true,
							ForceNew: true,
							Optional: true,
						},
						"column_names": {
							Type:     schema.TypeSet,
							ForceNew: true,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
							AtLeastOneOf: []string{
								"table_with_columns.0.column_names",
								"table_with_columns.0.wildcard",
							},
						},
						names.AttrDatabaseName: {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"excluded_column_names": {
							Type:     schema.TypeSet,
							ForceNew: true,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
						},
						names.AttrName: {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"wildcard": {
							Type:     schema.TypeBool,
							Default:  false,
							ForceNew: true,
							Optional: true,
							AtLeastOneOf: []string{
								"table_with_columns.0.column_names",
								"table_with_columns.0.wildcard",
							},
						},
					},
				},
			},
		},
	}
}

// The challenges with Lake Formation permissions are many. These are largely undocumented and
// discovered through trial and error. These are specific problems discovered thus far:
// 1. Implicit permissions granted by Lake Formation to data lake administrators are indistinguishable
//    from explicit permissions. However, implicit permissions cannot be changed, revoked, or narrowed.
// 2. One set of permissions for one LF Resource going in, can come back from AWS as multiple sets of
//    permissions for multiple LF Resources (e.g., SELECT, Table, TableWithColumns).
// 3. Valid permissions for a Table LF resource can come back in TableWithColumns and vice versa.

// For 2 & 3, some peeking at the config (i.e., d.Get()) is necessary to filter the permissions AWS
// returns.

func resourcePermissionsCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	input := &lakeformation.GrantPermissionsInput{
		Permissions: flex.ExpandStringyValueSet[awstypes.Permission](d.Get(names.AttrPermissions).(*schema.Set)),
		Principal: &awstypes.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(d.Get(names.AttrPrincipal).(string)),
		},
	}

	if v, ok := d.GetOk(names.AttrCatalogID); ok {
		input.CatalogId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("permissions_with_grant_option"); ok {
		input.PermissionsWithGrantOption = flex.ExpandStringyValueSet[awstypes.Permission](v.(*schema.Set))
	}

	populateResourceForCreate(input, d)

	var output *lakeformation.GrantPermissionsOutput
	err := tfresource.Retry(ctx, IAMPropagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		var err error
		output, err = conn.GrantPermissions(ctx, input)
		if err != nil {
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Invalid principal") {
				return tfresource.RetryableError(err)
			}
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Grantee has no permissions") {
				return tfresource.RetryableError(err)
			}
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "register the S3 path") {
				return tfresource.RetryableError(err)
			}
			if errs.IsA[*awstypes.ConcurrentModificationException](err) {
				return tfresource.RetryableError(err)
			}
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not authorized to access requested permissions") {
				return tfresource.RetryableError(err)
			}

			return tfresource.NonRetryableError(fmt.Errorf("creating Lake Formation Permissions: %w", err))
		}
		return nil
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lake Formation Permissions: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Lake Formation Permissions: empty response")
	}

	d.SetId(strconv.Itoa(create.StringHashcode(prettify(input))))

	return append(diags, resourcePermissionsRead(ctx, d, meta)...)
}

func resourcePermissionsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	var input lakeformation.ListPermissionsInput

	principalIdentifier := d.Get(names.AttrPrincipal).(string)
	if includePrincipalIdentifierInList(principalIdentifier) {
		principal := awstypes.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(principalIdentifier),
		}
		input.Principal = &principal
	}

	if v, ok := d.GetOk(names.AttrCatalogID); ok {
		input.CatalogId = aws.String(v.(string))
	}

	populateResourceForRead(&input, d)

	filter := permissionsFilter(d)

	permissions, err := waitPermissionsReady(ctx, conn, &input, filter)

	if !d.IsNewResource() {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			log.Printf("[WARN] Resource Lake Formation permissions (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Resource does not exist") {
			log.Printf("[WARN] Resource Lake Formation permissions (%s) not found, removing from state: %s", d.Id(), err)
			d.SetId("")
			return diags
		}

		if len(permissions) == 0 {
			log.Printf("[WARN] Resource Lake Formation permissions (%s) not found, removing from state (0 permissions)", d.Id())
			d.SetId("")
			return diags
		}
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lake Formation permissions: %s", err)
	}

	d.Set(names.AttrPrincipal, permissions[0].Principal.DataLakePrincipalIdentifier)
	d.Set(names.AttrPermissions, flattenResourcePermissions(permissions))
	d.Set("permissions_with_grant_option", flattenGrantPermissions(permissions))

	if permissions[0].Resource.Catalog != nil {
		d.Set("catalog_resource", true)
	} else {
		d.Set("catalog_resource", false)
	}

	if permissions[0].Resource.DataLocation != nil {
		if err := d.Set("data_location", []any{flattenDataLocationResource(permissions[0].Resource.DataLocation)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting data_location: %s", err)
		}
	} else {
		d.Set("data_location", nil)
	}

	if permissions[0].Resource.Database != nil {
		if err := d.Set(names.AttrDatabase, []any{flattenDatabaseResource(permissions[0].Resource.Database)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting database: %s", err)
		}
	} else {
		d.Set(names.AttrDatabase, nil)
	}

	if permissions[0].Resource.DataCellsFilter != nil {
		if err := d.Set("data_cells_filter", flattenDataCellsFilter(permissions[0].Resource.DataCellsFilter)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting data_cells_filter: %s", err)
		}
	} else {
		d.Set("data_cells_filter", nil)
	}

	if permissions[0].Resource.LFTag != nil {
		if err := d.Set("lf_tag", []any{flattenLFTagKeyResource(permissions[0].Resource.LFTag)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting database: %s", err)
		}
	} else {
		d.Set("lf_tag", nil)
	}

	if permissions[0].Resource.LFTagPolicy != nil {
		if err := d.Set("lf_tag_policy", []any{flattenLFTagPolicyResource(permissions[0].Resource.LFTagPolicy)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting database: %s", err)
		}
	} else {
		d.Set("lf_tag_policy", nil)
	}

	tableSet := false

	if v, ok := d.GetOk("table"); ok && len(v.([]any)) > 0 {
		// since perm list could include TableWithColumns, get the right one
		for _, perm := range permissions {
			if perm.Resource == nil {
				continue
			}

			if perm.Resource.TableWithColumns != nil && perm.Resource.TableWithColumns.ColumnWildcard != nil {
				if err := d.Set("table", []any{flattenTableColumnsResourceAsTable(perm.Resource.TableWithColumns)}); err != nil {
					return sdkdiag.AppendErrorf(diags, "setting table: %s", err)
				}
				tableSet = true
				break
			}

			if perm.Resource.Table != nil {
				if err := d.Set("table", []any{flattenTableResource(perm.Resource.Table)}); err != nil {
					return sdkdiag.AppendErrorf(diags, "setting table: %s", err)
				}
				tableSet = true
				break
			}
		}
	}

	if !tableSet {
		d.Set("table", nil)
	}

	twcSet := false

	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]any)) > 0 {
		// since perm list could include Table, get the right one
		for _, perm := range permissions {
			if perm.Resource.TableWithColumns != nil {
				if err := d.Set("table_with_columns", []any{flattenTableColumnsResource(perm.Resource.TableWithColumns)}); err != nil {
					return sdkdiag.AppendErrorf(diags, "setting table_with_columns: %s", err)
				}
				twcSet = true
				break
			}
		}
	}

	if !twcSet {
		d.Set("table_with_columns", nil)
	}

	return diags
}

func resourcePermissionsDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	input := &lakeformation.RevokePermissionsInput{
		Permissions:                flex.ExpandStringyValueSet[awstypes.Permission](d.Get(names.AttrPermissions).(*schema.Set)),
		PermissionsWithGrantOption: flex.ExpandStringyValueSet[awstypes.Permission](d.Get("permissions_with_grant_option").(*schema.Set)),
		Principal: &awstypes.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(d.Get(names.AttrPrincipal).(string)),
		},
	}

	if v, ok := d.GetOk(names.AttrCatalogID); ok {
		input.CatalogId = aws.String(v.(string))
	}

	populateResourceForDelete(input, d)

	if input.Resource == nil || reflect.DeepEqual(input.Resource, &awstypes.Resource{}) {
		// if resource is empty, don't delete = it won't delete anything since this is the predicate
		log.Printf("[WARN] No Lake Formation Resource with permissions to revoke")
		return diags
	}

	err := tfresource.Retry(ctx, permissionsDeleteRetryTimeout, func(ctx context.Context) *tfresource.RetryError {
		var err error
		_, err = conn.RevokePermissions(ctx, input)
		if err != nil {
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "register the S3 path") {
				return tfresource.RetryableError(err)
			}
			if errs.IsA[*awstypes.ConcurrentModificationException](err) {
				return tfresource.RetryableError(err)
			}
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not authorized to access requested permissions") {
				return tfresource.RetryableError(err)
			}

			return tfresource.NonRetryableError(fmt.Errorf("unable to revoke Lake Formation Permissions: %w", err))
		}
		return nil
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "No permissions revoked. Grantee") {
		return diags
	}

	if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "cannot grant/revoke permission on non-existent column") {
		return diags
	}

	if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Cell Filter not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "unable to revoke LakeFormation Permissions (input: %v): %s", input, err)
	}

	// Attempted to add a waiter here to wait for delete to complete. However, ListPermissions() returns
	// permissions, at least for catalog/CREATE_DATABASE permission, even if they do not exist. That makes
	// knowing when the delete is complete impossible. Instead, we'll retry until we get the right error.

	// Knowing when the delete is complete is complicated:
	// You can't just wait until permissions = 0 because there could be many other unrelated permissions
	// on the resource and filtering is non-trivial for table with columns.

	err = tfresource.Retry(ctx, permissionsDeleteRetryTimeout, func(ctx context.Context) *tfresource.RetryError {
		var err error
		_, err = conn.RevokePermissions(ctx, input)

		if !errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "No permissions revoked. Grantee has no") {
			return tfresource.RetryableError(err)
		}

		return nil
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "No permissions revoked. Grantee") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "unable to revoke LakeFormation Permissions (input: %v): %s", input, err)
	}

	return diags
}

type PermissionsFilter = tfslices.Predicate[awstypes.PrincipalResourcePermissions]

func permissionsFilter(d *schema.ResourceData) PermissionsFilter {
	principalIdentifier := d.Get(names.AttrPrincipal).(string)

	if _, ok := d.GetOk("catalog_resource"); ok {
		return filterCatalogPermissions(principalIdentifier)
	}
	if _, ok := d.GetOk("data_cells_filter"); ok {
		return filterDataCellsFilter(principalIdentifier)
	}
	if v, ok := d.GetOk("data_location"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		return filterDataLocationPermissions(principalIdentifier)
	}
	if v, ok := d.GetOk(names.AttrDatabase); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		return filterDatabasePermissions(principalIdentifier)
	}
	if v, ok := d.GetOk("lf_tag"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		return filterLFTagPermissions(principalIdentifier)
	}
	if v, ok := d.GetOk("lf_tag_policy"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		return filterLFTagPolicyPermissions(principalIdentifier)
	}
	if v, ok := d.GetOk("table"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		return filterTablePermissions(principalIdentifier, ExpandTableResource(v.([]any)[0].(map[string]any)))
	}
	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		var columnNames []string
		if v, ok := d.GetOk("table_with_columns.0.column_names"); ok {
			if v, ok := v.(*schema.Set); ok && v.Len() > 0 {
				columnNames = flex.ExpandStringValueSet(v)
			}
		}

		var excludedColumnNames []string
		if v, ok := d.GetOk("table_with_columns.0.excluded_column_names"); ok {
			if v, ok := v.(*schema.Set); ok && v.Len() > 0 {
				excludedColumnNames = flex.ExpandStringValueSet(v)
			}
		}

		var columnWildcard bool
		if v, ok := d.GetOk("table_with_columns.0.wildcard"); ok {
			columnWildcard = v.(bool)
		}

		return filterTableWithColumnsPermissions(principalIdentifier, ExpandTableWithColumnsResourceAsTable(v.([]any)[0].(map[string]any)), columnNames, excludedColumnNames, columnWildcard)
	}
	return nil
}

func populateResourceForCreate(input *lakeformation.GrantPermissionsInput, d *schema.ResourceData) {
	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Resource = &awstypes.Resource{
			TableWithColumns: expandTableColumnsResource(v.([]any)[0].(map[string]any)),
		}
	} else {
		input.Resource = expandResource(d)
	}
}

func populateResourceForRead(input *lakeformation.ListPermissionsInput, d *schema.ResourceData) {
	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		// can't ListPermissions for TableWithColumns, so use Table instead
		input.Resource = &awstypes.Resource{
			Table: ExpandTableWithColumnsResourceAsTable(v.([]any)[0].(map[string]any)),
		}
	} else {
		input.Resource = expandResource(d)
	}
}

func populateResourceForDelete(input *lakeformation.RevokePermissionsInput, d *schema.ResourceData) {
	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Resource = &awstypes.Resource{
			TableWithColumns: expandTableColumnsResource(v.([]any)[0].(map[string]any)),
		}
	} else {
		input.Resource = expandResource(d)
	}
}

func expandResource(d *schema.ResourceData) *awstypes.Resource {
	var resource *awstypes.Resource

	if _, ok := d.GetOk("catalog_resource"); ok {
		resource = &awstypes.Resource{
			Catalog: ExpandCatalogResource(),
		}
	} else if v, ok := d.GetOk("data_cells_filter"); ok {
		resource = &awstypes.Resource{
			DataCellsFilter: ExpandDataCellsFilter(v.([]any)),
		}
	} else if v, ok := d.GetOk("data_location"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		resource = &awstypes.Resource{
			DataLocation: ExpandDataLocationResource(v.([]any)[0].(map[string]any)),
		}
	} else if v, ok := d.GetOk(names.AttrDatabase); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		resource = &awstypes.Resource{
			Database: ExpandDatabaseResource(v.([]any)[0].(map[string]any)),
		}
	} else if v, ok := d.GetOk("lf_tag"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		resource = &awstypes.Resource{
			LFTag: ExpandLFTagKeyResource(v.([]any)[0].(map[string]any)),
		}
	} else if v, ok := d.GetOk("lf_tag_policy"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		resource = &awstypes.Resource{
			LFTagPolicy: ExpandLFTagPolicyResource(v.([]any)[0].(map[string]any)),
		}
	} else if v, ok := d.GetOk("table"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		resource = &awstypes.Resource{
			Table: ExpandTableResource(v.([]any)[0].(map[string]any)),
		}
	}

	return resource
}

func ExpandCatalogResource() *awstypes.CatalogResource {
	return &awstypes.CatalogResource{}
}

func ExpandDataCellsFilter(in []any) *awstypes.DataCellsFilterResource {
	if len(in) == 0 {
		return nil
	}

	m := in[0].(map[string]any)
	var out awstypes.DataCellsFilterResource

	if v, ok := m[names.AttrDatabaseName].(string); ok && v != "" {
		out.DatabaseName = aws.String(v)
	}

	if v, ok := m[names.AttrName].(string); ok && v != "" {
		out.Name = aws.String(v)
	}

	if v, ok := m["table_catalog_id"].(string); ok && v != "" {
		out.TableCatalogId = aws.String(v)
	}

	if v, ok := m[names.AttrTableName].(string); ok && v != "" {
		out.TableName = aws.String(v)
	}

	return &out
}

func flattenDataCellsFilter(in *awstypes.DataCellsFilterResource) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		names.AttrDatabaseName: aws.ToString(in.DatabaseName),
		names.AttrName:         aws.ToString(in.Name),
		"table_catalog_id":     aws.ToString(in.TableCatalogId),
		names.AttrTableName:    aws.ToString(in.TableName),
	}

	return []any{m}
}

func ExpandDataLocationResource(tfMap map[string]any) *awstypes.DataLocationResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataLocationResource{}

	if v, ok := tfMap[names.AttrCatalogID].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.ResourceArn = aws.String(v)
	}

	return apiObject
}

func flattenDataLocationResource(apiObject *awstypes.DataLocationResource) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CatalogId; v != nil {
		tfMap[names.AttrCatalogID] = aws.ToString(v)
	}

	if v := apiObject.ResourceArn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}

	return tfMap
}

func ExpandDatabaseResource(tfMap map[string]any) *awstypes.DatabaseResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DatabaseResource{}

	if v, ok := tfMap[names.AttrCatalogID].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func flattenDatabaseResource(apiObject *awstypes.DatabaseResource) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CatalogId; v != nil {
		tfMap[names.AttrCatalogID] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

func ExpandLFTagPolicyResource(tfMap map[string]any) *awstypes.LFTagPolicyResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LFTagPolicyResource{}

	if v, ok := tfMap[names.AttrCatalogID].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrExpression].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Expression = ExpandLFTagExpression(v.List())
	}

	if v, ok := tfMap[names.AttrResourceType].(string); ok && v != "" {
		apiObject.ResourceType = awstypes.ResourceType(v)
	}

	return apiObject
}

func ExpandLFTagExpression(expression []any) []awstypes.LFTag {
	tagSlice := []awstypes.LFTag{}
	for _, element := range expression {
		elementMap := element.(map[string]any)

		tag := awstypes.LFTag{
			TagKey:    aws.String(elementMap[names.AttrKey].(string)),
			TagValues: flex.ExpandStringValueSet(elementMap[names.AttrValues].(*schema.Set)),
		}

		tagSlice = append(tagSlice, tag)
	}

	return tagSlice
}

func flattenLFTagPolicyResource(apiObject *awstypes.LFTagPolicyResource) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CatalogId; v != nil {
		tfMap[names.AttrCatalogID] = aws.ToString(v)
	}

	if v := apiObject.Expression; v != nil {
		tfMap[names.AttrExpression] = flattenLFTagExpression(v)
	}

	if v := apiObject.ResourceType; v != "" {
		tfMap[names.AttrResourceType] = string(v)
	}

	return tfMap
}

func flattenLFTagExpression(ts []awstypes.LFTag) []map[string]any {
	tagSlice := make([]map[string]any, len(ts))
	if len(ts) > 0 {
		for i, t := range ts {
			tag := make(map[string]any)

			if v := aws.ToString(t.TagKey); v != "" {
				tag[names.AttrKey] = v
			}

			if v := flex.FlattenStringValueList(t.TagValues); v != nil {
				tag[names.AttrValues] = v
			}

			tagSlice[i] = tag
		}
	}

	return tagSlice
}

func ExpandLFTagKeyResource(tfMap map[string]any) *awstypes.LFTagKeyResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LFTagKeyResource{}

	if v, ok := tfMap[names.AttrCatalogID].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.TagKey = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValues].(*schema.Set); ok && v != nil {
		apiObject.TagValues = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func flattenLFTagKeyResource(apiObject *awstypes.LFTagKeyResource) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CatalogId; v != nil {
		tfMap[names.AttrCatalogID] = aws.ToString(v)
	}

	if v := apiObject.TagKey; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	if v := apiObject.TagValues; v != nil {
		tfMap[names.AttrValues] = flex.FlattenStringValueSet(v)
	}

	return tfMap
}

func ExpandTableResource(tfMap map[string]any) *awstypes.TableResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TableResource{}

	if v, ok := tfMap[names.AttrCatalogID].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrDatabaseName].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["wildcard"].(bool); ok && v {
		apiObject.TableWildcard = &awstypes.TableWildcard{}
	}

	return apiObject
}

func ExpandTableWithColumnsResourceAsTable(tfMap map[string]any) *awstypes.TableResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TableResource{}

	if v, ok := tfMap[names.AttrCatalogID].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrDatabaseName].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func flattenTableResource(apiObject *awstypes.TableResource) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CatalogId; v != nil {
		tfMap[names.AttrCatalogID] = aws.ToString(v)
	}

	if v := apiObject.DatabaseName; v != nil {
		tfMap[names.AttrDatabaseName] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		if aws.ToString(v) != TableNameAllTables || apiObject.TableWildcard == nil {
			tfMap[names.AttrName] = aws.ToString(v)
		}
	}

	if v := apiObject.TableWildcard; v != nil {
		tfMap["wildcard"] = true
	}

	return tfMap
}

func expandTableColumnsResource(tfMap map[string]any) *awstypes.TableWithColumnsResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TableWithColumnsResource{}

	if v, ok := tfMap[names.AttrCatalogID].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["column_names"]; ok {
		if v, ok := v.(*schema.Set); ok && v.Len() > 0 {
			apiObject.ColumnNames = flex.ExpandStringValueSet(v)
		}
	}

	if v, ok := tfMap[names.AttrDatabaseName].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap["excluded_column_names"]; ok {
		if v, ok := v.(*schema.Set); ok && v.Len() > 0 {
			apiObject.ColumnWildcard = &awstypes.ColumnWildcard{
				ExcludedColumnNames: flex.ExpandStringValueSet(v),
			}
		}
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["wildcard"].(bool); ok && v && apiObject.ColumnWildcard == nil {
		apiObject.ColumnWildcard = &awstypes.ColumnWildcard{}
	}

	return apiObject
}

func flattenTableColumnsResource(apiObject *awstypes.TableWithColumnsResource) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CatalogId; v != nil {
		tfMap[names.AttrCatalogID] = aws.ToString(v)
	}

	tfMap["column_names"] = flex.FlattenStringValueSet(apiObject.ColumnNames)

	if v := apiObject.DatabaseName; v != nil {
		tfMap[names.AttrDatabaseName] = aws.ToString(v)
	}

	if v := apiObject.ColumnWildcard; v != nil {
		tfMap["wildcard"] = true
		tfMap["excluded_column_names"] = flex.FlattenStringValueSet(v.ExcludedColumnNames)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

// This only happens in very specific situations:
// (Select) TWC + ColumnWildcard              = (Select) Table
// (Select) TWC + ColumnWildcard + ALL_TABLES = (Select) Table + TableWildcard
func flattenTableColumnsResourceAsTable(apiObject *awstypes.TableWithColumnsResource) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CatalogId; v != nil {
		tfMap[names.AttrCatalogID] = aws.ToString(v)
	}

	if v := apiObject.DatabaseName; v != nil {
		tfMap[names.AttrDatabaseName] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil && aws.ToString(v) == TableNameAllTables && apiObject.ColumnWildcard != nil {
		tfMap["wildcard"] = true
	} else if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

func flattenResourcePermissions(apiObjects []awstypes.PrincipalResourcePermissions) []string {
	if apiObjects == nil {
		return nil
	}

	tfList := make([]string, 0)

	for _, resourcePermission := range apiObjects {
		for _, permission := range resourcePermission.Permissions {
			tfList = append(tfList, string(permission))
		}
	}

	slices.Sort(tfList)

	return tfList
}

func flattenGrantPermissions(apiObjects []awstypes.PrincipalResourcePermissions) []string {
	if apiObjects == nil {
		return nil
	}

	tfList := make([]string, 0)

	for _, resourcePermission := range apiObjects {
		for _, grantPermission := range resourcePermission.PermissionsWithGrantOption {
			tfList = append(tfList, string(grantPermission))
		}
	}

	slices.Sort(tfList)

	return tfList
}

func includePrincipalIdentifierInList(principalIdentifier string) bool {
	arn, err := arn.Parse(principalIdentifier)
	if err != nil {
		return true
	}
	return !(arn.Service == "identitystore" && strings.HasPrefix(arn.Resource, "group/"))
}
