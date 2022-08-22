package iam

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSAMLProvider() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSAMLProviderCreate,
		ReadWithoutTimeout:   resourceSAMLProviderRead,
		UpdateWithoutTimeout: resourceSAMLProviderUpdate,
		DeleteWithoutTimeout: resourceSAMLProviderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"saml_metadata_document": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1000, 10000000),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"valid_until": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSAMLProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &iam.CreateSAMLProviderInput{
		Name:                 aws.String(name),
		SAMLMetadataDocument: aws.String(d.Get("saml_metadata_document").(string)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating IAM SAML Provider: %s", input)
	output, err := conn.CreateSAMLProvider(input)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if input.Tags != nil && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating IAM SAML Provider (%s) with tags: %s. Trying create without tags.", name, err)
		input.Tags = nil

		output, err = conn.CreateSAMLProvider(input)
	}

	if err != nil {
		return diag.Errorf("creating IAM SAML Provider (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.SAMLProviderArn))

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if input.Tags == nil && len(tags) > 0 {
		err := samlProviderUpdateTags(conn, d.Id(), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed adding tags after create for IAM SAML Provider (%s): %s", d.Id(), err)
			return resourceSAMLProviderRead(ctx, d, meta)
		}

		if err != nil {
			return diag.Errorf("adding tags after create for IAM SAML Provider (%s): %s", d.Id(), err)
		}
	}

	return resourceSAMLProviderRead(ctx, d, meta)
}

func resourceSAMLProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindSAMLProviderByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM SAML Provider %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading IAM SAML Provider (%s): %s", d.Id(), err)
	}

	name, err := nameFromSAMLProviderARN(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("arn", d.Id())
	d.Set("name", name)
	d.Set("saml_metadata_document", output.SAMLMetadataDocument)
	if output.ValidUntil != nil {
		d.Set("valid_until", aws.TimeValue(output.ValidUntil).Format(time.RFC3339))
	} else {
		d.Set("valid_until", nil)
	}

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceSAMLProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IAMConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &iam.UpdateSAMLProviderInput{
			SAMLProviderArn:      aws.String(d.Id()),
			SAMLMetadataDocument: aws.String(d.Get("saml_metadata_document").(string)),
		}

		log.Printf("[DEBUG] Updating IAM SAML Provider: %s", input)
		_, err := conn.UpdateSAMLProvider(input)

		if err != nil {
			return diag.Errorf("updating IAM SAML Provider (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := samlProviderUpdateTags(conn, d.Id(), o, n)

		// Some partitions (i.e., ISO) may not support tagging, giving error
		if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed updating tags for IAM SAML Provider (%s): %s", d.Id(), err)
			return resourceSAMLProviderRead(ctx, d, meta)
		}

		if err != nil {
			return diag.Errorf("updating tags for IAM SAML Provider (%s): %s", d.Id(), err)
		}
	}

	return resourceSAMLProviderRead(ctx, d, meta)
}

func resourceSAMLProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IAMConn

	log.Printf("[DEBUG] Deleting IAM SAML Provider: %s", d.Id())
	_, err := conn.DeleteSAMLProvider(&iam.DeleteSAMLProviderInput{
		SAMLProviderArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting IAM SAML Provider (%s): %s", d.Id(), err)
	}

	return nil
}

func nameFromSAMLProviderARN(v string) (string, error) {
	arn, err := arn.Parse(v)

	if err != nil {
		return "", fmt.Errorf("parsing IAM SAML Provider ARN (%s): %w", v, err)
	}

	return strings.TrimPrefix(arn.Resource, "saml-provider/"), nil
}
