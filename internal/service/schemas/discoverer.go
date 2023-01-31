package schemas

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDiscoverer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDiscovererCreate,
		ReadWithoutTimeout:   resourceDiscovererRead,
		UpdateWithoutTimeout: resourceDiscovererUpdate,
		DeleteWithoutTimeout: resourceDiscovererDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},

			"source_arn": {
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

func resourceDiscovererCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	sourceARN := d.Get("source_arn").(string)
	input := &schemas.CreateDiscovererInput{
		SourceArn: aws.String(sourceARN),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating EventBridge Schemas Discoverer: %s", input)
	output, err := conn.CreateDiscovererWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Schemas Discoverer (%s): %s", sourceARN, err)
	}

	d.SetId(aws.StringValue(output.DiscovererId))

	return append(diags, resourceDiscovererRead(ctx, d, meta)...)
}

func resourceDiscovererRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindDiscovererByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Discoverer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Schemas Discoverer (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.DiscovererArn)
	d.Set("description", output.Description)
	d.Set("source_arn", output.SourceArn)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for EventBridge Schemas Discoverer (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceDiscovererUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn()

	if d.HasChange("description") {
		input := &schemas.UpdateDiscovererInput{
			DiscovererId: aws.String(d.Id()),
			Description:  aws.String(d.Get("description").(string)),
		}

		log.Printf("[DEBUG] Updating EventBridge Schemas Discoverer: %s", input)
		_, err := conn.UpdateDiscovererWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EventBridge Schemas Discoverer (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceDiscovererRead(ctx, d, meta)...)
}

func resourceDiscovererDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn()

	log.Printf("[INFO] Deleting EventBridge Schemas Discoverer (%s)", d.Id())
	_, err := conn.DeleteDiscovererWithContext(ctx, &schemas.DeleteDiscovererInput{
		DiscovererId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Schemas Discoverer (%s): %s", d.Id(), err)
	}

	return diags
}
