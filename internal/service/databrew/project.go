// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package databrew

// // @SDKResource("aws_databrew_project")
// func ResourceDataBrewProject() *schema.Resource {
// 	return &schema.Resource{
// 		CreateWithoutTimeout: resourceDataBrewProjectCreate,
// 		UpdateWithoutTimeout: resourceDataBrewProjectUpdate,
// 		ReadWithoutTimeout:   resourceDataBrewProjectRead,
// 		DeleteWithoutTimeout: resourceDataBrewProjectDelete,

// 		Importer: &schema.ResourceImporter{
// 			StateContext: schema.ImportStatePassthroughContext,
// 		},

// 		Schema: map[string]*schema.Schema{
// 			names.AttrName: {
// 				Type:         schema.TypeString,
// 				ForceNew:     true,
// 				Optional:     true,
// 				Required:     true,
// 				ValidateFunc: validation.StringLenBetween(1, 255),
// 			},
// 			"DatasetName": {
// 				Type:         schema.TypeString,
// 				Required:     true,
// 				ValidateFunc: validation.StringLenBetween(1, 255),
// 			},
// 			"RecipeName": {
// 				Type:         schema.TypeString,
// 				Required:     true,
// 				ValidateFunc: validation.StringLenBetween(1, 255),
// 			},
// 			names.AttrRoleARN: {
// 				Type:         schema.TypeString,
// 				Required:     true,
// 				ValidateFunc: verify.ValidARN,
// 			},
// 			"Sample": {
// 				Type:     schema.TypeInt,
// 				Optional: true,
// 			},
// 			names.AttrTags: tftags.TagsSchema(),
// 		},
// 	}
// }

// func resourceDataBrewProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	var diags diag.Diagnostics
// 	conn := meta.(*conns.AWSClient).(ctx)

// 	input := &lakeformation.PutDataLakeSettingsInput{}

// 	if v, ok := d.GetOk(names.AttrCatalogID); ok {
// 		input.CatalogId = aws.String(v.(string))
// 	}

// 	settings := &awstypes.DataLakeSettings{}

// 	if v, ok := d.GetOk("admins"); ok {
// 		settings.DataLakeAdmins = expandDataLakeSettingsAdmins(v.(*schema.Set))
// 	}

// 	if v, ok := d.GetOk("read_only_admins"); ok {
// 		settings.ReadOnlyAdmins = expandDataLakeSettingsAdmins(v.(*schema.Set))
// 	}

// 	if v, ok := d.GetOk("allow_external_data_filtering"); ok {
// 		settings.AllowExternalDataFiltering = aws.Bool(v.(bool))
// 	}

// 	if v, ok := d.GetOk("authorized_session_tag_value_list"); ok {
// 		settings.AuthorizedSessionTagValueList = flex.ExpandStringValueList(v.([]interface{}))
// 	}

// 	if v, ok := d.GetOk("create_database_default_permissions"); ok {
// 		settings.CreateDatabaseDefaultPermissions = expandDataLakeSettingsCreateDefaultPermissions(v.([]interface{}))
// 	}

// 	if v, ok := d.GetOk("create_table_default_permissions"); ok {
// 		settings.CreateTableDefaultPermissions = expandDataLakeSettingsCreateDefaultPermissions(v.([]interface{}))
// 	}

// 	if v, ok := d.GetOk("external_data_filtering_allow_list"); ok {
// 		settings.ExternalDataFilteringAllowList = expandDataLakeSettingsDataFilteringAllowList(v.(*schema.Set))
// 	}

// 	if v, ok := d.GetOk("trusted_resource_owners"); ok {
// 		settings.TrustedResourceOwners = flex.ExpandStringValueList(v.([]interface{}))
// 	}

// 	input.DataLakeSettings = settings

// 	var output *lakeformation.PutDataLakeSettingsOutput
// 	err := retry.RetryContext(ctx, IAMPropagationTimeout, func() *retry.RetryError {
// 		var err error
// 		output, err = conn.PutDataLakeSettings(ctx, input)
// 		if err != nil {
// 			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Invalid principal") {
// 				return retry.RetryableError(err)
// 			}

// 			if errs.IsA[*awstypes.ConcurrentModificationException](err) {
// 				return retry.RetryableError(err)
// 			}

