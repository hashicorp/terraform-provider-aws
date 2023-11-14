// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
						"username": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"password": {
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
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	vcId := d.Get("voice_connector_id").(string)

	input := &chimesdkvoice.PutVoiceConnectorTerminationCredentialsInput{
		VoiceConnectorId: aws.String(vcId),
		Credentials:      expandCredentials(d.Get("credentials").(*schema.Set).List()),
	}

	if _, err := conn.PutVoiceConnectorTerminationCredentialsWithContext(ctx, input); err != nil {
		return diag.Errorf("creating Chime Voice Connector (%s) termination credentials: %s", vcId, err)
	}

	d.SetId(vcId)

	return resourceVoiceConnectorTerminationCredentialsRead(ctx, d, meta)
}

func resourceVoiceConnectorTerminationCredentialsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	_, err := FindVoiceConnectorResourceWithRetry(ctx, d.IsNewResource(), func() (*chimesdkvoice.ListVoiceConnectorTerminationCredentialsOutput, error) {
		return findVoiceConnectorTerminationCredentialsByID(ctx, conn, d.Id())
	})

	if tfresource.TimedOut(err) {
		_, err = findVoiceConnectorTerminationCredentialsByID(ctx, conn, d.Id())
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Chime Voice Connector (%s) termination credentials not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("getting Chime Voice Connector (%s) termination credentials: %s", d.Id(), err)
	}

	d.Set("voice_connector_id", d.Id())

	return nil
}

func resourceVoiceConnectorTerminationCredentialsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	if d.HasChanges("credentials") {
		input := &chimesdkvoice.PutVoiceConnectorTerminationCredentialsInput{
			VoiceConnectorId: aws.String(d.Id()),
			Credentials:      expandCredentials(d.Get("credentials").(*schema.Set).List()),
		}

		_, err := conn.PutVoiceConnectorTerminationCredentialsWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Chime Voice Connector (%s) termination credentials: %s", d.Id(), err)
		}
	}

	return resourceVoiceConnectorTerminationCredentialsRead(ctx, d, meta)
}

func resourceVoiceConnectorTerminationCredentialsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	input := &chimesdkvoice.DeleteVoiceConnectorTerminationCredentialsInput{
		VoiceConnectorId: aws.String(d.Id()),
		Usernames:        expandCredentialsUsernames(d.Get("credentials").(*schema.Set).List()),
	}

	_, err := conn.DeleteVoiceConnectorTerminationCredentialsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Chime Voice Connector (%s) termination credentials: %s", d.Id(), err)
	}

	return nil
}

func expandCredentialsUsernames(data []interface{}) []*string {
	var rawNames []*string

	for _, rData := range data {
		item := rData.(map[string]interface{})
		rawNames = append(rawNames, aws.String(item["username"].(string)))
	}

	return rawNames
}

func expandCredentials(data []interface{}) []*chimesdkvoice.Credential {
	var credentials []*chimesdkvoice.Credential

	for _, rItem := range data {
		item := rItem.(map[string]interface{})
		credentials = append(credentials, &chimesdkvoice.Credential{
			Username: aws.String(item["username"].(string)),
			Password: aws.String(item["password"].(string)),
		})
	}

	return credentials
}

func findVoiceConnectorTerminationCredentialsByID(ctx context.Context, conn *chimesdkvoice.ChimeSDKVoice, id string) (*chimesdkvoice.ListVoiceConnectorTerminationCredentialsOutput, error) {
	in := &chimesdkvoice.ListVoiceConnectorTerminationCredentialsInput{
		VoiceConnectorId: aws.String(id),
	}

	resp, err := conn.ListVoiceConnectorTerminationCredentialsWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
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
