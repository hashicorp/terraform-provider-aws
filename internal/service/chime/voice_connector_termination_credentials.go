// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_chime_voice_connector_termination_credentials")
func ResourceVoiceConnectorTerminationCredentials() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVoiceConnectorTerminationCredentialsCreate,
		ReadWithoutTimeout:   resourceVoiceConnectorTerminationCredentialsRead,
		UpdateWithoutTimeout: resourceVoiceConnectorTerminationCredentialsUpdate,
		DeleteWithoutTimeout: resourceVoiceConnectorTerminationCredentialsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"credentials": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrUsername: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						names.AttrPassword: {
							Type:         schema.TypeString,
							Required:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
			"voice_connector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVoiceConnectorTerminationCredentialsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	vcId := d.Get("voice_connector_id").(string)

	input := &chimesdkvoice.PutVoiceConnectorTerminationCredentialsInput{
		VoiceConnectorId: aws.String(vcId),
		Credentials:      expandCredentials(d.Get("credentials").(*schema.Set).List()),
	}

	if _, err := conn.PutVoiceConnectorTerminationCredentials(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Chime Voice Connector (%s) termination credentials: %s", vcId, err)
	}

	d.SetId(vcId)

	return append(diags, resourceVoiceConnectorTerminationCredentialsRead(ctx, d, meta)...)
}

func resourceVoiceConnectorTerminationCredentialsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	_, err := FindVoiceConnectorResourceWithRetry(ctx, d.IsNewResource(), func() (*chimesdkvoice.ListVoiceConnectorTerminationCredentialsOutput, error) {
		return findVoiceConnectorTerminationCredentialsByID(ctx, conn, d.Id())
	})

	if tfresource.TimedOut(err) {
		_, err = findVoiceConnectorTerminationCredentialsByID(ctx, conn, d.Id())
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Chime Voice Connector (%s) termination credentials not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Chime Voice Connector (%s) termination credentials: %s", d.Id(), err)
	}

	d.Set("voice_connector_id", d.Id())

	return diags
}

func resourceVoiceConnectorTerminationCredentialsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	if d.HasChanges("credentials") {
		input := &chimesdkvoice.PutVoiceConnectorTerminationCredentialsInput{
			VoiceConnectorId: aws.String(d.Id()),
			Credentials:      expandCredentials(d.Get("credentials").(*schema.Set).List()),
		}

		_, err := conn.PutVoiceConnectorTerminationCredentials(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Chime Voice Connector (%s) termination credentials: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVoiceConnectorTerminationCredentialsRead(ctx, d, meta)...)
}

func resourceVoiceConnectorTerminationCredentialsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.DeleteVoiceConnectorTerminationCredentialsInput{
		VoiceConnectorId: aws.String(d.Id()),
		Usernames:        expandCredentialsUsernames(d.Get("credentials").(*schema.Set).List()),
	}

	_, err := conn.DeleteVoiceConnectorTerminationCredentials(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Chime Voice Connector (%s) termination credentials: %s", d.Id(), err)
	}

	return diags
}

func expandCredentialsUsernames(data []interface{}) []string {
	var rawNames []string

	for _, rData := range data {
		item := rData.(map[string]interface{})
		rawNames = append(rawNames, item[names.AttrUsername].(string))
	}

	return rawNames
}

func expandCredentials(data []interface{}) []awstypes.Credential {
	var credentials []awstypes.Credential

	for _, rItem := range data {
		item := rItem.(map[string]interface{})
		credentials = append(credentials, awstypes.Credential{
			Username: aws.String(item[names.AttrUsername].(string)),
			Password: aws.String(item[names.AttrPassword].(string)),
		})
	}

	return credentials
}

func findVoiceConnectorTerminationCredentialsByID(ctx context.Context, conn *chimesdkvoice.Client, id string) (*chimesdkvoice.ListVoiceConnectorTerminationCredentialsOutput, error) {
	in := &chimesdkvoice.ListVoiceConnectorTerminationCredentialsInput{
		VoiceConnectorId: aws.String(id),
	}

	resp, err := conn.ListVoiceConnectorTerminationCredentials(ctx, in)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if resp == nil || resp.Usernames == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}
