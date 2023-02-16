package iam

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOpenIDConnectProvider() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOpenIDConnectProviderCreate,
		ReadWithoutTimeout:   resourceOpenIDConnectProviderRead,
		UpdateWithoutTimeout: resourceOpenIDConnectProviderUpdate,
		DeleteWithoutTimeout: resourceOpenIDConnectProviderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceOpenIDConnectProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &iam.CreateOpenIDConnectProviderInput{
		Url:            aws.String(d.Get("url").(string)),
		ClientIDList:   flex.ExpandStringList(d.Get("client_id_list").([]interface{})),
		ThumbprintList: flex.ExpandStringList(d.Get("thumbprint_list").([]interface{})),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateOpenIDConnectProviderWithContext(ctx, input)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if input.Tags != nil && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating IAM OIDC Provider with tags: %s. Trying create without tags.", err)
		input.Tags = nil

		out, err = conn.CreateOpenIDConnectProviderWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM OIDC Provider: %s", err)
	}

	d.SetId(aws.StringValue(out.OpenIDConnectProviderArn))

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if input.Tags == nil && len(tags) > 0 {
		err := openIDConnectProviderUpdateTags(ctx, conn, d.Id(), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed adding tags after create for IAM OIDC Provider (%s): %s", d.Id(), err)
			return append(diags, resourceOpenIDConnectProviderRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "adding tags after create for IAM OIDC Provider (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceOpenIDConnectProviderRead(ctx, d, meta)...)
}

func resourceOpenIDConnectProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &iam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(d.Id()),
	}
	out, err := conn.GetOpenIDConnectProviderWithContext(ctx, input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM OIDC Provider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM OIDC Provider (%s): %s", d.Id(), err)
	}

	d.Set("arn", d.Id())
	d.Set("url", out.Url)
	d.Set("client_id_list", flex.FlattenStringList(out.ClientIDList))
	d.Set("thumbprint_list", flex.FlattenStringList(out.ThumbprintList))

	tags := KeyValueTags(out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceOpenIDConnectProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	if d.HasChange("thumbprint_list") {
		input := &iam.UpdateOpenIDConnectProviderThumbprintInput{
			OpenIDConnectProviderArn: aws.String(d.Id()),
			ThumbprintList:           flex.ExpandStringList(d.Get("thumbprint_list").([]interface{})),
		}

		_, err := conn.UpdateOpenIDConnectProviderThumbprintWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM OIDC Provider (%s) thumbprint: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := openIDConnectProviderUpdateTags(ctx, conn, d.Id(), o, n)

		// Some partitions (i.e., ISO) may not support tagging, giving error
		if verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed updating tags for IAM OIDC Provider (%s): %s", d.Id(), err)
			return append(diags, resourceOpenIDConnectProviderRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags for IAM OIDC Provider (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceOpenIDConnectProviderRead(ctx, d, meta)...)
}

func resourceOpenIDConnectProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	input := &iam.DeleteOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(d.Id()),
	}
	_, err := conn.DeleteOpenIDConnectProviderWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM OIDC Provider (%s): %s", d.Id(), err)
	}

	return diags
}
