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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVoiceProfileDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn()

	in := &chimesdkvoice.CreateVoiceProfileDomainInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if description, ok := d.GetOk("description"); ok {
		in.Description = aws.String(description.(string))
	}

	if serverSideEncryptionConfiguration, ok := d.GetOk("server_side_encryption_configuration"); ok && len(serverSideEncryptionConfiguration.([]interface{})) > 0 {
		in.ServerSideEncryptionConfiguration = expandServerSideEncryptionConfiguration(serverSideEncryptionConfiguration.([]interface{}))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
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
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn()

	out, err := findVoiceProfileDomainByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ChimeSDKVoice VoiceProfileDomain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.ChimeSDKVoice, create.ErrActionReading, ResNameVoiceProfileDomain, d.Id(), err)
	}

	d.Set("arn", out.VoiceProfileDomainArn)
	d.Set("id", out.VoiceProfileDomainId)
	d.Set("name", out.Name)
	d.Set("description", out.Description)

	if err := d.Set("server_side_encryption_configuration", flattenServerSideEncryptionConfiguration(out.ServerSideEncryptionConfiguration)); err != nil {
		return create.DiagError(names.ChimeSDKVoice, create.ErrActionSetting, ResNameVoiceProfileDomain, d.Id(), err)
	}

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return create.DiagError(names.ChimeSDKVoice, create.ErrActionReading, ResNameVoiceProfileDomain, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.ChimeSDKVoice, create.ErrActionSetting, ResNameVoiceProfileDomain, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.ChimeSDKVoice, create.ErrActionSetting, ResNameVoiceProfileDomain, d.Id(), err)
	}

	return nil
}

func resourceVoiceProfileDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn()

	if d.HasChanges(names.AttrName, names.AttrDescription) {
		in := &chimesdkvoice.UpdateVoiceProfileDomainInput{
			VoiceProfileDomainId: aws.String(d.Id()),
			Name:                 aws.String(d.Get("name").(string)),
		}

		if description, ok := d.GetOk("description"); ok {
			in.Description = aws.String(description.(string))
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
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn()

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

func findVoiceProfileDomainByID(ctx context.Context, conn *chimesdkvoice.ChimeSDKVoice, id string) (*chimesdkvoice.VoiceProfileDomain, error) {
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
