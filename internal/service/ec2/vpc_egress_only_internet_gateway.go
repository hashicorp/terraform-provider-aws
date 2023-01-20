package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEgressOnlyInternetGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEgressOnlyInternetGatewayCreate,
		ReadWithoutTimeout:   resourceEgressOnlyInternetGatewayRead,
		UpdateWithoutTimeout: resourceEgressOnlyInternetGatewayUpdate,
		DeleteWithoutTimeout: resourceEgressOnlyInternetGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceEgressOnlyInternetGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateEgressOnlyInternetGatewayInput{
		ClientToken:       aws.String(resource.UniqueId()),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeEgressOnlyInternetGateway),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	output, err := conn.CreateEgressOnlyInternetGatewayWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Egress-only Internet Gateway: %s", err)
	}

	d.SetId(aws.StringValue(output.EgressOnlyInternetGateway.EgressOnlyInternetGatewayId))

	return append(diags, resourceEgressOnlyInternetGatewayRead(ctx, d, meta)...)
}

func resourceEgressOnlyInternetGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindEgressOnlyInternetGatewayByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Egress-only Internet Gateway %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Egress-only Internet Gateway (%s): %s", d.Id(), err)
	}

	ig := outputRaw.(*ec2.EgressOnlyInternetGateway)

	if len(ig.Attachments) == 1 && aws.StringValue(ig.Attachments[0].State) == ec2.AttachmentStatusAttached {
		d.Set("vpc_id", ig.Attachments[0].VpcId)
	} else {
		d.Set("vpc_id", nil)
	}

	tags := KeyValueTags(ig.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceEgressOnlyInternetGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Egress-only Internet Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEgressOnlyInternetGatewayRead(ctx, d, meta)...)
}

func resourceEgressOnlyInternetGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[INFO] Deleting EC2 Egress-only Internet Gateway: %s", d.Id())
	_, err := conn.DeleteEgressOnlyInternetGatewayWithContext(ctx, &ec2.DeleteEgressOnlyInternetGatewayInput{
		EgressOnlyInternetGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidGatewayIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Egress-only Internet Gateway (%s): %s", d.Id(), err)
	}

	return diags
}
