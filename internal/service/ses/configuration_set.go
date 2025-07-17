// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_configuration_set", name="Configuration Set")
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
			"delivery_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tls_policy": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.TlsPolicyOptional,
							ValidateDiagFunc: enum.Validate[awstypes.TlsPolicy](),
						},
					},
				},
			},
			"last_fresh_start": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"reputation_metrics_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"sending_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"tracking_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_redirect_domain": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringDoesNotMatch(regexache.MustCompile(`\.$`), "cannot end with a period"),
						},
					},
				},
			},
		},
	}
}

func resourceConfigurationSetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	configurationSetName := d.Get(names.AttrName).(string)
	input := &ses.CreateConfigurationSetInput{
		ConfigurationSet: &awstypes.ConfigurationSet{
			Name: aws.String(configurationSetName),
		},
	}

	_, err := conn.CreateConfigurationSet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES Configuration Set (%s): %s", configurationSetName, err)
	}

	d.SetId(configurationSetName)

	if v, ok := d.GetOk("delivery_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input := &ses.PutConfigurationSetDeliveryOptionsInput{
			ConfigurationSetName: aws.String(configurationSetName),
			DeliveryOptions:      expandDeliveryOptions(v.([]any)),
		}

		_, err := conn.PutConfigurationSetDeliveryOptions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting SES Configuration Set (%s) delivery options: %s", d.Id(), err)
		}
	}

	if v := d.Get("reputation_metrics_enabled"); v.(bool) {
		input := &ses.UpdateConfigurationSetReputationMetricsEnabledInput{
			ConfigurationSetName: aws.String(configurationSetName),
			Enabled:              v.(bool),
		}

		_, err := conn.UpdateConfigurationSetReputationMetricsEnabled(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling SES Configuration Set (%s) reputation metrics %s", d.Id(), err)
		}
	}

	if v := d.Get("sending_enabled"); !v.(bool) {
		input := &ses.UpdateConfigurationSetSendingEnabledInput{
			ConfigurationSetName: aws.String(configurationSetName),
			Enabled:              v.(bool),
		}

		_, err := conn.UpdateConfigurationSetSendingEnabled(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling SES Configuration Set (%s) sending: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("tracking_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input := &ses.CreateConfigurationSetTrackingOptionsInput{
			ConfigurationSetName: aws.String(configurationSetName),
			TrackingOptions:      expandTrackingOptions(v.([]any)),
		}

		_, err := conn.CreateConfigurationSetTrackingOptions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating SES Configuration Set (%s) tracking options: %s", d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationSetRead(ctx, d, meta)...)
}

func resourceConfigurationSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	input := &ses.DescribeConfigurationSetInput{
		ConfigurationSetAttributeNames: []awstypes.ConfigurationSetAttribute{
			awstypes.ConfigurationSetAttributeDeliveryOptions,
			awstypes.ConfigurationSetAttributeReputationOptions,
			awstypes.ConfigurationSetAttributeTrackingOptions,
		},
		ConfigurationSetName: aws.String(d.Id()),
	}

	output, err := findConfigurationSet(ctx, conn, input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Configuration Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Configuration Set (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("configuration-set/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	if err := d.Set("delivery_options", flattenDeliveryOptions(output.DeliveryOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting delivery_options: %s", err)
	}
	d.Set(names.AttrName, output.ConfigurationSet.Name)
	if err := d.Set("tracking_options", flattenTrackingOptions(output.TrackingOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tracking_options: %s", err)
	}

	if apiObject := output.ReputationOptions; apiObject != nil {
		d.Set("last_fresh_start", aws.ToTime(apiObject.LastFreshStart).Format(time.RFC3339))
		d.Set("reputation_metrics_enabled", apiObject.ReputationMetricsEnabled)
		d.Set("sending_enabled", apiObject.SendingEnabled)
	}

	return diags
}

func resourceConfigurationSetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	if d.HasChange("delivery_options") {
		input := &ses.PutConfigurationSetDeliveryOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
			DeliveryOptions:      expandDeliveryOptions(d.Get("delivery_options").([]any)),
		}

		_, err := conn.PutConfigurationSetDeliveryOptions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SES Configuration Set (%s) delivery options: %s", d.Id(), err)
		}
	}

	if d.HasChange("reputation_metrics_enabled") {
		input := &ses.UpdateConfigurationSetReputationMetricsEnabledInput{
			ConfigurationSetName: aws.String(d.Id()),
			Enabled:              d.Get("reputation_metrics_enabled").(bool),
		}

		_, err := conn.UpdateConfigurationSetReputationMetricsEnabled(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SES Configuration Set (%s) reputation metrics enabled: %s", d.Id(), err)
		}
	}

	if d.HasChange("sending_enabled") {
		input := &ses.UpdateConfigurationSetSendingEnabledInput{
			ConfigurationSetName: aws.String(d.Id()),
			Enabled:              d.Get("sending_enabled").(bool),
		}

		_, err := conn.UpdateConfigurationSetSendingEnabled(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SES Configuration Set (%s) reputation metrics enabled: %s", d.Id(), err)
		}
	}

	if d.HasChange("tracking_options") {
		input := &ses.UpdateConfigurationSetTrackingOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
			TrackingOptions:      expandTrackingOptions(d.Get("tracking_options").([]any)),
		}

		_, err := conn.UpdateConfigurationSetTrackingOptions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SES Configuration Set (%s) tracking options: %s", d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationSetRead(ctx, d, meta)...)
}

func resourceConfigurationSetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	log.Printf("[DEBUG] Deleting Configuration Set: %s", d.Id())
	_, err := conn.DeleteConfigurationSet(ctx, &ses.DeleteConfigurationSetInput{
		ConfigurationSetName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ConfigurationSetDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Configuration Set (%s): %s", d.Id(), err)
	}

	return diags
}

func findConfigurationSetByName(ctx context.Context, conn *ses.Client, name string) (*ses.DescribeConfigurationSetOutput, error) {
	input := &ses.DescribeConfigurationSetInput{
		ConfigurationSetName: aws.String(name),
	}

	return findConfigurationSet(ctx, conn, input)
}

func findConfigurationSet(ctx context.Context, conn *ses.Client, input *ses.DescribeConfigurationSetInput) (*ses.DescribeConfigurationSetOutput, error) {
	output, err := conn.DescribeConfigurationSet(ctx, input)

	if errs.IsA[*awstypes.ConfigurationSetDoesNotExistException](err) {
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

func expandDeliveryOptions(tfList []any) *awstypes.DeliveryOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DeliveryOptions{}

	if v, ok := tfMap["tls_policy"].(string); ok && v != "" {
		apiObject.TlsPolicy = awstypes.TlsPolicy(v)
	}

	return apiObject
}

func flattenDeliveryOptions(apiObject *awstypes.DeliveryOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"tls_policy": string(apiObject.TlsPolicy),
	}

	return []any{tfMap}
}

func expandTrackingOptions(tfList []any) *awstypes.TrackingOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TrackingOptions{}

	if v, ok := tfMap["custom_redirect_domain"].(string); ok && v != "" {
		apiObject.CustomRedirectDomain = aws.String(v)
	}

	return apiObject
}

func flattenTrackingOptions(apiObject *awstypes.TrackingOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"custom_redirect_domain": aws.ToString(apiObject.CustomRedirectDomain),
	}

	return []any{tfMap}
}
