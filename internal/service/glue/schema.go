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

func ResourceSchema() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSchemaCreate,
		ReadWithoutTimeout:   resourceSchemaRead,
		UpdateWithoutTimeout: resourceSchemaUpdate,
		DeleteWithoutTimeout: resourceSchemaDelete,
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
			"registry_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"registry_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latest_schema_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"next_schema_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"schema_checkpoint": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"compatibility": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(glue.Compatibility_Values(), false),
			},
			"data_format": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(glue.DataFormat_Values(), false),
			},
			"schema_definition": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 170000),
					validation.StringMatch(regexp.MustCompile(`.*\S.*`), ""),
				),
			},
			"schema_name": {
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

func resourceSchemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &glue.CreateSchemaInput{
		SchemaName:       aws.String(d.Get("schema_name").(string)),
		SchemaDefinition: aws.String(d.Get("schema_definition").(string)),
		DataFormat:       aws.String(d.Get("data_format").(string)),
		Tags:             Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("registry_arn"); ok {
		input.RegistryId = createRegistryID(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("compatibility"); ok {
		input.Compatibility = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Schema: %s", input)
	output, err := conn.CreateSchemaWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Schema: %s", err)
	}
	d.SetId(aws.StringValue(output.SchemaArn))

	_, err = waitSchemaAvailable(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Glue Schema (%s) to be Available: %s", d.Id(), err)
	}

	return append(diags, resourceSchemaRead(ctx, d, meta)...)
}

func resourceSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindSchemaByID(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Glue Schema (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Glue Schema (%s): %s", d.Id(), err)
	}

	if output == nil {
		log.Printf("[WARN] Glue Schema (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	arn := aws.StringValue(output.SchemaArn)
	d.Set("arn", arn)
	d.Set("description", output.Description)
	d.Set("schema_name", output.SchemaName)
	d.Set("compatibility", output.Compatibility)
	d.Set("data_format", output.DataFormat)
	d.Set("latest_schema_version", output.LatestSchemaVersion)
	d.Set("next_schema_version", output.NextSchemaVersion)
	d.Set("registry_arn", output.RegistryArn)
	d.Set("registry_name", output.RegistryName)
	d.Set("schema_checkpoint", output.SchemaCheckpoint)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Glue Schema (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	schemeDefOutput, err := FindSchemaVersionByID(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Schema Definition (%s): %s", d.Id(), err)
	}

	d.Set("schema_definition", schemeDefOutput.SchemaDefinition)

	return diags
}

func resourceSchemaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

	input := &glue.UpdateSchemaInput{
		SchemaId: createSchemaID(d.Id()),
		SchemaVersionNumber: &glue.SchemaVersionNumber{
			LatestVersion: aws.Bool(true),
		},
	}
	update := false

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
		update = true
	}

	if d.HasChange("compatibility") {
		input.Compatibility = aws.String(d.Get("compatibility").(string))
		update = true
	}

	if update {
		log.Printf("[DEBUG] Updating Glue Schema: %#v", input)
		_, err := conn.UpdateSchemaWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Schema (%s): %s", d.Id(), err)
		}

		_, err = waitSchemaAvailable(ctx, conn, d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Glue Schema (%s) to be Available: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	if d.HasChange("schema_definition") {
		defInput := &glue.RegisterSchemaVersionInput{
			SchemaId:         createSchemaID(d.Id()),
			SchemaDefinition: aws.String(d.Get("schema_definition").(string)),
		}

		_, err := conn.RegisterSchemaVersionWithContext(ctx, defInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Schema Definition (%s): %s", d.Id(), err)
		}

		_, err = waitSchemaVersionAvailable(ctx, conn, d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Glue Schema Version (%s) to be Available: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSchemaRead(ctx, d, meta)...)
}

func resourceSchemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

	log.Printf("[DEBUG] Deleting Glue Schema: %s", d.Id())
	input := &glue.DeleteSchemaInput{
		SchemaId: createSchemaID(d.Id()),
	}

	_, err := conn.DeleteSchemaWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Glue Schema (%s): %s", d.Id(), err)
	}

	_, err = waitSchemaDeleted(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for Glue Schema (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}
