package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsChimeVoiceConnectorTerminationCredentials() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsChimeVoiceConnectorTerminationCredentialsCreate,
		ReadWithoutTimeout:   resourceAwsChimeVoiceConnectorTerminationCredentialsRead,
		UpdateWithoutTimeout: resourceAwsChimeVoiceConnectorTerminationCredentialsUpdate,
		DeleteWithoutTimeout: resourceAwsChimeVoiceConnectorTerminationCredentialsDelete,

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

func resourceAwsChimeVoiceConnectorTerminationCredentialsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)

	input := &chime.PutVoiceConnectorTerminationCredentialsInput{
		VoiceConnectorId: aws.String(vcId),
		Credentials:      expandCredentials(d.Get("credentials").(*schema.Set).List()),
	}

	if _, err := conn.PutVoiceConnectorTerminationCredentialsWithContext(ctx, input); err != nil {
		return diag.Errorf("error creating Chime Voice Connector (%s) termination credentials: %s", vcId, err)

	}

	d.SetId(vcId)

	return resourceAwsChimeVoiceConnectorTerminationCredentialsRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorTerminationCredentialsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	input := &chime.ListVoiceConnectorTerminationCredentialsInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	_, err := conn.ListVoiceConnectorTerminationCredentialsWithContext(ctx, input)
	if !d.IsNewResource() && isAWSErr(err, chime.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Chime Voice Connector (%s) termination credentials not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error getting Chime Voice Connector (%s) termination credentials: %s", d.Id(), err)
	}

	d.Set("voice_connector_id", d.Id())

	return nil
}

func resourceAwsChimeVoiceConnectorTerminationCredentialsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	if d.HasChanges("credentials") {
		input := &chime.PutVoiceConnectorTerminationCredentialsInput{
			VoiceConnectorId: aws.String(d.Id()),
			Credentials:      expandCredentials(d.Get("credentials").(*schema.Set).List()),
		}

		_, err := conn.PutVoiceConnectorTerminationCredentialsWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error updating Chime Voice Connector (%s) termination credentials: %s", d.Id(), err)
		}
	}

	return resourceAwsChimeVoiceConnectorTerminationCredentialsRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorTerminationCredentialsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	input := &chime.DeleteVoiceConnectorTerminationCredentialsInput{
		VoiceConnectorId: aws.String(d.Id()),
		Usernames:        expandCredentialsUsernames(d.Get("credentials").(*schema.Set).List()),
	}

	_, err := conn.DeleteVoiceConnectorTerminationCredentialsWithContext(ctx, input)

	if isAWSErr(err, chime.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Chime Voice Connector (%s) termination credentials: %s", d.Id(), err)
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

func expandCredentials(data []interface{}) []*chime.Credential {
	var credentials []*chime.Credential

	for _, rItem := range data {
		item := rItem.(map[string]interface{})
		credentials = append(credentials, &chime.Credential{
			Username: aws.String(item["username"].(string)),
			Password: aws.String(item["password"].(string)),
		})
	}

	return credentials
}
