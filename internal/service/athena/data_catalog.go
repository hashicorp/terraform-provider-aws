package athena

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDataCatalog() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataCatalogCreate,
		ReadContext:   resourceDataCatalogRead,
		UpdateContext: resourceDataCatalogUpdate,
		DeleteContext: resourceDataCatalogDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 129),
					validation.StringMatch(regexp.MustCompile(`[\w@-]*`), ""),
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					athena.DataCatalogTypeLambda,
					athena.DataCatalogTypeGlue,
					athena.DataCatalogTypeHive,
				}, false),
			},
			"parameters": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					//ValidateFunc: validation.StringLenBetween(1, 51200), -- WIP
				},
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceDataCatalogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AthenaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &athena.CreateDataCatalogInput{
		Name:        aws.String(name),
		Description: aws.String(d.Get("description").(string)),
		Type:        aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if err := input.Validate(); err != nil {
		return diag.Errorf("Error validating Athena Data Catalog (%s): %s", name, err)
	}

	log.Printf("[DEBUG] Creating Data Catalog: %s", input)
	_, err := conn.CreateDataCatalogWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating Athena Data Catalog (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceDataCatalogRead(ctx, d, meta)
}

func resourceDataCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AthenaConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &athena.UpdateDataCatalogInput{
			Name: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("description") {
			input.Name = aws.String(d.Get("description").(string))
		}

		if d.HasChange("type") {
			input.Name = aws.String(d.Get("type").(string))
		}

		if d.HasChange("parameters") {
			parameters := map[string]*string{}
			if v, ok := d.GetOk("parameters"); ok {
				if m, ok := v.(map[string]interface{}); ok {
					parameters = flex.ExpandStringMap(m)
				}
			}
			input.Parameters = parameters
		}

		log.Printf("[DEBUG] Updating Athena Data Catalog (%s)", d.Id())

		if err := input.Validate(); err != nil {
			return diag.Errorf("Error validating Athena Data Catalog (%s): %s", d.Id(), err)
		}

		_, err := conn.UpdateDataCatalogWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error updating Athena Data Catalog (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		log.Printf("[DEBUG] Updating Athena Data Catalog (%s) tags", d.Id())
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating Athena Data Catalog (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceDataCatalogRead(ctx, d, meta)
}

func resourceDataCatalogDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AthenaConn

	log.Printf("[DEBUG] Deleting Athena Data Catalog: (%s)", d.Id())

	_, err := conn.DeleteDataCatalogWithContext(ctx, &athena.DeleteDataCatalogInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, athena.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Athena Data Catalog (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceDataCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AthenaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &athena.GetDataCatalogInput{
		Name: aws.String(d.Id()),
	}

	dataCatalog, err := conn.GetDataCatalogWithContext(ctx, input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Athena Data Catalog (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Athena Data Catalog (%s): %s", d.Id(), err)
	}

	d.Set("description", dataCatalog.DataCatalog.Description)
	d.Set("type", dataCatalog.DataCatalog.Type)
	d.Set("parameters", aws.StringValueMap(dataCatalog.DataCatalog.Parameters))

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return diag.Errorf("error listing tags for Athena Data Catalog (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags for Athena Data Catalog (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all for Athena Data Catalog (%s): %s", d.Id(), err)
	}

	return nil
}
