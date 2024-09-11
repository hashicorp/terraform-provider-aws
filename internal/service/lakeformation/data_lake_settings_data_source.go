// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lakeformation_data_lake_settings")
func DataSourceDataLakeSettings() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDataLakeSettingsRead,

		Schema: map[string]*schema.Schema{
			"admins": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"read_only_admins": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"allow_external_data_filtering": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"allow_full_table_external_data_access": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"authorized_session_tag_value_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"create_database_default_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrPermissions: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrPrincipal: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"create_table_default_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrPermissions: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrPrincipal: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"external_data_filtering_allow_list": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"trusted_resource_owners": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDataLakeSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	input := &lakeformation.GetDataLakeSettingsInput{}

	if v, ok := d.GetOk(names.AttrCatalogID); ok {
		input.CatalogId = aws.String(v.(string))
	}
	d.SetId(fmt.Sprintf("%d", create.StringHashcode(prettify(input))))

	output, err := conn.GetDataLakeSettings(ctx, input)

	if !d.IsNewResource() && errs.IsA[*awstypes.EntityNotFoundException](err) {
		log.Printf("[WARN] Lake Formation data lake settings (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lake Formation data lake settings (%s): %s", d.Id(), err)
	}

	if output == nil || output.DataLakeSettings == nil {
		return sdkdiag.AppendErrorf(diags, "reading Lake Formation data lake settings (%s): empty response", d.Id())
	}

	settings := output.DataLakeSettings

	d.Set("admins", flattenDataLakeSettingsAdmins(settings.DataLakeAdmins))
	d.Set("read_only_admins", flattenDataLakeSettingsAdmins(settings.ReadOnlyAdmins))
	d.Set("allow_external_data_filtering", settings.AllowExternalDataFiltering)
	d.Set("authorized_session_tag_value_list", flex.FlattenStringValueList(settings.AuthorizedSessionTagValueList))
	d.Set("create_database_default_permissions", flattenDataLakeSettingsCreateDefaultPermissions(settings.CreateDatabaseDefaultPermissions))
	d.Set("create_table_default_permissions", flattenDataLakeSettingsCreateDefaultPermissions(settings.CreateTableDefaultPermissions))
	d.Set("external_data_filtering_allow_list", flattenDataLakeSettingsDataFilteringAllowList(settings.ExternalDataFilteringAllowList))
	d.Set("trusted_resource_owners", flex.FlattenStringyValueList(settings.TrustedResourceOwners))
	d.Set("allow_full_table_external_data_access", settings.AllowFullTableExternalDataAccess)

	return diags
}
