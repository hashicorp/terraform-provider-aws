// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_region": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			names.AttrEndpoints: {
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
			names.AttrName: {
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
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	createInput := &chimesdkvoice.CreateSipMediaApplicationInput{
		AwsRegion: aws.String(d.Get("aws_region").(string)),
		Name:      aws.String(d.Get(names.AttrName).(string)),
		Endpoints: expandSipMediaApplicationEndpoints(d.Get(names.AttrEndpoints).([]interface{})),
		Tags:      getTagsIn(ctx),
	}

	resp, err := conn.CreateSipMediaApplication(ctx, createInput)
	if err != nil || resp.SipMediaApplication == nil {
		return sdkdiag.AppendErrorf(diags, "creating Chime Sip Media Application: %s", err)
	}

	d.SetId(aws.ToString(resp.SipMediaApplication.SipMediaApplicationId))
	return append(diags, resourceSipMediaApplicationRead(ctx, d, meta)...)
}

func resourceSipMediaApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	resp, err := FindSIPResourceWithRetry(ctx, d.IsNewResource(), func() (*awstypes.SipMediaApplication, error) {
		return findSIPMediaApplicationByID(ctx, conn, d.Id())
	})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Chime Sip Media Application %s not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Chime Sip Media Application (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, resp.SipMediaApplicationArn)
	d.Set("aws_region", resp.AwsRegion)
	d.Set(names.AttrName, resp.Name)
	d.Set(names.AttrEndpoints, flattenSipMediaApplicationEndpoints(resp.Endpoints))

	return diags
}

func resourceSipMediaApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	if d.HasChanges(names.AttrName, names.AttrEndpoints) {
		updateInput := &chimesdkvoice.UpdateSipMediaApplicationInput{
			SipMediaApplicationId: aws.String(d.Id()),
			Name:                  aws.String(d.Get(names.AttrName).(string)),
			Endpoints:             expandSipMediaApplicationEndpoints(d.Get(names.AttrEndpoints).([]interface{})),
		}

		if _, err := conn.UpdateSipMediaApplication(ctx, updateInput); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Sip Media Application (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSipMediaApplicationRead(ctx, d, meta)...)
}

func resourceSipMediaApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.DeleteSipMediaApplicationInput{
		SipMediaApplicationId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteSipMediaApplication(ctx, input); err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			log.Printf("[WARN] Chime Sip Media Application %s not found", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Sip Media Application (%s)", d.Id())
	}

	return diags
}

func expandSipMediaApplicationEndpoints(data []interface{}) []awstypes.SipMediaApplicationEndpoint {
	var sipMediaApplicationEndpoint []awstypes.SipMediaApplicationEndpoint

	tfMap, ok := data[0].(map[string]interface{})
	if !ok {
		return nil
	}

	sipMediaApplicationEndpoint = append(sipMediaApplicationEndpoint, awstypes.SipMediaApplicationEndpoint{
		LambdaArn: aws.String(tfMap["lambda_arn"].(string))})
	return sipMediaApplicationEndpoint
}

func flattenSipMediaApplicationEndpoints(apiObject []awstypes.SipMediaApplicationEndpoint) []interface{} {
	var rawSipMediaApplicationEndpoints []interface{}

	for _, e := range apiObject {
		rawEndpoint := map[string]interface{}{
			"lambda_arn": aws.ToString(e.LambdaArn),
		}
		rawSipMediaApplicationEndpoints = append(rawSipMediaApplicationEndpoints, rawEndpoint)
	}
	return rawSipMediaApplicationEndpoints
}

func findSIPMediaApplicationByID(ctx context.Context, conn *chimesdkvoice.Client, id string) (*awstypes.SipMediaApplication, error) {
	in := &chimesdkvoice.GetSipMediaApplicationInput{
		SipMediaApplicationId: aws.String(id),
	}

	resp, err := conn.GetSipMediaApplication(ctx, in)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if resp == nil || resp.SipMediaApplication == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if err != nil {
		return nil, err
	}

	return resp.SipMediaApplication, nil
}
