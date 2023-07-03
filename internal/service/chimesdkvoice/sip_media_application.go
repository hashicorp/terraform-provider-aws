// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_chimesdkvoice_sip_media_application", name="Sip Media Application")
// @Tags(identifierAttribute="arn")
func ResourceSipMediaApplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSipMediaApplicationCreate,
		ReadWithoutTimeout:   resourceSipMediaApplicationRead,
		UpdateWithoutTimeout: resourceSipMediaApplicationUpdate,
		DeleteWithoutTimeout: resourceSipMediaApplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_region": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"endpoints": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lambda_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSipMediaApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	createInput := &chimesdkvoice.CreateSipMediaApplicationInput{
		AwsRegion: aws.String(d.Get("aws_region").(string)),
		Name:      aws.String(d.Get("name").(string)),
		Endpoints: expandSipMediaApplicationEndpoints(d.Get("endpoints").([]interface{})),
		Tags:      getTagsIn(ctx),
	}

	resp, err := conn.CreateSipMediaApplicationWithContext(ctx, createInput)
	if err != nil || resp.SipMediaApplication == nil {
		return sdkdiag.AppendErrorf(diags, "creating Chime Sip Media Application: %s", err)
	}

	d.SetId(aws.StringValue(resp.SipMediaApplication.SipMediaApplicationId))
	return append(diags, resourceSipMediaApplicationRead(ctx, d, meta)...)
}

func resourceSipMediaApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	getInput := &chimesdkvoice.GetSipMediaApplicationInput{
		SipMediaApplicationId: aws.String(d.Id()),
	}

	resp, err := conn.GetSipMediaApplicationWithContext(ctx, getInput)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
		log.Printf("[WARN] Chime Sip Media Application %s not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil || resp.SipMediaApplication == nil {
		return sdkdiag.AppendErrorf(diags, "getting Sip Media Application (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.SipMediaApplication.SipMediaApplicationArn)
	d.Set("aws_region", resp.SipMediaApplication.AwsRegion)
	d.Set("name", resp.SipMediaApplication.Name)
	d.Set("endpoints", flattenSipMediaApplicationEndpoints(resp.SipMediaApplication.Endpoints))

	return diags
}

func resourceSipMediaApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	if d.HasChanges("name", "endpoints") {
		updateInput := &chimesdkvoice.UpdateSipMediaApplicationInput{
			SipMediaApplicationId: aws.String(d.Id()),
			Name:                  aws.String(d.Get("name").(string)),
			Endpoints:             expandSipMediaApplicationEndpoints(d.Get("endpoints").([]interface{})),
		}

		if _, err := conn.UpdateSipMediaApplicationWithContext(ctx, updateInput); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Sip Media Application (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSipMediaApplicationRead(ctx, d, meta)...)
}

func resourceSipMediaApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	input := &chimesdkvoice.DeleteSipMediaApplicationInput{
		SipMediaApplicationId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteSipMediaApplicationWithContext(ctx, input); err != nil {
		if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
			log.Printf("[WARN] Chime Sip Media Application %s not found", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Sip Media Application (%s)", d.Id())
	}

	return diags
}

func expandSipMediaApplicationEndpoints(data []interface{}) []*chimesdkvoice.SipMediaApplicationEndpoint {
	var sipMediaApplicationEndpoint []*chimesdkvoice.SipMediaApplicationEndpoint

	tfMap, ok := data[0].(map[string]interface{})
	if !ok {
		return nil
	}

	sipMediaApplicationEndpoint = append(sipMediaApplicationEndpoint, &chimesdkvoice.SipMediaApplicationEndpoint{
		LambdaArn: aws.String(tfMap["lambda_arn"].(string))})
	return sipMediaApplicationEndpoint
}

func flattenSipMediaApplicationEndpoints(apiObject []*chimesdkvoice.SipMediaApplicationEndpoint) []interface{} {
	var rawSipMediaApplicationEndpoints []interface{}

	for _, e := range apiObject {
		rawEndpoint := map[string]interface{}{
			"lambda_arn": aws.StringValue(e.LambdaArn),
		}
		rawSipMediaApplicationEndpoints = append(rawSipMediaApplicationEndpoints, rawEndpoint)
	}
	return rawSipMediaApplicationEndpoints
}
