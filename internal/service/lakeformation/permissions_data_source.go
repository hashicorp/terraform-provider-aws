// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lakeformation_permissions", name="Permissions")
func DataSourcePermissions() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePermissionsRead,

		Schema: map[string]*schema.Schema{
			names.AttrCatalogID: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"catalog_resource": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"data_cells_filter": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
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
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrCatalogID: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidAccountID,
						},
					},
				},
			},
			names.AttrDatabase: {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidAccountID,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"lf_tag": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						names.AttrKey: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						names.AttrValues: {
							Type:     schema.TypeSet,
							Required: true,
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
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidAccountID,
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
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									names.AttrValues: {
										Type:     schema.TypeSet,
										Required: true,
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
							ValidateDiagFunc: enum.Validate[awstypes.ResourceType](),
						},
					},
				},
			},
			names.AttrPermissions: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"permissions_with_grant_option": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrPrincipal: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validPrincipal,
			},
			"table": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidAccountID,
						},
						names.AttrDatabaseName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"wildcard": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"table_with_columns": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidAccountID,
						},
						"column_names": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
						},
						names.AttrDatabaseName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"excluded_column_names": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"wildcard": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePermissionsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	input := &lakeformation.ListPermissionsInput{
		Principal: &awstypes.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(d.Get(names.AttrPrincipal).(string)),
		},
		Resource: &awstypes.Resource{},
	}

	if v, ok := d.GetOk(names.AttrCatalogID); ok {
		input.CatalogId = aws.String(v.(string))
	}

	if _, ok := d.GetOk("catalog_resource"); ok {
		input.Resource.Catalog = ExpandCatalogResource()
	}

	if v, ok := d.GetOk("data_cells_filter"); ok {
		input.Resource.DataCellsFilter = ExpandDataCellsFilter(v.([]any))
	}

	if v, ok := d.GetOk("data_location"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Resource.DataLocation = ExpandDataLocationResource(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrDatabase); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Resource.Database = ExpandDatabaseResource(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("lf_tag"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Resource.LFTag = ExpandLFTagKeyResource(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("lf_tag_policy"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Resource.LFTagPolicy = ExpandLFTagPolicyResource(v.([]any)[0].(map[string]any))
	}

	tableType := ""

	if v, ok := d.GetOk("table"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Resource.Table = ExpandTableResource(v.([]any)[0].(map[string]any))
		tableType = TableTypeTable
	}

	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		// can't ListPermissions for TableWithColumns, so use Table instead
		input.Resource.Table = ExpandTableWithColumnsResourceAsTable(v.([]any)[0].(map[string]any))
		tableType = TableTypeTableWithColumns
	}

	columnNames := make([]string, 0)
	excludedColumnNames := make([]string, 0)
	columnWildcard := false

	if tableType == TableTypeTableWithColumns {
		if v, ok := d.GetOk("table_with_columns.0.wildcard"); ok {
			columnWildcard = v.(bool)
		}

		if v, ok := d.GetOk("table_with_columns.0.column_names"); ok {
			if v, ok := v.(*schema.Set); ok && v.Len() > 0 {
				columnNames = flex.ExpandStringValueSet(v)
			}
		}

		if v, ok := d.GetOk("table_with_columns.0.excluded_column_names"); ok {
			if v, ok := v.(*schema.Set); ok && v.Len() > 0 {
				excludedColumnNames = flex.ExpandStringValueSet(v)
			}
		}
	}

	log.Printf("[DEBUG] Reading Lake Formation permissions: %v", input)

	allPermissions, err := waitPermissionsReady(ctx, conn, input, tableType, columnNames, excludedColumnNames, columnWildcard)

	d.SetId(strconv.Itoa(create.StringHashcode(prettify(input))))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lake Formation permissions: %s", err)
	}

	// clean permissions = filter out permissions that do not pertain to this specific resource
	cleanPermissions := FilterPermissions(input, tableType, columnNames, excludedColumnNames, columnWildcard, allPermissions)

	if len(cleanPermissions) != len(allPermissions) {
		log.Printf("[INFO] Resource Lake Formation clean permissions (%d) and all permissions (%d) have different lengths (this is not necessarily a problem): %s", len(cleanPermissions), len(allPermissions), d.Id())
	}

	d.Set(names.AttrPrincipal, cleanPermissions[0].Principal.DataLakePrincipalIdentifier)
	d.Set(names.AttrPermissions, flattenResourcePermissions(cleanPermissions))
	d.Set("permissions_with_grant_option", flattenGrantPermissions(cleanPermissions))

	if cleanPermissions[0].Resource.Catalog != nil {
		d.Set("catalog_resource", true)
	} else {
		d.Set("catalog_resource", false)
	}

	if cleanPermissions[0].Resource.DataLocation != nil {
		if err := d.Set("data_location", []any{flattenDataLocationResource(cleanPermissions[0].Resource.DataLocation)}); err != nil { // nosemgrep:ci.data-source-with-resource-read
			return sdkdiag.AppendErrorf(diags, "setting data_location: %s", err)
		}
	} else {
		d.Set("data_location", nil)
	}

	if cleanPermissions[0].Resource.DataCellsFilter != nil {
		if err := d.Set("data_cells_filter", flattenDataCellsFilter(cleanPermissions[0].Resource.DataCellsFilter)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting data_cells_filter: %s", err)
		}
	} else {
		d.Set("data_cells_filter", nil)
	}

	if cleanPermissions[0].Resource.Database != nil {
		if err := d.Set(names.AttrDatabase, []any{flattenDatabaseResource(cleanPermissions[0].Resource.Database)}); err != nil { // nosemgrep:ci.data-source-with-resource-read
			return sdkdiag.AppendErrorf(diags, "setting database: %s", err)
		}
	} else {
		d.Set(names.AttrDatabase, nil)
	}

	if cleanPermissions[0].Resource.LFTag != nil {
		if err := d.Set("lf_tag", []any{flattenLFTagKeyResource(cleanPermissions[0].Resource.LFTag)}); err != nil { // nosemgrep:ci.data-source-with-resource-read
			return sdkdiag.AppendErrorf(diags, "setting LF-tag: %s", err)
		}
	} else {
		d.Set("lf_tag", nil)
	}

	if cleanPermissions[0].Resource.LFTagPolicy != nil {
		if err := d.Set("lf_tag_policy", []any{flattenLFTagPolicyResource(cleanPermissions[0].Resource.LFTagPolicy)}); err != nil { // nosemgrep:ci.data-source-with-resource-read
			return sdkdiag.AppendErrorf(diags, "setting LF-tag policy: %s", err)
		}
	} else {
		d.Set("lf_tag_policy", nil)
	}

	tableSet := false

	if v, ok := d.GetOk("table"); ok && len(v.([]any)) > 0 {
		// since perm list could include TableWithColumns, get the right one
		for _, perm := range cleanPermissions {
			if perm.Resource == nil {
				continue
			}

			if perm.Resource.TableWithColumns != nil && perm.Resource.TableWithColumns.ColumnWildcard != nil {
				if err := d.Set("table", []any{flattenTableColumnsResourceAsTable(perm.Resource.TableWithColumns)}); err != nil { // nosemgrep:ci.data-source-with-resource-read
					return sdkdiag.AppendErrorf(diags, "setting table: %s", err)
				}
				tableSet = true
				break
			}

			if perm.Resource.Table != nil {
				if err := d.Set("table", []any{flattenTableResource(perm.Resource.Table)}); err != nil { // nosemgrep:ci.data-source-with-resource-read
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
		for _, perm := range cleanPermissions {
			if perm.Resource.TableWithColumns != nil {
				if err := d.Set("table_with_columns", []any{flattenTableColumnsResource(perm.Resource.TableWithColumns)}); err != nil { // nosemgrep:ci.data-source-with-resource-read
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
