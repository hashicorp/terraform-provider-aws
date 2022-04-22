package athena

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 129),
					validation.StringMatch(regexp.MustCompile(`[\w@-]*`), ""),
				),
			},
			"parameters": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ValidateDiagFunc: allDiagFunc(
					validation.MapKeyLenBetween(1, 255),
					validation.MapValueLenBetween(0, 51200),
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(athena.DataCatalogType_Values(), false),
			},
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

	if v, ok := d.GetOk("parameters"); ok && len(v.(map[string]interface{})) > 0 {
		input.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Athena Data Catalog: %s", input)
	_, err := conn.CreateDataCatalogWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Athena Data Catalog (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceDataCatalogRead(ctx, d, meta)
}

func resourceDataCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AthenaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &athena.GetDataCatalogInput{
		Name: aws.String(d.Id()),
	}

	dataCatalog, err := conn.GetDataCatalogWithContext(ctx, input)

	// If the resource doesn't exist, the API returns a `ErrCodeInvalidRequestException` error.
	if !d.IsNewResource() && tfawserr.ErrMessageContains(err, athena.ErrCodeInvalidRequestException, "was not found") {
		log.Printf("[WARN] Athena Data Catalog (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Athena Data Catalog (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "athena",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("datacatalog/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("description", dataCatalog.DataCatalog.Description)
	d.Set("name", dataCatalog.DataCatalog.Name)
	d.Set("type", dataCatalog.DataCatalog.Type)

	// NOTE: This is a workaround for the fact that the API sets default values for parameters that are not set.
	// Because the API sets default values, what's returned by the API is different than what's set by the user.
	if v, ok := d.GetOk("parameters"); ok && len(v.(map[string]interface{})) > 0 {
		parameters := make(map[string]string, 0)

		for key, val := range v.(map[string]interface{}) {
			if v, ok := dataCatalog.DataCatalog.Parameters[key]; ok {
				parameters[key] = aws.StringValue(v)
			} else {
				parameters[key] = val.(string)
			}
		}

		d.Set("parameters", parameters)
	} else {
		d.Set("parameters", nil)
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Athena Data Catalog (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceDataCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AthenaConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &athena.UpdateDataCatalogInput{
			Name:        aws.String(d.Id()),
			Type:        aws.String(d.Get("type").(string)),
			Description: aws.String(d.Get("description").(string)),
		}

		if d.HasChange("parameters") {
			if v, ok := d.GetOk("parameters"); ok && len(v.(map[string]interface{})) > 0 {
				input.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
			}
		}

		log.Printf("[DEBUG] Updating Athena Data Catalog: %s", input)
		_, err := conn.UpdateDataCatalogWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Athena Data Catalog (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating Athena Data Catalog (%s) tags: %s", d.Id(), err)
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
		return diag.Errorf("deleting Athena Data Catalog (%s): %s", d.Id(), err)
	}

	return nil
}

func allDiagFunc(validators ...schema.SchemaValidateDiagFunc) schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		for _, validator := range validators {
			diags = append(diags, validator(i, k)...)
		}
		return diags
	}
}
