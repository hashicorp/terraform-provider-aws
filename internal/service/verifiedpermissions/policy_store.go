// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

//import (
//	"context"
//	"errors"
//	"log"
//	"time"
//
//	"github.com/aws/aws-sdk-go-v2/aws"
//	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
//	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
//	"github.com/hashicorp/terraform-provider-aws/internal/conns"
//	"github.com/hashicorp/terraform-provider-aws/internal/create"
//	"github.com/hashicorp/terraform-provider-aws/internal/enum"
//	"github.com/hashicorp/terraform-provider-aws/internal/errs"
//	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
//	"github.com/hashicorp/terraform-provider-aws/internal/verify"
//	"github.com/hashicorp/terraform-provider-aws/names"
//)
//
//// @SDKResource("aws_verifiedpermissions_policy_store", name="Policy Store")
//func ResourcePolicyStore() *schema.Resource {
//	return &schema.Resource{
//		CreateWithoutTimeout: resourcePolicyStoreCreate,
//		ReadWithoutTimeout:   resourcePolicyStoreRead,
//		UpdateWithoutTimeout: resourcePolicyStoreUpdate,
//		DeleteWithoutTimeout: resourcePolicyStoreDelete,
//		Timeouts: &schema.ResourceTimeout{
//			Create: schema.DefaultTimeout(30 * time.Minute),
//			Update: schema.DefaultTimeout(30 * time.Minute),
//			Delete: schema.DefaultTimeout(30 * time.Minute),
//		},
//		Importer: &schema.ResourceImporter{
//			StateContext: schema.ImportStatePassthroughContext,
//		},
//		Schema: map[string]*schema.Schema{
//			"policy_store_id": {
//				Type:     schema.TypeString,
//				Computed: true,
//			},
//			"arn": {
//				Type:     schema.TypeString,
//				Computed: true,
//			},
//			"created_date": {
//				Type:     schema.TypeString,
//				Computed: true,
//			},
//			"last_updated_date": {
//				Type:     schema.TypeString,
//				Computed: true,
//			},
//			"validation_settings": {
//				Type:     schema.TypeList,
//				Required: true,
//				MaxItems: 1,
//				Elem: &schema.Resource{
//					Schema: map[string]*schema.Schema{
//						"mode": {
//							Type:             schema.TypeString,
//							ValidateDiagFunc: enum.Validate[types.ValidationMode](),
//							Required:         true,
//						},
//					},
//				},
//			},
//			"schema": {
//				Type:     schema.TypeList,
//				Optional: true,
//				MaxItems: 1,
//				Elem: &schema.Resource{
//					Schema: map[string]*schema.Schema{
//						"cedar_json": {
//							Type:         schema.TypeString,
//							Required:     true,
//							ValidateFunc: validation.StringIsJSON,
//							StateFunc: func(v interface{}) string {
//								json, _ := structure.NormalizeJsonString(v)
//								return json
//							},
//							DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
//						},
//					},
//				},
//			},
//			"schema_created_date": {
//				Type:     schema.TypeString,
//				Computed: true,
//			},
//			"schema_last_updated_date": {
//				Type:     schema.TypeString,
//				Computed: true,
//			},
//		},
//	}
//}
//
//func resourcePolicyStoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//
//	conn := meta.(*conns.AWSClient).VerifiedPermissionsClient(ctx)
//
//	in := &verifiedpermissions.CreatePolicyStoreInput{
//		ClientToken:        aws.String(id.UniqueId()),
//		ValidationSettings: expandValidationSettingsSlice(d.Get("validation_settings").([]interface{})),
//	}
//
//	out, err := conn.CreatePolicyStore(ctx, in)
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicyStore, "", err)...)
//	}
//
//	if out == nil || out.PolicyStoreId == nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicyStore, "", errors.New("empty output"))...)
//	}
//
//	policyStoreId := aws.ToString(out.PolicyStoreId)
//	d.SetId(policyStoreId)
//
//	cedarJson := expandDefinitions(d.Get("schema").([]interface{}))
//	inSchema := &verifiedpermissions.PutSchemaInput{
//		Definition:    cedarJson,
//		PolicyStoreId: &policyStoreId,
//	}
//
//	_, err = conn.PutSchema(ctx, inSchema)
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicyStoreSchema, "", err)...)
//	}
//
//	return append(diags, resourcePolicyStoreRead(ctx, d, meta)...)
//}
//
//func resourcePolicyStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//
//	conn := meta.(*conns.AWSClient).VerifiedPermissionsClient(ctx)
//
//	out, err := findPolicyStoreByID(ctx, conn, d.Id())
//	if !d.IsNewResource() && tfresource.NotFound(err) {
//		log.Printf("[WARN] VerifiedPermissions PolicyStore (%s) not found, removing from state", d.Id())
//		d.SetId("")
//		return diags
//	}
//
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionReading, ResNamePolicyStore, d.Id(), err)...)
//	}
//
//	d.Set("arn", out.Arn)
//	d.Set("policy_store_id", out.PolicyStoreId)
//	d.Set("created_date", out.CreatedDate.Format(time.RFC3339))
//	d.Set("last_updated_date", out.LastUpdatedDate.Format(time.RFC3339))
//
//	if err := d.Set("validation_settings", flattenValidationSettingsSlice(out.ValidationSettings)); err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionSetting, ResNamePolicyStore, d.Id(), err)...)
//	}
//
//	inSchema := &verifiedpermissions.GetSchemaInput{
//		PolicyStoreId: out.PolicyStoreId,
//	}
//	outSchema, err := conn.GetSchema(ctx, inSchema)
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionReading, ResNamePolicyStoreSchema, *out.PolicyStoreId, err)...)
//	}
//
//	d.Set("schema", flattenSchemaDefinitionSlice(outSchema.Schema))
//	d.Set("schema_created_date", outSchema.CreatedDate.Format(time.RFC3339))
//	d.Set("schema_last_updated_date", outSchema.LastUpdatedDate.Format(time.RFC3339))
//
//	return diags
//}
//
//func resourcePolicyStoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//
//	conn := meta.(*conns.AWSClient).VerifiedPermissionsClient(ctx)
//
//	policyStoreID := aws.String(d.Id())
//	update := false
//
//	policyStoreIn := &verifiedpermissions.UpdatePolicyStoreInput{
//		PolicyStoreId: policyStoreID,
//	}
//
//	if d.HasChanges("validation_settings") {
//		policyStoreIn.ValidationSettings = expandValidationSettingsSlice(d.Get("validation_settings").([]interface{}))
//		update = true
//	}
//
//	if update {
//		log.Printf("[DEBUG] Updating VerifiedPermissions PolicyStore (%s): %#v", d.Id(), policyStoreIn)
//		_, err := conn.UpdatePolicyStore(ctx, policyStoreIn)
//		if err != nil {
//			return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionUpdating, ResNamePolicyStore, d.Id(), err)...)
//		}
//	}
//
//	var updateSchema bool
//	schemaIn := &verifiedpermissions.PutSchemaInput{
//		PolicyStoreId: policyStoreID,
//	}
//	if d.HasChanges("schema") {
//		schemaIn.Definition = expandDefinitions(d.Get("schema").([]interface{}))
//		updateSchema = true
//	}
//
//	if updateSchema {
//		log.Printf("[DEBUG] Updating VerifiedPermissions PolicyStore Schema (%s): %#v", d.Id(), policyStoreIn)
//		_, err := conn.PutSchema(ctx, schemaIn)
//		if err != nil {
//			return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionUpdating, ResNamePolicyStoreSchema, d.Id(), err)...)
//		}
//	}
//
//	return append(diags, resourcePolicyStoreRead(ctx, d, meta)...)
//}
//
//func resourcePolicyStoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//
//	conn := meta.(*conns.AWSClient).VerifiedPermissionsClient(ctx)
//
//	log.Printf("[INFO] Deleting VerifiedPermissions PolicyStore %s", d.Id())
//
//	_, err := conn.DeletePolicyStore(ctx, &verifiedpermissions.DeletePolicyStoreInput{
//		PolicyStoreId: aws.String(d.Id()),
//	})
//
//	if errs.IsA[*types.ResourceNotFoundException](err) {
//		return diags
//	}
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionDeleting, ResNamePolicyStore, d.Id(), err)...)
//	}
//
//	return diags
//}
//
//func flattenValidationSettings(apiObject *types.ValidationSettings) map[string]interface{} {
//	if apiObject == nil {
//		return nil
//	}
//
//	m := make(map[string]interface{})
//
//	if v := apiObject.Mode; v != "" {
//		m["mode"] = string(v)
//	}
//
//	return m
//}
//
//func flattenValidationSettingsSlice(apiObject *types.ValidationSettings) []interface{} {
//	if apiObject == nil {
//		return nil
//	}
//
//	return []interface{}{
//		flattenValidationSettings(apiObject),
//	}
//}
//
//func expandValidationSettingsSlice(list []interface{}) *types.ValidationSettings {
//	if len(list) == 0 {
//		return nil
//	}
//	tfMap := list[0].(map[string]interface{})
//
//	return expandValidationSettings(tfMap)
//}
//
//func expandValidationSettings(tfMap map[string]interface{}) *types.ValidationSettings {
//	if tfMap == nil {
//		return nil
//	}
//
//	var out types.ValidationSettings
//
//	mode, ok := tfMap["mode"].(string)
//	if ok {
//		out.Mode = types.ValidationMode(mode)
//	}
//
//	return &out
//}
//
//func expandDefinition(tfMap map[string]interface{}) *types.SchemaDefinitionMemberCedarJson {
//	a := &types.SchemaDefinitionMemberCedarJson{
//		Value: "{}",
//	}
//
//	if v, ok := tfMap["cedar_json"].(string); ok && v != "" {
//		var err error
//		a.Value, err = structure.NormalizeJsonString(v)
//		if err != nil {
//			return a
//		}
//	}
//
//	return a
//}
//
//func expandDefinitions(tfList []interface{}) *types.SchemaDefinitionMemberCedarJson {
//	if len(tfList) == 0 {
//		return &types.SchemaDefinitionMemberCedarJson{
//			Value: "{}",
//		}
//	}
//
//	tfMap := tfList[0]
//	if tfMap == nil {
//		return &types.SchemaDefinitionMemberCedarJson{
//			Value: "{}",
//		}
//	}
//
//	return expandDefinition(tfMap.(map[string]interface{}))
//}
//
//func flattenSchemaDefinitionSlice(definition *string) []interface{} {
//	def := flattenSchemaDefinition(definition)
//	if def == nil {
//		return nil
//	}
//
//	return []interface{}{def}
//}
//
//func flattenSchemaDefinition(definition *string) map[string]interface{} {
//	if definition == nil || aws.ToString(definition) == "{}" {
//		return nil
//	}
//
//	specificationToSet, err := structure.NormalizeJsonString(aws.ToString(definition))
//	if err != nil {
//		return nil
//	}
//
//	return map[string]interface{}{
//		"cedar_json": specificationToSet,
//	}
//}