// 			return retry.NonRetryableError(fmt.Errorf("creating Lake Formation data lake settings: %w", err))
// 		}
// 		return nil
// 	})

// 	if tfresource.TimedOut(err) {
// 		output, err = conn.PutDataLakeSettings(ctx, input)
// 	}

// 	if err != nil {
// 		return sdkdiag.AppendErrorf(diags, "creating Lake Formation data lake settings: %s", err)
// 	}

// 	if output == nil {
// 		return sdkdiag.AppendErrorf(diags, "creating Lake Formation data lake settings: empty response")
// 	}

// 	d.SetId(fmt.Sprintf("%d", create.StringHashcode(prettify(input))))

// 	return append(diags, resourceDataLakeSettingsRead(ctx, d, meta)...)
// }

// func resourceDataLakeSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	var diags diag.Diagnostics
// 	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

// 	input := &lakeformation.GetDataLakeSettingsInput{}

// 	if v, ok := d.GetOk(names.AttrCatalogID); ok {
// 		input.CatalogId = aws.String(v.(string))
// 	}

// 	output, err := conn.GetDataLakeSettings(ctx, input)

// 	if !d.IsNewResource() && errs.IsA[*awstypes.EntityNotFoundException](err) {
// 		log.Printf("[WARN] Lake Formation data lake settings (%s) not found, removing from state", d.Id())
// 		d.SetId("")
// 		return diags
// 	}

// 	if err != nil {
// 		return sdkdiag.AppendErrorf(diags, "reading Lake Formation data lake settings (%s): %s", d.Id(), err)
// 	}

// 	if output == nil || output.DataLakeSettings == nil {
// 		return sdkdiag.AppendErrorf(diags, "reading Lake Formation data lake settings (%s): empty response", d.Id())
// 	}

// 	settings := output.DataLakeSettings

// 	d.Set("admins", flattenDataLakeSettingsAdmins(settings.DataLakeAdmins))
// 	d.Set("read_only_admins", flattenDataLakeSettingsAdmins(settings.ReadOnlyAdmins))
// 	d.Set("allow_external_data_filtering", settings.AllowExternalDataFiltering)
// 	d.Set("authorized_session_tag_value_list", flex.FlattenStringValueList(settings.AuthorizedSessionTagValueList))
// 	d.Set("create_database_default_permissions", flattenDataLakeSettingsCreateDefaultPermissions(settings.CreateDatabaseDefaultPermissions))
// 	d.Set("create_table_default_permissions", flattenDataLakeSettingsCreateDefaultPermissions(settings.CreateTableDefaultPermissions))
// 	d.Set("external_data_filtering_allow_list", flattenDataLakeSettingsDataFilteringAllowList(settings.ExternalDataFilteringAllowList))
// 	d.Set("trusted_resource_owners", flex.FlattenStringValueList(settings.TrustedResourceOwners))

// 	return diags
// }

// func resourceDataLakeSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	var diags diag.Diagnostics
// 	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

// 	input := &lakeformation.PutDataLakeSettingsInput{
// 		DataLakeSettings: &awstypes.DataLakeSettings{
// 			CreateDatabaseDefaultPermissions: make([]awstypes.PrincipalPermissions, 0),
// 			CreateTableDefaultPermissions:    make([]awstypes.PrincipalPermissions, 0),
// 			DataLakeAdmins:                   make([]awstypes.DataLakePrincipal, 0),
// 			ReadOnlyAdmins:                   make([]awstypes.DataLakePrincipal, 0),
// 			TrustedResourceOwners:            make([]string, 0),
// 		},
// 	}

// 	if v, ok := d.GetOk(names.AttrCatalogID); ok {
// 		input.CatalogId = aws.String(v.(string))
// 	}

// 	_, err := conn.PutDataLakeSettings(ctx, input)

// 	if errs.IsA[*awstypes.EntityNotFoundException](err) {
// 		log.Printf("[WARN] Lake Formation data lake settings (%s) not found, removing from state", d.Id())
// 		return diags
// 	}

// 	if err != nil {
// 		return sdkdiag.AppendErrorf(diags, "deleting Lake Formation data lake settings (%s): %s", d.Id(), err)
// 	}

// 	return diags
// }

// func expandDataLakeSettingsCreateDefaultPermissions(tfMaps []interface{}) []awstypes.PrincipalPermissions {
// 	apiObjects := make([]awstypes.PrincipalPermissions, 0, len(tfMaps))

