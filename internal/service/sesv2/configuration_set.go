// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_configuration_set", name="Configuration Set")
// @Tags(identifierAttribute="arn")
func resourceConfigurationSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationSetCreate,
		ReadWithoutTimeout:   resourceConfigurationSetRead,
		UpdateWithoutTimeout: resourceConfigurationSetUpdate,
		DeleteWithoutTimeout: resourceConfigurationSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_set_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"delivery_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_delivery_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(300, 50400),
						},
						"sending_pool_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tls_policy": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.TlsPolicy](),
						},
					},
				},
			},
			"reputation_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"last_fresh_start": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"reputation_metrics_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"sending_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sending_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"suppression_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"suppressed_reasons": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[types.SuppressionListReason](),
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tracking_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_redirect_domain": {
							Type:     schema.TypeString,
							Required: true,
						},
						"https_policy": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.HttpsPolicy](),
						},
					},
				},
			},
			"vdm_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dashboard_options": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"engagement_metrics": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.FeatureStatus](),
									},
								},
							},
						},
						"guardian_options": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"optimized_shared_delivery": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.FeatureStatus](),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

const (
	resNameConfigurationSet = "Configuration Set"
)

func resourceConfigurationSetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	name := d.Get("configuration_set_name").(string)
	input := &sesv2.CreateConfigurationSetInput{
		ConfigurationSetName: aws.String(name),
		Tags:                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("delivery_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DeliveryOptions = expandDeliveryOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("reputation_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ReputationOptions = expandReputationOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("sending_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.SendingOptions = expandSendingOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetRawConfig().AsValueMap()["suppression_options"]; ok && v.LengthInt() > 0 {
		if v, ok := v.Index(cty.NumberIntVal(0)).AsValueMap()["suppressed_reasons"]; ok && !v.IsNull() {
			tfMap := map[string]any{
				"suppressed_reasons": []any{},
			}

			for _, v := range v.AsValueSlice() {
				tfMap["suppressed_reasons"] = append(tfMap["suppressed_reasons"].([]any), v.AsString())
			}

			input.SuppressionOptions = expandSuppressionOptions(tfMap)
		}
	}

	if v, ok := d.GetOk("tracking_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.TrackingOptions = expandTrackingOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("vdm_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.VdmOptions = expandVDMOptions(v.([]any)[0].(map[string]any))
	}

	_, err := conn.CreateConfigurationSet(ctx, input)

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameConfigurationSet, name, err)
	}

	d.SetId(name)

	return append(diags, resourceConfigurationSetRead(ctx, d, meta)...)
}

func resourceConfigurationSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	output, err := findConfigurationSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 ConfigurationSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, resNameConfigurationSet, d.Id(), err)
	}

	d.Set(names.AttrARN, configurationSetARN(ctx, meta.(*conns.AWSClient), aws.ToString(output.ConfigurationSetName)))
	d.Set("configuration_set_name", output.ConfigurationSetName)
	if output.DeliveryOptions != nil {
		if err := d.Set("delivery_options", []any{flattenDeliveryOptions(output.DeliveryOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting delivery_options: %s", err)
		}
	} else {
		d.Set("delivery_options", nil)
	}
	if output.ReputationOptions != nil {
		if err := d.Set("reputation_options", []any{flattenReputationOptions(output.ReputationOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting reputation_options: %s", err)
		}
	} else {
		d.Set("reputation_options", nil)
	}
	if output.SendingOptions != nil {
		if err := d.Set("sending_options", []any{flattenSendingOptions(output.SendingOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting sending_options: %s", err)
		}
	} else {
		d.Set("sending_options", nil)
	}
	if output.SuppressionOptions != nil {
		if err := d.Set("suppression_options", []any{flattenSuppressionOptions(output.SuppressionOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting suppression_options: %s", err)
		}
	} else {
		d.Set("suppression_options", nil)
	}
	if output.TrackingOptions != nil {
		if err := d.Set("tracking_options", []any{flattenTrackingOptions(output.TrackingOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tracking_options: %s", err)
		}
	} else {
		d.Set("tracking_options", nil)
	}
	if output.VdmOptions != nil {
		if err := d.Set("vdm_options", []any{flattenVDMOptions(output.VdmOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vdm_options: %s", err)
		}
	} else {
		d.Set("vdm_options", nil)
	}

	return diags
}

func resourceConfigurationSetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	if d.HasChanges("delivery_options") {
		input := &sesv2.PutConfigurationSetDeliveryOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("delivery_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			tfMap := v.([]any)[0].(map[string]any)

			if v, ok := tfMap["max_delivery_seconds"].(int); ok && v != 0 {
				input.MaxDeliverySeconds = aws.Int64(int64(v))
			}

			if v, ok := tfMap["sending_pool_name"].(string); ok && v != "" {
				input.SendingPoolName = aws.String(v)
			}

			if v, ok := tfMap["tls_policy"].(string); ok && v != "" {
				input.TlsPolicy = types.TlsPolicy(v)
			}
		}

		_, err := conn.PutConfigurationSetDeliveryOptions(ctx, input)

		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, resNameConfigurationSet, d.Id(), err)
		}
	}

	if d.HasChanges("reputation_options") {
		input := &sesv2.PutConfigurationSetReputationOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("reputation_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			tfMap := v.([]any)[0].(map[string]any)

			if v, ok := tfMap["reputation_metrics_enabled"].(bool); ok {
				input.ReputationMetricsEnabled = v
			}
		}

		_, err := conn.PutConfigurationSetReputationOptions(ctx, input)

		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, resNameConfigurationSet, d.Id(), err)
		}
	}

	if d.HasChanges("sending_options") {
		input := &sesv2.PutConfigurationSetSendingOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("sending_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			tfMap := v.([]any)[0].(map[string]any)

			if v, ok := tfMap["sending_enabled"].(bool); ok {
				input.SendingEnabled = v
			}
		}

		_, err := conn.PutConfigurationSetSendingOptions(ctx, input)

		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, resNameConfigurationSet, d.Id(), err)
		}
	}

	if d.HasChanges("suppression_options") {
		input := &sesv2.PutConfigurationSetSuppressionOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("suppression_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			tfMap := v.([]any)[0].(map[string]any)

			if v, ok := tfMap["suppressed_reasons"].([]any); ok && len(v) > 0 {
				input.SuppressedReasons = flex.ExpandStringyValueList[types.SuppressionListReason](v)
			}
		}

		_, err := conn.PutConfigurationSetSuppressionOptions(ctx, input)

		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, resNameConfigurationSet, d.Id(), err)
		}
	}

	if d.HasChanges("tracking_options") {
		input := &sesv2.PutConfigurationSetTrackingOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("tracking_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			tfMap := v.([]any)[0].(map[string]any)

			if v, ok := tfMap["custom_redirect_domain"].(string); ok && v != "" {
				input.CustomRedirectDomain = aws.String(v)
			}

			if v, ok := tfMap["https_policy"].(string); ok && v != "" {
				input.HttpsPolicy = types.HttpsPolicy(v)
			}
		}

		_, err := conn.PutConfigurationSetTrackingOptions(ctx, input)

		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, resNameConfigurationSet, d.Id(), err)
		}
	}

	if d.HasChanges("vdm_options") {
		input := &sesv2.PutConfigurationSetVdmOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("vdm_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.VdmOptions = expandVDMOptions(v.([]any)[0].(map[string]any))
		}

		_, err := conn.PutConfigurationSetVdmOptions(ctx, input)

		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, resNameConfigurationSet, d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationSetRead(ctx, d, meta)...)
}

func resourceConfigurationSetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	log.Printf("[INFO] Deleting SESV2 ConfigurationSet: %s", d.Id())
	_, err := conn.DeleteConfigurationSet(ctx, &sesv2.DeleteConfigurationSetInput{
		ConfigurationSetName: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, resNameConfigurationSet, d.Id(), err)
	}

	return diags
}

func findConfigurationSetByID(ctx context.Context, conn *sesv2.Client, id string) (*sesv2.GetConfigurationSetOutput, error) {
	input := &sesv2.GetConfigurationSetInput{
		ConfigurationSetName: aws.String(id),
	}

	return findConfigurationSet(ctx, conn, input)
}

func findConfigurationSet(ctx context.Context, conn *sesv2.Client, input *sesv2.GetConfigurationSetInput) (*sesv2.GetConfigurationSetOutput, error) {
	output, err := conn.GetConfigurationSet(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func flattenDeliveryOptions(apiObject *types.DeliveryOptions) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"tls_policy": apiObject.TlsPolicy,
	}

	if v := apiObject.MaxDeliverySeconds; v != nil {
		tfMap["max_delivery_seconds"] = aws.ToInt64(v)
	}

	if v := apiObject.SendingPoolName; v != nil {
		tfMap["sending_pool_name"] = aws.ToString(v)
	}

	return tfMap
}

func flattenReputationOptions(apiObject *types.ReputationOptions) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"reputation_metrics_enabled": apiObject.ReputationMetricsEnabled,
	}

	if v := apiObject.LastFreshStart; v != nil {
		tfMap["last_fresh_start"] = v.Format(time.RFC3339)
	}

	return tfMap
}

