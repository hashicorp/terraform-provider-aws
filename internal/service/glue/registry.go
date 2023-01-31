package glue

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRegistry() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegistryCreate,
		ReadWithoutTimeout:   resourceRegistryRead,
		UpdateWithoutTimeout: resourceRegistryUpdate,
		DeleteWithoutTimeout: resourceRegistryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"registry_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9-_$#]+$`), ""),
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceRegistryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &glue.CreateRegistryInput{
		RegistryName: aws.String(d.Get("registry_name").(string)),
		Tags:         Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Registry: %s", input)
	output, err := conn.CreateRegistryWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Registry: %s", err)
	}
	d.SetId(aws.StringValue(output.RegistryArn))

	return append(diags, resourceRegistryRead(ctx, d, meta)...)
}

func resourceRegistryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindRegistryByID(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Glue Registry (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Glue Registry (%s): %s", d.Id(), err)
	}

	if output == nil {
		log.Printf("[WARN] Glue Registry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	arn := aws.StringValue(output.RegistryArn)
	d.Set("arn", arn)
	d.Set("description", output.Description)
	d.Set("registry_name", output.RegistryName)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Glue Registry (%s): %s", arn, err)
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

func resourceRegistryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

	if d.HasChanges("description") {
		input := &glue.UpdateRegistryInput{
			RegistryId: createRegistryID(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Glue Registry: %#v", input)
		_, err := conn.UpdateRegistryWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Registry (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceRegistryRead(ctx, d, meta)...)
}

func resourceRegistryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

	log.Printf("[DEBUG] Deleting Glue Registry: %s", d.Id())
	input := &glue.DeleteRegistryInput{
		RegistryId: createRegistryID(d.Id()),
	}

	_, err := conn.DeleteRegistryWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Glue Registry (%s): %s", d.Id(), err)
	}

	_, err = waitRegistryDeleted(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for Glue Registry (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}