// 	for _, tfMap := range tfMaps {
// 		apiObjects = append(apiObjects, expandDataLakeSettingsCreateDefaultPermission(tfMap.(map[string]interface{})))
// 	}

// 	return apiObjects
// }

// func expandDataLakeSettingsCreateDefaultPermission(tfMap map[string]interface{}) awstypes.PrincipalPermissions {
// 	apiObject := awstypes.PrincipalPermissions{
// 		Permissions: flex.ExpandStringyValueList[awstypes.Permission](tfMap[names.AttrPermissions].(*schema.Set).List()),
// 		Principal: &awstypes.DataLakePrincipal{
// 			DataLakePrincipalIdentifier: aws.String(tfMap[names.AttrPrincipal].(string)),
// 		},
// 	}

// 	return apiObject
// }

// func flattenDataLakeSettingsCreateDefaultPermissions(apiObjects []awstypes.PrincipalPermissions) []map[string]interface{} {
// 	if apiObjects == nil {
// 		return nil
// 	}

// 	tfMaps := make([]map[string]interface{}, len(apiObjects))
// 	for i, v := range apiObjects {
// 		tfMaps[i] = flattenDataLakeSettingsCreateDefaultPermission(v)
// 	}

// 	return tfMaps
// }

// func flattenDataLakeSettingsCreateDefaultPermission(apiObject awstypes.PrincipalPermissions) map[string]interface{} {
// 	tfMap := make(map[string]interface{})

// 	if reflect.ValueOf(apiObject).IsZero() {
// 		return tfMap
// 	}

// 	if apiObject.Permissions != nil {
// 		// tfMap["permissions"] = flex.FlattenStringValueSet(flattenPermissions(apiObject.Permissions))
// 		tfMap[names.AttrPermissions] = flex.FlattenStringyValueSet(apiObject.Permissions)
// 	}

// 	if v := aws.ToString(apiObject.Principal.DataLakePrincipalIdentifier); v != "" {
// 		tfMap[names.AttrPrincipal] = v
// 	}

// 	return tfMap
// }

// func expandDataLakeSettingsAdmins(tfSet *schema.Set) []awstypes.DataLakePrincipal {
// 	tfSlice := tfSet.List()
// 	apiObjects := make([]awstypes.DataLakePrincipal, 0, len(tfSlice))

// 	for _, tfItem := range tfSlice {
// 		val, ok := tfItem.(string)
// 		if ok && val != "" {
// 			apiObjects = append(apiObjects, awstypes.DataLakePrincipal{
// 				DataLakePrincipalIdentifier: aws.String(tfItem.(string)),
// 			})
// 		}
// 	}

// 	return apiObjects
// }

// func flattenDataLakeSettingsAdmins(apiObjects []awstypes.DataLakePrincipal) []interface{} {
// 	if apiObjects == nil {
// 		return nil
// 	}

// 	tfSlice := make([]interface{}, 0, len(apiObjects))

// 	for _, apiObject := range apiObjects {
// 		tfSlice = append(tfSlice, aws.ToString(apiObject.DataLakePrincipalIdentifier))
// 	}

// 	return tfSlice
// }

// func expandDataLakeSettingsDataFilteringAllowList(tfSet *schema.Set) []awstypes.DataLakePrincipal {
// 	tfSlice := tfSet.List()
// 	apiObjects := make([]awstypes.DataLakePrincipal, 0, len(tfSlice))

// 	for _, tfItem := range tfSlice {
// 		val, ok := tfItem.(string)
// 		if ok && val != "" {
// 			apiObjects = append(apiObjects, awstypes.DataLakePrincipal{
// 				DataLakePrincipalIdentifier: aws.String(tfItem.(string)),
// 			})
// 		}
// 	}

// 	return apiObjects
// }

// func flattenDataLakeSettingsDataFilteringAllowList(apiObjects []awstypes.DataLakePrincipal) []interface{} {
// 	if apiObjects == nil {
// 		return nil
// 	}

// 	tfSlice := make([]interface{}, 0, len(apiObjects))

// 	for _, apiObject := range apiObjects {
// 		tfSlice = append(tfSlice, aws.ToString(apiObject.DataLakePrincipalIdentifier))
// 	}

// 	return tfSlice
// }
