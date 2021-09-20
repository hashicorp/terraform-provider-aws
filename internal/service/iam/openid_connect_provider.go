package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOpenIDConnectProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpenIDConnectProviderCreate,
		Read:   resourceOpenIDConnectProviderRead,
		Update: resourceOpenIDConnectProviderUpdate,
		Delete: resourceOpenIDConnectProviderDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     validOpenIDURL,
				DiffSuppressFunc: suppressOpenIDURL,
			},
			"client_id_list": {
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
			},
			"thumbprint_list": {
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(40, 40),
				},
				Type:     schema.TypeList,
				Required: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOpenIDConnectProviderCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &iam.CreateOpenIDConnectProviderInput{
		Url:            aws.String(d.Get("url").(string)),
		ClientIDList:   flex.ExpandStringList(d.Get("client_id_list").([]interface{})),
		ThumbprintList: flex.ExpandStringList(d.Get("thumbprint_list").([]interface{})),
		Tags:           Tags(tags.IgnoreAws()),
	}

	out, err := conn.CreateOpenIDConnectProvider(input)
	if err != nil {
		return fmt.Errorf("error creating IAM OIDC Provider: %w", err)
	}

	d.SetId(aws.StringValue(out.OpenIDConnectProviderArn))

	return resourceOpenIDConnectProviderRead(d, meta)
}

func resourceOpenIDConnectProviderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &iam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(d.Id()),
	}
	out, err := conn.GetOpenIDConnectProvider(input)
	if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
		log.Printf("[WARN] IAM OIDC Provider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading IAM OIDC Provider (%s): %w", d.Id(), err)
	}

	d.Set("arn", d.Id())
	d.Set("url", out.Url)
	d.Set("client_id_list", flex.FlattenStringList(out.ClientIDList))
	d.Set("thumbprint_list", flex.FlattenStringList(out.ThumbprintList))

	tags := KeyValueTags(out.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceOpenIDConnectProviderUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	if d.HasChange("thumbprint_list") {
		input := &iam.UpdateOpenIDConnectProviderThumbprintInput{
			OpenIDConnectProviderArn: aws.String(d.Id()),
			ThumbprintList:           flex.ExpandStringList(d.Get("thumbprint_list").([]interface{})),
		}

		_, err := conn.UpdateOpenIDConnectProviderThumbprint(input)
		if err != nil {
			return fmt.Errorf("error updating IAM OIDC Provider (%s) thumbprint: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := openIDConnectProviderUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags for IAM OIDC Provider (%s): %w", d.Id(), err)
		}
	}

	return resourceOpenIDConnectProviderRead(d, meta)
}

func resourceOpenIDConnectProviderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	input := &iam.DeleteOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(d.Id()),
	}
	_, err := conn.DeleteOpenIDConnectProvider(input)
	if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting IAM OIDC Provider (%s): %w", d.Id(), err)
	}

	return nil
}
