// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
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

	d.SetId(strconv.Itoa(create.StringHashcode(prettify(input))))

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
		if err := d.Set("data_location", []any{flattenDataLocationResource(permissions[0].Resource.DataLocation)}); err != nil { // nosemgrep:ci.data-source-with-resource-read
			return sdkdiag.AppendErrorf(diags, "setting data_location: %s", err)
		}
	} else {
		d.Set("data_location", nil)
	}

	if permissions[0].Resource.DataCellsFilter != nil {
		if err := d.Set("data_cells_filter", flattenDataCellsFilter(permissions[0].Resource.DataCellsFilter)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting data_cells_filter: %s", err)
		}
	} else {
		d.Set("data_cells_filter", nil)
	}

	if permissions[0].Resource.Database != nil {
		if err := d.Set(names.AttrDatabase, []any{flattenDatabaseResource(permissions[0].Resource.Database)}); err != nil { // nosemgrep:ci.data-source-with-resource-read
			return sdkdiag.AppendErrorf(diags, "setting database: %s", err)
		}
	} else {
		d.Set(names.AttrDatabase, nil)
	}

	if permissions[0].Resource.LFTag != nil {
		if err := d.Set("lf_tag", []any{flattenLFTagKeyResource(permissions[0].Resource.LFTag)}); err != nil { // nosemgrep:ci.data-source-with-resource-read
			return sdkdiag.AppendErrorf(diags, "setting LF-tag: %s", err)
		}
	} else {
		d.Set("lf_tag", nil)
	}

	if permissions[0].Resource.LFTagPolicy != nil {
		if err := d.Set("lf_tag_policy", []any{flattenLFTagPolicyResource(permissions[0].Resource.LFTagPolicy)}); err != nil { // nosemgrep:ci.data-source-with-resource-read
			return sdkdiag.AppendErrorf(diags, "setting LF-tag policy: %s", err)
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
		for _, perm := range permissions {
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
