// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rolesanywhere

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rolesanywhere/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const configuredByDefault string = "rolesanywhere.amazonaws.com"

// @SDKResource("aws_rolesanywhere_trust_anchor", name="Trust Anchor")
// @Tags(identifierAttribute="arn")
func resourceTrustAnchor() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrustAnchorCreate,
		ReadWithoutTimeout:   resourceTrustAnchorRead,
		UpdateWithoutTimeout: resourceTrustAnchorUpdate,
		DeleteWithoutTimeout: resourceTrustAnchorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"notification_settings": {
				Type:     schema.TypeSet,
				Computed: true,
				ForceNew: true,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel": {
							Type:             schema.TypeString,
							Computed:         true,
							ForceNew:         true,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.NotificationChannel](),
						},
						"configured_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
							ForceNew: true,
							Optional: true,
						},
						"event": {
							Type:             schema.TypeString,
							Computed:         true,
							ForceNew:         true,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.NotificationEvent](),
						},
						"threshold": {
							Type:     schema.TypeInt,
							Computed: true,
							ForceNew: true,
							Optional: true,
						},
					},
				},
			},
			names.AttrSource: {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_data": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"acm_pca_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"x509_certificate_data": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						names.AttrSourceType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TrustAnchorType](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customizeDiffNotificationSettings,
	}
}

// The AWS API returns two default entries for notification_settings. These entries are overwritten based on if a user defines a notification setting with the same event type.
// Since notification settings cannot be updated, we need to force a resource recreation when they change, while at the same time allowing computed values.
// Because both computed and user-defined arguments need to be supported, a custom diff function is required to handle this.
// The function checks the diff, and if the difference is due to computed value change, the diff is suppressed based on the configuredBy attribute.
func customizeDiffNotificationSettings(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	oldSet, newSet := diff.GetChange("notification_settings")

	oldSetTyped, okOld := oldSet.(*schema.Set)
	newSetTyped, okNew := newSet.(*schema.Set)

	if !okOld || !okNew {
		return fmt.Errorf("unexpected type for notification_settings: oldSet: %T, newSet: %T", oldSet, newSet)
	}

	oldList := oldSetTyped.List()
	newList := newSetTyped.List()

	for _, obj1 := range oldList {
		found := false
		for _, obj2 := range newList {
			if reflect.DeepEqual(obj1, obj2) {
				found = true
				break
			}
		}
		if !found {
			if object, okNew := obj1.(map[string]any); okNew && object["configured_by"] == configuredByDefault {
				if err := diff.Clear("notification_settings"); err != nil {
					return err
				}
				break
			}
		}
	}

	return nil
}

func resourceTrustAnchorCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &rolesanywhere.CreateTrustAnchorInput{
		Enabled: aws.Bool(d.Get(names.AttrEnabled).(bool)),
		Name:    aws.String(name),
		Source:  expandSource(d.Get(names.AttrSource).([]any)),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("notification_settings"); ok && v.(*schema.Set).Len() > 0 {
		input.NotificationSettings = expandNotificationSettings(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Creating RolesAnywhere Trust Anchor (%s): %#v", d.Id(), input)
	output, err := conn.CreateTrustAnchor(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RolesAnywhere Trust Anchor (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.TrustAnchor.TrustAnchorId))

	return append(diags, resourceTrustAnchorRead(ctx, d, meta)...)
}

func resourceTrustAnchorRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	trustAnchor, err := findTrustAnchorByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RolesAnywhere Trust Anchor (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RolesAnywhere Trust Anchor (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, trustAnchor.TrustAnchorArn)
	d.Set(names.AttrEnabled, trustAnchor.Enabled)
	d.Set(names.AttrName, trustAnchor.Name)

	if err := d.Set("notification_settings", flattenNotificationSettings(trustAnchor.NotificationSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting source: %s", err)
	}

	if err := d.Set(names.AttrSource, flattenSource(trustAnchor.Source)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting source: %s", err)
	}

	return diags
}

func resourceTrustAnchorUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &rolesanywhere.UpdateTrustAnchorInput{
			TrustAnchorId: aws.String(d.Id()),
			Name:          aws.String(d.Get(names.AttrName).(string)),
			Source:        expandSource(d.Get(names.AttrSource).([]any)),
		}

		log.Printf("[DEBUG] Updating RolesAnywhere Trust Anchor (%s): %#v", d.Id(), input)
		_, err := conn.UpdateTrustAnchor(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RolesAnywhere Trust Anchor (%s): %s", d.Id(), err)
		}

		if d.HasChange(names.AttrEnabled) {
			_, n := d.GetChange(names.AttrEnabled)
			if n == true {
				if err := enableTrustAnchor(ctx, d.Id(), meta); err != nil {
					return sdkdiag.AppendErrorf(diags, "enabling RolesAnywhere Trust Anchor (%s): %s", d.Id(), err)
				}
			} else {
				if err := disableTrustAnchor(ctx, d.Id(), meta); err != nil {
					return sdkdiag.AppendErrorf(diags, "disabling RolesAnywhere Trust Anchor (%s): %s", d.Id(), err)
				}
			}
		}
	}

	return append(diags, resourceTrustAnchorRead(ctx, d, meta)...)
}

func resourceTrustAnchorDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	log.Printf("[DEBUG] Deleting RolesAnywhere Trust Anchor (%s)", d.Id())
	_, err := conn.DeleteTrustAnchor(ctx, &rolesanywhere.DeleteTrustAnchorInput{
		TrustAnchorId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RolesAnywhere Trust Anchor (%s): %s", d.Id(), err)
	}

	return diags
}

func findTrustAnchorByID(ctx context.Context, conn *rolesanywhere.Client, id string) (*awstypes.TrustAnchorDetail, error) {
	in := &rolesanywhere.GetTrustAnchorInput{
		TrustAnchorId: aws.String(id),
	}

	out, err := conn.GetTrustAnchor(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.TrustAnchor == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.TrustAnchor, nil
}

func flattenSource(apiObject *awstypes.Source) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{}

	m[names.AttrSourceType] = apiObject.SourceType
	m["source_data"] = flattenSourceData(apiObject.SourceData)

	return []any{m}
}

func flattenSourceData(apiObject awstypes.SourceData) []any {
	if apiObject == nil {
		return []any{}
	}

	m := map[string]any{}

	switch v := apiObject.(type) {
	case *awstypes.SourceDataMemberAcmPcaArn:
		m["acm_pca_arn"] = v.Value
	case *awstypes.SourceDataMemberX509CertificateData:
		m["x509_certificate_data"] = v.Value
	case *awstypes.UnknownUnionMember:
		log.Println("unknown tag:", v.Tag)
	default:
		log.Println("union is nil or unknown type")
	}

	return []any{m}
}

func flattenNotificationSettings(apiObjects []awstypes.NotificationSettingDetail) []map[string]any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenNotificationSetting(&apiObject))
	}

	return tfList
}

