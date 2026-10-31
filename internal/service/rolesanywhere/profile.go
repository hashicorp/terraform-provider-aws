// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package rolesanywhere

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rolesanywhere/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rolesanywhere_profile", name="Profile")
// @Tags(identifierAttribute="arn")
func resourceProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProfileCreate,
		ReadWithoutTimeout:   resourceProfileRead,
		UpdateWithoutTimeout: resourceProfileUpdate,
		DeleteWithoutTimeout: resourceProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"accept_role_session_name": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"attribute_mappings": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"certificate_field": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.CertificateField](),
							},
							"mapping_rules": {
								Type:     schema.TypeSet,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"specifier": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
						},
					},
				},
				"duration_seconds": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				names.AttrEnabled: {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"managed_policy_arns": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: verify.ValidARN,
					},
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				"require_instance_properties": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"role_arns": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: verify.ValidARN,
					},
				},
				"session_policy": {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},
	}
}

func resourceProfileCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := rolesanywhere.CreateProfileInput{
		Name:     aws.String(name),
		RoleArns: flex.ExpandStringValueSet(d.Get("role_arns").(*schema.Set)), // Send [] if not configured.
		Tags:     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("accept_role_session_name"); ok {
		input.AcceptRoleSessionName = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("duration_seconds"); ok {
		input.DurationSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrEnabled); ok {
		input.Enabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("managed_policy_arns"); ok {
		input.ManagedPolicyArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("require_instance_properties"); ok {
		input.RequireInstanceProperties = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("session_policy"); ok {
		input.SessionPolicy = aws.String(v.(string))
	}

	output, err := conn.CreateProfile(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RolesAnywhere Profile (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Profile.ProfileId))

	if v, ok := d.GetOk("attribute_mappings"); ok {
		if err := putProfileAttributeMappings(ctx, conn, d.Id(), v.(*schema.Set).List()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting RolesAnywhere Profile (%s) attribute mappings: %s", d.Id(), err)
		}
	}

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	profile, err := findProfileByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] RolesAnywhere Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RolesAnywhere Profile (%s): %s", d.Id(), err)
	}

	d.Set("accept_role_session_name", profile.AcceptRoleSessionName)
	d.Set(names.AttrARN, profile.ProfileArn)
	if err := d.Set("attribute_mappings", flattenAttributeMappings(profile.AttributeMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attribute_mappings: %s", err)
	}
	d.Set("duration_seconds", profile.DurationSeconds)
	d.Set(names.AttrEnabled, profile.Enabled)
	d.Set("managed_policy_arns", profile.ManagedPolicyArns)
	d.Set(names.AttrName, profile.Name)
	d.Set("require_instance_properties", profile.RequireInstanceProperties)
	d.Set("role_arns", profile.RoleArns)
	d.Set("session_policy", profile.SessionPolicy)

	return diags
}

func resourceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "attribute_mappings") {
		input := rolesanywhere.UpdateProfileInput{
			ProfileId: aws.String(d.Id()),
		}

		if d.HasChange("accept_role_session_name") {
			input.AcceptRoleSessionName = aws.Bool(d.Get("accept_role_session_name").(bool))
		}

		if d.HasChange("duration_seconds") {
			input.DurationSeconds = aws.Int32(int32(d.Get("duration_seconds").(int)))
		}

		if d.HasChange("managed_policy_arns") {
			input.ManagedPolicyArns = flex.ExpandStringValueSet(d.Get("managed_policy_arns").(*schema.Set))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("role_arns") {
			input.RoleArns = flex.ExpandStringValueSet(d.Get("role_arns").(*schema.Set))
		}

		if d.HasChange("session_policy") {
			input.SessionPolicy = aws.String(d.Get("session_policy").(string))
		}

		_, err := conn.UpdateProfile(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RolesAnywhere Profile (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("attribute_mappings") {
		o, n := d.GetChange("attribute_mappings")
		oldCertificateFields := attributeMappingCertificateFields(o.(*schema.Set).List())
		newMappings := n.(*schema.Set).List()
		newCertificateFields := attributeMappingCertificateFields(newMappings)

		// Delete mappings for certificate fields that are no longer configured;
		// PutAttributeMapping only replaces a field, it does not remove one.
		for certificateField := range oldCertificateFields {
			if _, ok := newCertificateFields[certificateField]; !ok {
				if err := deleteProfileAttributeMapping(ctx, conn, d.Id(), awstypes.CertificateField(certificateField)); err != nil {
					return sdkdiag.AppendErrorf(diags, "deleting RolesAnywhere Profile (%s) attribute mapping (%s): %s", d.Id(), certificateField, err)
				}
			}
		}

		if err := putProfileAttributeMappings(ctx, conn, d.Id(), newMappings); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RolesAnywhere Profile (%s) attribute mappings: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrEnabled) {
		if _, n := d.GetChange(names.AttrEnabled); n == true {
			err := enableProfile(ctx, conn, d.Id())
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling RolesAnywhere Profile (%s): %s", d.Id(), err)
			}
		} else {
			err := disableProfile(ctx, conn, d.Id())
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling RolesAnywhere Profile (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	log.Printf("[DEBUG] Deleting RolesAnywhere Profile (%s)", d.Id())
	input := rolesanywhere.DeleteProfileInput{
		ProfileId: aws.String(d.Id()),
	}
	_, err := conn.DeleteProfile(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RolesAnywhere Profile: (%s): %s", d.Id(), err)
	}

	return diags
}

func findProfileByID(ctx context.Context, conn *rolesanywhere.Client, id string) (*awstypes.ProfileDetail, error) {
	input := rolesanywhere.GetProfileInput{
		ProfileId: aws.String(id),
	}

	return findProfile(ctx, conn, &input)
}

func findProfile(ctx context.Context, conn *rolesanywhere.Client, input *rolesanywhere.GetProfileInput) (*awstypes.ProfileDetail, error) {
	output, err := conn.GetProfile(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Profile == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Profile, nil
}

func disableProfile(ctx context.Context, conn *rolesanywhere.Client, id string) error {
	input := rolesanywhere.DisableProfileInput{
		ProfileId: aws.String(id),
	}

	_, err := conn.DisableProfile(ctx, &input)
	return err
}

func enableProfile(ctx context.Context, conn *rolesanywhere.Client, id string) error {
	input := rolesanywhere.EnableProfileInput{
		ProfileId: aws.String(id),
	}

	_, err := conn.EnableProfile(ctx, &input)
	return err
}

func putProfileAttributeMappings(ctx context.Context, conn *rolesanywhere.Client, profileID string, tfList []any) error {
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		input := rolesanywhere.PutAttributeMappingInput{
			CertificateField: awstypes.CertificateField(tfMap["certificate_field"].(string)),
			MappingRules:     expandMappingRules(tfMap["mapping_rules"].(*schema.Set).List()),
			ProfileId:        aws.String(profileID),
		}

		if _, err := conn.PutAttributeMapping(ctx, &input); err != nil {
			return err
		}
	}

	return nil
}

func deleteProfileAttributeMapping(ctx context.Context, conn *rolesanywhere.Client, profileID string, certificateField awstypes.CertificateField) error {
	input := rolesanywhere.DeleteAttributeMappingInput{
		CertificateField: certificateField,
		ProfileId:        aws.String(profileID),
	}

	_, err := conn.DeleteAttributeMapping(ctx, &input)
	return err
}

func attributeMappingCertificateFields(tfList []any) map[string]struct{} {
	fields := make(map[string]struct{}, len(tfList))
	for _, tfMapRaw := range tfList {
		if tfMap, ok := tfMapRaw.(map[string]any); ok {
			fields[tfMap["certificate_field"].(string)] = struct{}{}
		}
	}

	return fields
}

func expandMappingRules(tfList []any) []awstypes.MappingRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.MappingRule
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, awstypes.MappingRule{
			Specifier: aws.String(tfMap["specifier"].(string)),
		})
	}

	return apiObjects
}

func flattenAttributeMappings(apiObjects []awstypes.AttributeMapping) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any
	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"certificate_field": string(apiObject.CertificateField),
			"mapping_rules":     flattenMappingRules(apiObject.MappingRules),
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenMappingRules(apiObjects []awstypes.MappingRule) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any
	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"specifier": aws.ToString(apiObject.Specifier),
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}
