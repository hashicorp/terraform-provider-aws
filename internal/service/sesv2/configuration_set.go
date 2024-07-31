// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_configuration_set", name="Configuration Set")
// @Tags(identifierAttribute="arn")
func ResourceConfigurationSet() *schema.Resource {
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
							MinItems: 1,
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameConfigurationSet = "Configuration Set"
)

func resourceConfigurationSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	in := &sesv2.CreateConfigurationSetInput{
		ConfigurationSetName: aws.String(d.Get("configuration_set_name").(string)),
		Tags:                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("delivery_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.DeliveryOptions = expandDeliveryOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("reputation_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.ReputationOptions = expandReputationOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("sending_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.SendingOptions = expandSendingOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("suppression_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.SuppressionOptions = expandSuppressionOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("tracking_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.TrackingOptions = expandTrackingOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("vdm_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.VdmOptions = expandVDMOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	out, err := conn.CreateConfigurationSet(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameConfigurationSet, d.Get("configuration_set_name").(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameConfigurationSet, d.Get("configuration_set_name").(string), errors.New("empty output"))
	}

	d.SetId(d.Get("configuration_set_name").(string))

	return append(diags, resourceConfigurationSetRead(ctx, d, meta)...)
}

func resourceConfigurationSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	out, err := FindConfigurationSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 ConfigurationSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameConfigurationSet, d.Id(), err)
	}

	d.Set(names.AttrARN, configurationSetNameToARN(meta, aws.ToString(out.ConfigurationSetName)))
	d.Set("configuration_set_name", out.ConfigurationSetName)

	if out.DeliveryOptions != nil {
		if err := d.Set("delivery_options", []interface{}{flattenDeliveryOptions(out.DeliveryOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("delivery_options", nil)
	}

	if out.ReputationOptions != nil {
		if err := d.Set("reputation_options", []interface{}{flattenReputationOptions(out.ReputationOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("reputation_options", nil)
	}

	if out.SendingOptions != nil {
		if err := d.Set("sending_options", []interface{}{flattenSendingOptions(out.SendingOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("sending_options", nil)
	}

	if out.SuppressionOptions != nil {
		if err := d.Set("suppression_options", []interface{}{flattenSuppressionOptions(out.SuppressionOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("suppression_options", nil)
	}

	if out.TrackingOptions != nil {
		if err := d.Set("tracking_options", []interface{}{flattenTrackingOptions(out.TrackingOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("tracking_options", nil)
	}

	if out.VdmOptions != nil {
		if err := d.Set("vdm_options", []interface{}{flattenVDMOptions(out.VdmOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("vdm_options", nil)
	}

	return diags
}

func resourceConfigurationSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	if d.HasChanges("delivery_options") {
		in := &sesv2.PutConfigurationSetDeliveryOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("delivery_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			if v, ok := tfMap["sending_pool_name"].(string); ok && v != "" {
				in.SendingPoolName = aws.String(v)
			}

			if v, ok := tfMap["tls_policy"].(string); ok && v != "" {
				in.TlsPolicy = types.TlsPolicy(v)
			}
		}

		log.Printf("[DEBUG] Updating SESV2 ConfigurationSet DeliveryOptions (%s): %#v", d.Id(), in)
		_, err := conn.PutConfigurationSetDeliveryOptions(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, ResNameConfigurationSet, d.Id(), err)
		}
	}

	if d.HasChanges("reputation_options") {
		in := &sesv2.PutConfigurationSetReputationOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("reputation_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			if v, ok := tfMap["reputation_metrics_enabled"].(bool); ok {
				in.ReputationMetricsEnabled = v
			}
		}

		log.Printf("[DEBUG] Updating SESV2 ConfigurationSet ReputationOptions (%s): %#v", d.Id(), in)
		_, err := conn.PutConfigurationSetReputationOptions(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, ResNameConfigurationSet, d.Id(), err)
		}
	}

	if d.HasChanges("sending_options") {
		if v, ok := d.GetOk("sending_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			in := &sesv2.PutConfigurationSetSendingOptionsInput{
				ConfigurationSetName: aws.String(d.Id()),
			}

			if v, ok := tfMap["sending_enabled"].(bool); ok {
				in.SendingEnabled = v
			}

			log.Printf("[DEBUG] Updating SESV2 ConfigurationSet SendingOptions (%s): %#v", d.Id(), in)
			_, err := conn.PutConfigurationSetSendingOptions(ctx, in)
			if err != nil {
				return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, ResNameConfigurationSet, d.Id(), err)
			}
		}
	}

	if d.HasChanges("suppression_options") {
		in := &sesv2.PutConfigurationSetSuppressionOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("suppression_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			if v, ok := tfMap["suppressed_reasons"].([]interface{}); ok && len(v) > 0 {
				in.SuppressedReasons = expandSuppressedReasons(v)
			}
		}

		log.Printf("[DEBUG] Updating SESV2 ConfigurationSet SuppressionOptions (%s): %#v", d.Id(), in)
		_, err := conn.PutConfigurationSetSuppressionOptions(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, ResNameConfigurationSet, d.Id(), err)
		}
	}

	if d.HasChanges("tracking_options") {
		in := &sesv2.PutConfigurationSetTrackingOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("tracking_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			if v, ok := tfMap["custom_redirect_domain"].(string); ok && v != "" {
				in.CustomRedirectDomain = aws.String(v)
			}
		}

		log.Printf("[DEBUG] Updating SESV2 ConfigurationSet TrackingOptions (%s): %#v", d.Id(), in)
		_, err := conn.PutConfigurationSetTrackingOptions(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, ResNameConfigurationSet, d.Id(), err)
		}
	}

	if d.HasChanges("vdm_options") {
		in := &sesv2.PutConfigurationSetVdmOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("vdm_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			in.VdmOptions = expandVDMOptions(v.([]interface{})[0].(map[string]interface{}))
		}

		log.Printf("[DEBUG] Updating SESV2 ConfigurationSet VdmOptions (%s): %#v", d.Id(), in)
		_, err := conn.PutConfigurationSetVdmOptions(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, ResNameConfigurationSet, d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationSetRead(ctx, d, meta)...)
}

func resourceConfigurationSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	log.Printf("[INFO] Deleting SESV2 ConfigurationSet %s", d.Id())

	_, err := conn.DeleteConfigurationSet(ctx, &sesv2.DeleteConfigurationSetInput{
		ConfigurationSetName: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, ResNameConfigurationSet, d.Id(), err)
	}

	return diags
}

func FindConfigurationSetByID(ctx context.Context, conn *sesv2.Client, id string) (*sesv2.GetConfigurationSetOutput, error) {
	in := &sesv2.GetConfigurationSetInput{
		ConfigurationSetName: aws.String(id),
	}
	out, err := conn.GetConfigurationSet(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenDeliveryOptions(apiObject *types.DeliveryOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SendingPoolName; v != nil {
		m["sending_pool_name"] = aws.ToString(v)
	}

	m["tls_policy"] = string(apiObject.TlsPolicy)

	return m
}

func flattenReputationOptions(apiObject *types.ReputationOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"reputation_metrics_enabled": apiObject.ReputationMetricsEnabled,
	}

	if v := apiObject.LastFreshStart; v != nil {
		m["last_fresh_start"] = v.Format(time.RFC3339)
	}

	return m
}

func flattenSendingOptions(apiObject *types.SendingOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"sending_enabled": apiObject.SendingEnabled,
	}

	return m
}

func flattenSuppressionOptions(apiObject *types.SuppressionOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SuppressedReasons; v != nil {
		m["suppressed_reasons"] = flattenSuppressedReasons(apiObject.SuppressedReasons)
	}

	return m
}

func flattenSuppressedReasons(apiObject []types.SuppressionListReason) []string {
	var out []string

	for _, v := range apiObject {
		out = append(out, string(v))
	}

	return out
}

func flattenTrackingOptions(apiObject *types.TrackingOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.CustomRedirectDomain; v != nil {
		m["custom_redirect_domain"] = aws.ToString(v)
	}

	return m
}

func flattenVDMOptions(apiObject *types.VdmOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.DashboardOptions; v != nil {
		m["dashboard_options"] = []interface{}{flattenDashboardOptions(v)}
	}

	if v := apiObject.GuardianOptions; v != nil {
		m["guardian_options"] = []interface{}{flattenGuardianOptions(v)}
	}

	return m
}

func flattenDashboardOptions(apiObject *types.DashboardOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"engagement_metrics": string(apiObject.EngagementMetrics),
	}

	return m
}

func flattenGuardianOptions(apiObject *types.GuardianOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"optimized_shared_delivery": string(apiObject.OptimizedSharedDelivery),
	}

	return m
}

func expandDeliveryOptions(tfMap map[string]interface{}) *types.DeliveryOptions {
	if tfMap == nil {
		return nil
	}

	a := &types.DeliveryOptions{}

	if v, ok := tfMap["sending_pool_name"].(string); ok && v != "" {
		a.SendingPoolName = aws.String(v)
	}

	if v, ok := tfMap["tls_policy"].(string); ok && v != "" {
		a.TlsPolicy = types.TlsPolicy(v)
	}

	return a
}

func expandReputationOptions(tfMap map[string]interface{}) *types.ReputationOptions {
	if tfMap == nil {
		return nil
	}

	a := &types.ReputationOptions{}

	if v, ok := tfMap["reputation_metrics_enabled"].(bool); ok {
		a.ReputationMetricsEnabled = v
	}

	return a
}

func expandSendingOptions(tfMap map[string]interface{}) *types.SendingOptions {
	if tfMap == nil {
		return nil
	}

	a := &types.SendingOptions{}

	if v, ok := tfMap["sending_enabled"].(bool); ok {
		a.SendingEnabled = v
	}

	return a
}

func expandSuppressionOptions(tfMap map[string]interface{}) *types.SuppressionOptions {
	if tfMap == nil {
		return nil
	}

	a := &types.SuppressionOptions{}

	if v, ok := tfMap["suppressed_reasons"].([]interface{}); ok && len(v) > 0 {
		a.SuppressedReasons = expandSuppressedReasons(v)
	}

	return a
}

func expandSuppressedReasons(tfList []interface{}) []types.SuppressionListReason {
	var out []types.SuppressionListReason

	for _, v := range tfList {
		if v, ok := v.(string); ok && v != "" {
			out = append(out, types.SuppressionListReason(v))
		}
	}

	return out
}

func expandTrackingOptions(tfMap map[string]interface{}) *types.TrackingOptions {
	if tfMap == nil {
		return nil
	}

	a := &types.TrackingOptions{}

	if v, ok := tfMap["custom_redirect_domain"].(string); ok && v != "" {
		a.CustomRedirectDomain = aws.String(v)
	}

	return a
}

func expandVDMOptions(tfMap map[string]interface{}) *types.VdmOptions {
	if tfMap == nil {
		return nil
	}

	a := &types.VdmOptions{}

	if v, ok := tfMap["dashboard_options"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.DashboardOptions = expandDashboardOptions(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["guardian_options"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.GuardianOptions = expandGuardianOptions(v[0].(map[string]interface{}))
	}

	return a
}

func expandDashboardOptions(tfMap map[string]interface{}) *types.DashboardOptions {
	if tfMap == nil {
		return nil
	}

	a := &types.DashboardOptions{}

	if v, ok := tfMap["engagement_metrics"].(string); ok && v != "" {
		a.EngagementMetrics = types.FeatureStatus(v)
	}

	return a
}

func expandGuardianOptions(tfMap map[string]interface{}) *types.GuardianOptions {
	if tfMap == nil {
		return nil
	}

	a := &types.GuardianOptions{}

	if v, ok := tfMap["optimized_shared_delivery"].(string); ok && v != "" {
		a.OptimizedSharedDelivery = types.FeatureStatus(v)
	}

	return a
}

func configurationSetNameToARN(meta interface{}, configurationSetName string) string {
	return arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("configuration-set/%s", configurationSetName),
	}.String()
}
