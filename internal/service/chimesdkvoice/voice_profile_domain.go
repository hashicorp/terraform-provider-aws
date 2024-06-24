// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameVoiceProfileDomain = "Voice Profile Domain"
)

// @SDKResource("aws_chimesdkvoice_voice_profile_domain", name="Voice Profile Domain")
// @Tags(identifierAttribute="arn")
func ResourceVoiceProfileDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVoiceProfileDomainCreate,
		ReadWithoutTimeout:   resourceVoiceProfileDomainRead,
		UpdateWithoutTimeout: resourceVoiceProfileDomainUpdate,
		DeleteWithoutTimeout: resourceVoiceProfileDomainDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 200),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_ .-]+`), "Name must match expression: ^[0-9A-Za-z_ .-]+"),
				),
			},
			"server_side_encryption_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVoiceProfileDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	in := &chimesdkvoice.CreateVoiceProfileDomainInput{
		Name:                              aws.String(d.Get(names.AttrName).(string)),
		ServerSideEncryptionConfiguration: expandServerSideEncryptionConfiguration(d.Get("server_side_encryption_configuration").([]interface{})),
		Tags:                              getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		in.Description = aws.String(v.(string))
	}

	out, err := conn.CreateVoiceProfileDomain(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.ChimeSDKVoice, create.ErrActionCreating, ResNameVoiceProfileDomain, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.VoiceProfileDomain == nil {
		return create.AppendDiagError(diags, names.ChimeSDKVoice, create.ErrActionCreating, ResNameVoiceProfileDomain, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.VoiceProfileDomain.VoiceProfileDomainId))

	return append(diags, resourceVoiceProfileDomainRead(ctx, d, meta)...)
}

func resourceVoiceProfileDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	out, err := FindVoiceProfileDomainByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ChimeSDKVoice VoiceProfileDomain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.ChimeSDKVoice, create.ErrActionReading, ResNameVoiceProfileDomain, d.Id(), err)
	}

	d.SetId(aws.ToString(out.VoiceProfileDomainId))
	d.Set(names.AttrARN, out.VoiceProfileDomainArn)
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrDescription, out.Description)

	if err := d.Set("server_side_encryption_configuration", flattenServerSideEncryptionConfiguration(out.ServerSideEncryptionConfiguration)); err != nil {
		return create.AppendDiagError(diags, names.ChimeSDKVoice, create.ErrActionSetting, ResNameVoiceProfileDomain, d.Id(), err)
	}

	return diags
}

func resourceVoiceProfileDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	if d.HasChanges(names.AttrName, names.AttrDescription) {
		in := &chimesdkvoice.UpdateVoiceProfileDomainInput{
			VoiceProfileDomainId: aws.String(d.Id()),
			Name:                 aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			in.Description = aws.String(v.(string))
		}

		_, err := conn.UpdateVoiceProfileDomain(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.ChimeSDKMediaPipelines, create.ErrActionUpdating, ResNameVoiceProfileDomain, d.Id(), err)
		}

		return append(diags, resourceVoiceProfileDomainRead(ctx, d, meta)...)
	}

	return diags
}

func resourceVoiceProfileDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	log.Printf("[INFO] Deleting ChimeSDKVoice VoiceProfileDomain %s", d.Id())

	_, err := conn.DeleteVoiceProfileDomain(ctx, &chimesdkvoice.DeleteVoiceProfileDomainInput{
		VoiceProfileDomainId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.ChimeSDKVoice, create.ErrActionDeleting, ResNameVoiceProfileDomain, d.Id(), err)
	}

	return diags
}

func FindVoiceProfileDomainByID(ctx context.Context, conn *chimesdkvoice.Client, id string) (*awstypes.VoiceProfileDomain, error) {
	in := &chimesdkvoice.GetVoiceProfileDomainInput{
		VoiceProfileDomainId: aws.String(id),
	}
	out, err := conn.GetVoiceProfileDomain(ctx, in)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.VoiceProfileDomain == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.VoiceProfileDomain, nil
}

func flattenServerSideEncryptionConfiguration(apiObject *awstypes.ServerSideEncryptionConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		names.AttrKMSKeyARN: apiObject.KmsKeyArn,
	}}
}

func expandServerSideEncryptionConfiguration(tfList []interface{}) *awstypes.ServerSideEncryptionConfiguration {
	if len(tfList) != 1 {
		return nil
	}
	return &awstypes.ServerSideEncryptionConfiguration{
		KmsKeyArn: aws.String(tfList[0].(map[string]interface{})[names.AttrKMSKeyARN].(string)),
	}
}
