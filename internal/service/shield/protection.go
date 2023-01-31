package shield

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceProtection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProtectionCreate,
		UpdateWithoutTimeout: resourceProtectionUpdate,
		ReadWithoutTimeout:   resourceProtectionRead,
		DeleteWithoutTimeout: resourceProtectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProtectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceProtectionRead(ctx, d, meta)...)
}

func resourceProtectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldConn()

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &shield.CreateProtectionInput{
		Name:        aws.String(d.Get("name").(string)),
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
		Tags:        Tags(tags.IgnoreAWS()),
	}

	resp, err := conn.CreateProtectionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Shield Protection: %s", err)
	}
	d.SetId(aws.StringValue(resp.ProtectionId))
	return append(diags, resourceProtectionRead(ctx, d, meta)...)
}

func resourceProtectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &shield.DescribeProtectionInput{
		ProtectionId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeProtectionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Shield Protection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Shield Protection (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(resp.Protection.ProtectionArn)
	d.Set("arn", arn)
	d.Set("name", resp.Protection.Name)
	d.Set("resource_arn", resp.Protection.ResourceArn)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Shield Protection (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceProtectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldConn()

	input := &shield.DeleteProtectionInput{
		ProtectionId: aws.String(d.Id()),
	}

	_, err := conn.DeleteProtectionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Shield Protection (%s): %s", d.Id(), err)
	}
	return diags
}