func flattenSendingOptions(apiObject *types.SendingOptions) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"sending_enabled": apiObject.SendingEnabled,
	}

	return tfMap
}

func flattenSuppressionOptions(apiObject *types.SuppressionOptions) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.SuppressedReasons; v != nil {
		tfMap["suppressed_reasons"] = apiObject.SuppressedReasons
	}

	return tfMap
}

func flattenTrackingOptions(apiObject *types.TrackingOptions) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"https_policy": apiObject.HttpsPolicy,
	}

	if v := apiObject.CustomRedirectDomain; v != nil {
		tfMap["custom_redirect_domain"] = aws.ToString(v)
	}

	return tfMap
}

func flattenVDMOptions(apiObject *types.VdmOptions) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.DashboardOptions; v != nil {
		tfMap["dashboard_options"] = []any{flattenDashboardOptions(v)}
	}

	if v := apiObject.GuardianOptions; v != nil {
		tfMap["guardian_options"] = []any{flattenGuardianOptions(v)}
	}

	return tfMap
}

func flattenDashboardOptions(apiObject *types.DashboardOptions) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"engagement_metrics": apiObject.EngagementMetrics,
	}

	return tfMap
}

func flattenGuardianOptions(apiObject *types.GuardianOptions) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"optimized_shared_delivery": apiObject.OptimizedSharedDelivery,
	}

	return tfMap
}

func expandDeliveryOptions(tfMap map[string]any) *types.DeliveryOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.DeliveryOptions{}

	if v, ok := tfMap["max_delivery_seconds"].(int); ok && v != 0 {
		apiObject.MaxDeliverySeconds = aws.Int64(int64(v))
	}

	if v, ok := tfMap["sending_pool_name"].(string); ok && v != "" {
		apiObject.SendingPoolName = aws.String(v)
	}

	if v, ok := tfMap["tls_policy"].(string); ok && v != "" {
		apiObject.TlsPolicy = types.TlsPolicy(v)
	}

	return apiObject
}

func expandReputationOptions(tfMap map[string]any) *types.ReputationOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ReputationOptions{}

	if v, ok := tfMap["reputation_metrics_enabled"].(bool); ok {
		apiObject.ReputationMetricsEnabled = v
	}

	return apiObject
}

func expandSendingOptions(tfMap map[string]any) *types.SendingOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SendingOptions{}

	if v, ok := tfMap["sending_enabled"].(bool); ok {
		apiObject.SendingEnabled = v
	}

	return apiObject
}

func expandSuppressionOptions(tfMap map[string]any) *types.SuppressionOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SuppressionOptions{}

	if v, ok := tfMap["suppressed_reasons"].([]any); ok {
		if len(v) > 0 {
			apiObject.SuppressedReasons = flex.ExpandStringyValueList[types.SuppressionListReason](v)
		} else {
			apiObject.SuppressedReasons = make([]types.SuppressionListReason, 0)
		}
	}

	return apiObject
}

func expandTrackingOptions(tfMap map[string]any) *types.TrackingOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.TrackingOptions{}

	if v, ok := tfMap["custom_redirect_domain"].(string); ok && v != "" {
		apiObject.CustomRedirectDomain = aws.String(v)
	}

	if v, ok := tfMap["https_policy"].(string); ok && v != "" {
		apiObject.HttpsPolicy = types.HttpsPolicy(v)
	}

	return apiObject
}

func expandVDMOptions(tfMap map[string]any) *types.VdmOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.VdmOptions{}

	if v, ok := tfMap["dashboard_options"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.DashboardOptions = expandDashboardOptions(v[0].(map[string]any))
	}

	if v, ok := tfMap["guardian_options"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.GuardianOptions = expandGuardianOptions(v[0].(map[string]any))
	}

	return apiObject
}

func expandDashboardOptions(tfMap map[string]any) *types.DashboardOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.DashboardOptions{}

	if v, ok := tfMap["engagement_metrics"].(string); ok && v != "" {
		apiObject.EngagementMetrics = types.FeatureStatus(v)
	}

	return apiObject
}

func expandGuardianOptions(tfMap map[string]any) *types.GuardianOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.GuardianOptions{}

	if v, ok := tfMap["optimized_shared_delivery"].(string); ok && v != "" {
		apiObject.OptimizedSharedDelivery = types.FeatureStatus(v)
	}

	return apiObject
}

func configurationSetARN(ctx context.Context, c *conns.AWSClient, configurationSetName string) string {
	return c.RegionalARN(ctx, "ses", fmt.Sprintf("configuration-set/%s", configurationSetName))
}