func flattenNotificationSetting(apiObject *awstypes.NotificationSettingDetail) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"channel": apiObject.Channel,
		"event":   apiObject.Event,
	}

	if v := apiObject.ConfiguredBy; v != nil {
		tfMap["configured_by"] = aws.ToString(v)
	}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.Threshold; v != nil {
		tfMap["threshold"] = aws.ToInt32(v)
	}

	return tfMap
}

func expandSource(tfList []any) *awstypes.Source {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	result := &awstypes.Source{}

	if v, ok := tfMap[names.AttrSourceType].(string); ok && v != "" {
		result.SourceType = awstypes.TrustAnchorType(v)
	}

	if v, ok := tfMap["source_data"].([]any); ok && len(v) > 0 && v[0] != nil {
		if result.SourceType == awstypes.TrustAnchorTypeAwsAcmPca {
			result.SourceData = expandSourceDataACMPCA(v[0].(map[string]any))
		} else if result.SourceType == awstypes.TrustAnchorTypeCertificateBundle {
			result.SourceData = expandSourceDataCertificateBundle(v[0].(map[string]any))
		}
	}

	return result
}

func expandNotificationSettings(tfList []any) []awstypes.NotificationSetting {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]awstypes.NotificationSetting, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandNotificationSetting(tfMap))
	}

	return apiObjects
}

func expandNotificationSetting(tfMap map[string]any) awstypes.NotificationSetting {
	apiObject := awstypes.NotificationSetting{}

	if v, ok := tfMap["channel"].(string); ok {
		apiObject.Channel = awstypes.NotificationChannel(v)
	}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["event"].(string); ok {
		apiObject.Event = awstypes.NotificationEvent(v)
	}

	if v, ok := tfMap["threshold"].(int); ok {
		apiObject.Threshold = aws.Int32(int32(v))
	}

	return apiObject
}

func expandSourceDataACMPCA(tfMap map[string]any) *awstypes.SourceDataMemberAcmPcaArn {
	result := &awstypes.SourceDataMemberAcmPcaArn{}

	if v, ok := tfMap["acm_pca_arn"].(string); ok && v != "" {
		result.Value = v
	}

	return result
}

func expandSourceDataCertificateBundle(tfMap map[string]any) *awstypes.SourceDataMemberX509CertificateData {
	result := &awstypes.SourceDataMemberX509CertificateData{}

	if v, ok := tfMap["x509_certificate_data"].(string); ok && v != "" {
		result.Value = v
	}

	return result
}

func disableTrustAnchor(ctx context.Context, trustAnchorId string, meta any) error {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	input := &rolesanywhere.DisableTrustAnchorInput{
		TrustAnchorId: aws.String(trustAnchorId),
	}

	_, err := conn.DisableTrustAnchor(ctx, input)
	return err
}

func enableTrustAnchor(ctx context.Context, trustAnchorId string, meta any) error {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	input := &rolesanywhere.EnableTrustAnchorInput{
		TrustAnchorId: aws.String(trustAnchorId),
	}

	_, err := conn.EnableTrustAnchor(ctx, input)
	return err
}
