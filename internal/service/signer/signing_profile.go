package signer

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSigningProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceSigningProfileCreate,
		Read:   resourceSigningProfileRead,
		Update: resourceSigningProfileUpdate,
		Delete: resourceSigningProfileDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"platform_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"AWSLambda-SHA384-ECDSA"},
					false),
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_]{0,64}$`), "must be alphanumeric with max length of 64 characters"),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_]{0,38}$`), "must be alphanumeric with max length of 38 characters"),
			},
			"signature_validity_period": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(signer.ValidityType_Values(), false),
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revocation_record": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"revocation_effective_from": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSigningProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SignerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	log.Printf("[DEBUG] Creating Signer signing profile")

	profileName := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	profileName = strings.Replace(profileName, "-", "_", -1)

	signingProfileInput := &signer.PutSigningProfileInput{
		ProfileName: aws.String(profileName),
		PlatformId:  aws.String(d.Get("platform_id").(string)),
	}

	if v, exists := d.GetOk("signature_validity_period"); exists {
		signatureValidityPeriod := v.([]interface{})[0].(map[string]interface{})
		signingProfileInput.SignatureValidityPeriod = &signer.SignatureValidityPeriod{
			Value: aws.Int64(int64(signatureValidityPeriod["value"].(int))),
			Type:  aws.String(signatureValidityPeriod["type"].(string)),
		}
	}

	if len(tags) > 0 {
		signingProfileInput.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := conn.PutSigningProfile(signingProfileInput)
	if err != nil {
		return fmt.Errorf("error creating Signer signing profile: %s", err)
	}

	d.SetId(profileName)

	return resourceSigningProfileRead(d, meta)
}

func resourceSigningProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SignerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	signingProfileOutput, err := conn.GetSigningProfile(&signer.GetSigningProfileInput{
		ProfileName: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, signer.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Signer Signing Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Signer signing profile (%s): %s", d.Id(), err)
	}

	if err := d.Set("platform_id", signingProfileOutput.PlatformId); err != nil {
		return fmt.Errorf("error setting signer signing profile platform id: %s", err)
	}

	if err := d.Set("signature_validity_period", []interface{}{
		map[string]interface{}{
			"value": signingProfileOutput.SignatureValidityPeriod.Value,
			"type":  signingProfileOutput.SignatureValidityPeriod.Type,
		},
	}); err != nil {
		return fmt.Errorf("error setting signer signing profile signature validity period: %s", err)
	}

	if err := d.Set("platform_display_name", signingProfileOutput.PlatformDisplayName); err != nil {
		return fmt.Errorf("error setting signer signing profile platform display name: %s", err)
	}

	if err := d.Set("name", signingProfileOutput.ProfileName); err != nil {
		return fmt.Errorf("error setting signer signing profile name: %s", err)
	}

	if err := d.Set("arn", signingProfileOutput.Arn); err != nil {
		return fmt.Errorf("error setting signer signing profile arn: %s", err)
	}

	if err := d.Set("version", signingProfileOutput.ProfileVersion); err != nil {
		return fmt.Errorf("error setting signer signing profile version: %s", err)
	}

	if err := d.Set("version_arn", signingProfileOutput.ProfileVersionArn); err != nil {
		return fmt.Errorf("error setting signer signing profile version arn: %s", err)
	}

	if err := d.Set("status", signingProfileOutput.Status); err != nil {
		return fmt.Errorf("error setting signer signing profile status: %s", err)
	}

	tags := KeyValueTags(signingProfileOutput.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if err := d.Set("revocation_record", flattenSigningProfileRevocationRecord(signingProfileOutput.RevocationRecord)); err != nil {
		return fmt.Errorf("error setting signer signing profile revocation record: %s", err)
	}

	return nil
}

func resourceSigningProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SignerConn

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Signer signing profile (%s) tags: %s", arn, err)
		}
	}

	return resourceSigningProfileRead(d, meta)
}

func resourceSigningProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SignerConn

	_, err := conn.CancelSigningProfile(&signer.CancelSigningProfileInput{
		ProfileName: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, signer.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error canceling Signer signing profile (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Signer signing profile %q canceled", d.Id())
	return nil
}

func flattenSigningProfileRevocationRecord(apiObject *signer.SigningProfileRevocationRecord) interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RevocationEffectiveFrom; v != nil {
		tfMap["revocation_effective_from"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := apiObject.RevokedAt; v != nil {
		tfMap["revoked_at"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := apiObject.RevokedBy; v != nil {
		tfMap["revoked_by"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}
