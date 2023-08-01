// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
					validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9 _.-]+`), "Name must match expression: ^[0-9a-zA-Z._-]+"),
				),
			},
			"server_side_encryption_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_arn": {
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
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	in := &chimesdkvoice.CreateVoiceProfileDomainInput{
		Name:                              aws.String(d.Get(names.AttrName).(string)),
		ServerSideEncryptionConfiguration: expandServerSideEncryptionConfiguration(d.Get("server_side_encryption_configuration").([]interface{})),
		Tags:                              getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		in.Description = aws.String(v.(string))
	}

	out, err := conn.CreateVoiceProfileDomainWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.ChimeSDKVoice, create.ErrActionCreating, ResNameVoiceProfileDomain, d.Get("name").(string), err)
	}

	if out == nil || out.VoiceProfileDomain == nil {
		return create.DiagError(names.ChimeSDKVoice, create.ErrActionCreating, ResNameVoiceProfileDomain, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.VoiceProfileDomain.VoiceProfileDomainId))

	return resourceVoiceProfileDomainRead(ctx, d, meta)
}

func resourceVoiceProfileDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	out, err := FindVoiceProfileDomainByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ChimeSDKVoice VoiceProfileDomain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.ChimeSDKVoice, create.ErrActionReading, ResNameVoiceProfileDomain, d.Id(), err)
	}

	d.SetId(aws.StringValue(out.VoiceProfileDomainId))
	d.Set(names.AttrARN, out.VoiceProfileDomainArn)
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrDescription, out.Description)

	if err := d.Set("server_side_encryption_configuration", flattenServerSideEncryptionConfiguration(out.ServerSideEncryptionConfiguration)); err != nil {
		return create.DiagError(names.ChimeSDKVoice, create.ErrActionSetting, ResNameVoiceProfileDomain, d.Id(), err)
	}

	return nil
}

func resourceVoiceProfileDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	if d.HasChanges(names.AttrName, names.AttrDescription) {
		in := &chimesdkvoice.UpdateVoiceProfileDomainInput{
			VoiceProfileDomainId: aws.String(d.Id()),
			Name:                 aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			in.Description = aws.String(v.(string))
		}

		_, err := conn.UpdateVoiceProfileDomainWithContext(ctx, in)
		if err != nil {
			return create.DiagError(names.ChimeSDKMediaPipelines, create.ErrActionUpdating, ResNameVoiceProfileDomain, d.Id(), err)
		}

		return resourceVoiceProfileDomainRead(ctx, d, meta)
	}

	return nil
}

func resourceVoiceProfileDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	log.Printf("[INFO] Deleting ChimeSDKVoice VoiceProfileDomain %s", d.Id())

	_, err := conn.DeleteVoiceProfileDomainWithContext(ctx, &chimesdkvoice.DeleteVoiceProfileDomainInput{
		VoiceProfileDomainId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.ChimeSDKVoice, create.ErrActionDeleting, ResNameVoiceProfileDomain, d.Id(), err)
	}

	return nil
}

func FindVoiceProfileDomainByID(ctx context.Context, conn *chimesdkvoice.ChimeSDKVoice, id string) (*chimesdkvoice.VoiceProfileDomain, error) {
	in := &chimesdkvoice.GetVoiceProfileDomainInput{
		VoiceProfileDomainId: aws.String(id),
	}
	out, err := conn.GetVoiceProfileDomainWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
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

func flattenServerSideEncryptionConfiguration(apiObject *chimesdkvoice.ServerSideEncryptionConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"kms_key_arn": apiObject.KmsKeyArn,
	}}
}

func expandServerSideEncryptionConfiguration(tfList []interface{}) *chimesdkvoice.ServerSideEncryptionConfiguration {
	if len(tfList) != 1 {
		return nil
	}
	return &chimesdkvoice.ServerSideEncryptionConfiguration{
		KmsKeyArn: aws.String(tfList[0].(map[string]interface{})["kms_key_arn"].(string)),
	}
}
