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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceEmailIdentity() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEmailIdentityCreate,
		ReadWithoutTimeout:   resourceEmailIdentityRead,
		UpdateWithoutTimeout: resourceEmailIdentityUpdate,
		DeleteWithoutTimeout: resourceEmailIdentityDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_set_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"dkim_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"current_signing_key_length": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_key_generation_timestamp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"next_signing_key_length": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"signing_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"signing_attributes_origin": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tokens": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"email_identity": {
				Type:     schema.TypeString,
				Required: true,
			},
			"feedback_forwarding_status": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"identity_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mail_from_attributes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"behavior_on_mx_failure": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mail_from_domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mail_from_domain_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"verified_for_sending_status": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameEmailIdentity = "Email Identity"
)

func resourceEmailIdentityCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	in := &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(d.Get("email_identity").(string)),
	}

	if v, ok := d.GetOk("configuration_set_name"); ok {
		in.ConfigurationSetName = aws.String(v.(string))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateEmailIdentity(ctx, in)
	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionCreating, ResNameEmailIdentity, d.Get("email_identity").(string), err)
	}

	if out == nil {
		return create.DiagError(names.SESV2, create.ErrActionCreating, ResNameEmailIdentity, d.Get("email_identity").(string), errors.New("empty output"))
	}

	d.SetId(d.Get("email_identity").(string))

	return resourceEmailIdentityRead(ctx, d, meta)
}

func resourceEmailIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	out, err := FindEmailIdentityByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 EmailIdentity (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionReading, ResNameEmailIdentity, d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("identity/%s", d.Id()),
	}.String()

	d.Set("arn", arn)
	d.Set("configuration_set_name", out.ConfigurationSetName)
	d.Set("email_identity", d.Id())

	if out.DkimAttributes != nil {
		if err := d.Set("dkim_attributes", []interface{}{flattenDkimAttributes(out.DkimAttributes)}); err != nil {
			return create.DiagError(names.SESV2, create.ErrActionSetting, ResNameEmailIdentity, d.Id(), err)
		}
	} else {
		d.Set("dkim_attributes", nil)
	}

	d.Set("feedback_forwarding_status", out.FeedbackForwardingStatus)
	d.Set("identity_type", string(out.IdentityType))

	if out.MailFromAttributes != nil {
		if err := d.Set("mail_from_attributes", []interface{}{flattenMailFromAttributes(out.MailFromAttributes)}); err != nil {
			return create.DiagError(names.SESV2, create.ErrActionSetting, ResNameEmailIdentity, d.Id(), err)
		}
	} else {
		d.Set("mail_from_attributes", nil)
	}

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionReading, ResNameEmailIdentity, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.SESV2, create.ErrActionSetting, ResNameEmailIdentity, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.SESV2, create.ErrActionSetting, ResNameEmailIdentity, d.Id(), err)
	}

	d.Set("verified_for_sending_status", out.VerifiedForSendingStatus)

	return nil
}

func resourceEmailIdentityUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	if d.HasChanges("configuration_set_name") {
		in := &sesv2.PutEmailIdentityConfigurationSetAttributesInput{
			EmailIdentity: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("configuration_set_name"); ok {
			in.ConfigurationSetName = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating SESV2 EmailIdentity ConfigurationSetAttributes (%s): %#v", d.Id(), in)
		_, err := conn.PutEmailIdentityConfigurationSetAttributes(ctx, in)
		if err != nil {
			return create.DiagError(names.SESV2, create.ErrActionUpdating, ResNameEmailIdentity, d.Id(), err)
		}
	}

	if d.HasChanges("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return create.DiagError(names.SESV2, create.ErrActionUpdating, ResNameEmailIdentity, d.Id(), err)
		}
	}

	return resourceEmailIdentityRead(ctx, d, meta)
}

func resourceEmailIdentityDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	log.Printf("[INFO] Deleting SESV2 EmailIdentity %s", d.Id())

	_, err := conn.DeleteEmailIdentity(ctx, &sesv2.DeleteEmailIdentityInput{
		EmailIdentity: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.SESV2, create.ErrActionDeleting, ResNameEmailIdentity, d.Id(), err)
	}

	return nil
}

func FindEmailIdentityByID(ctx context.Context, conn *sesv2.Client, id string) (*sesv2.GetEmailIdentityOutput, error) {
	in := &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(id),
	}
	out, err := conn.GetEmailIdentity(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
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

func flattenDkimAttributes(apiObject *types.DkimAttributes) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"current_signing_key_length": string(apiObject.CurrentSigningKeyLength),
		"next_signing_key_length":    string(apiObject.NextSigningKeyLength),
		"signing_enabled":            apiObject.SigningEnabled,
		"signing_attributes_origin":  string(apiObject.SigningAttributesOrigin),
		"status":                     string(apiObject.Status),
	}

	if v := apiObject.LastKeyGenerationTimestamp; v != nil {
		m["last_key_generation_timestamp"] = v.Format(time.RFC3339)
	}

	if v := apiObject.Tokens; v != nil {
		m["tokens"] = apiObject.Tokens
	}

	return m
}

func flattenMailFromAttributes(apiObject *types.MailFromAttributes) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"behavior_on_mx_failure":  string(apiObject.BehaviorOnMxFailure),
		"mail_from_domain_status": string(apiObject.MailFromDomainStatus),
	}

	if v := apiObject.MailFromDomain; v != nil {
		m["mail_from_domain"] = aws.ToString(apiObject.MailFromDomain)
	}

	return m
}

// func dkimSigningKeyLengthValues(in []types.DkimSigningKeyLength) []string {
// 	var out []string

// 	for _, v := range in {
// 		out = append(out, string(v))
// 	}

// 	return out
// }

// func dkimSigningAttributesOriginValues(in []types.DkimSigningAttributesOrigin) []string {
// 	var out []string

// 	for _, v := range in {
// 		out = append(out, string(v))
// 	}

// 	return out
// }
