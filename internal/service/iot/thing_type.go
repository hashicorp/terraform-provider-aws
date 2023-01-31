package iot

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
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

// https://docs.aws.amazon.com/iot/latest/apireference/API_CreateThingType.html
func ResourceThingType() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingTypeCreate,
		ReadWithoutTimeout:   resourceThingTypeRead,
		UpdateWithoutTimeout: resourceThingTypeUpdate,
		DeleteWithoutTimeout: resourceThingTypeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validThingTypeName,
			},
			"properties": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validThingTypeDescription,
						},
						"searchable_attributes": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 3,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validThingTypeSearchableAttribute,
							},
						},
					},
				},
			},
			"deprecated": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceThingTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	params := &iot.CreateThingTypeInput{
		ThingTypeName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("properties"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.ThingTypeProperties = expandThingTypeProperties(config)
		}
	}
	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating IoT Thing Type: %s", params)
	out, err := conn.CreateThingTypeWithContext(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Thing Type (%s): %s", d.Get("name").(string), err)
	}

	d.SetId(aws.StringValue(out.ThingTypeName))

	if v := d.Get("deprecated").(bool); v {
		params := &iot.DeprecateThingTypeInput{
			ThingTypeName: aws.String(d.Id()),
			UndoDeprecate: aws.Bool(false),
		}

		_, err := conn.DeprecateThingTypeWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IoT Thing Type (%s): deprecating Thing Type: %s", d.Get("name").(string), err)
		}
	}

	return append(diags, resourceThingTypeRead(ctx, d, meta)...)
}

func resourceThingTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &iot.DescribeThingTypeInput{
		ThingTypeName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Thing Type: %s", params)
	out, err := conn.DescribeThingTypeWithContext(ctx, params)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] IoT Thing Type (%s) not found, removing from state", d.Id())
			d.SetId("")
		}
		return sdkdiag.AppendErrorf(diags, "reading IoT Thing Type (%s): %s", d.Id(), err)
	}

	if out.ThingTypeMetadata != nil {
		d.Set("deprecated", out.ThingTypeMetadata.Deprecated)
	}

	d.Set("arn", out.ThingTypeArn)

	tags, err := ListTags(ctx, conn, aws.StringValue(out.ThingTypeArn))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for IoT Thing Type (%s): %s", aws.StringValue(out.ThingTypeArn), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	if err := d.Set("properties", flattenThingTypeProperties(out.ThingTypeProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting properties: %s", err)
	}

	return diags
}

func resourceThingTypeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	if d.HasChange("deprecated") {
		params := &iot.DeprecateThingTypeInput{
			ThingTypeName: aws.String(d.Id()),
			UndoDeprecate: aws.Bool(!d.Get("deprecated").(bool)),
		}

		log.Printf("[DEBUG] Updating IoT Thing Type: %s", params)
		_, err := conn.DeprecateThingTypeWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT Thing Type (%s): deprecating Thing Type: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceThingTypeRead(ctx, d, meta)...)
}

func resourceThingTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	// In order to delete an IoT Thing Type, you must deprecate it first and wait
	// at least 5 minutes.
	deprecateParams := &iot.DeprecateThingTypeInput{
		ThingTypeName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deprecating IoT Thing Type: %s", deprecateParams)
	_, err := conn.DeprecateThingTypeWithContext(ctx, deprecateParams)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing Type (%s): deprecating Thing Type: %s", d.Id(), err)
	}

	deleteParams := &iot.DeleteThingTypeInput{
		ThingTypeName: aws.String(d.Id()),
	}

	err = resource.RetryContext(ctx, 6*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteThingTypeWithContext(ctx, deleteParams)

		if err != nil {
			if tfawserr.ErrMessageContains(err, iot.ErrCodeInvalidRequestException, "Please wait for 5 minutes after deprecation and then retry") {
				return resource.RetryableError(err)
			}

			// As the delay post-deprecation is about 5 minutes, it may have been
			// deleted in between, thus getting a Not Found Exception.
			if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
				return nil
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteThingTypeWithContext(ctx, deleteParams)
		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			return diags
		}
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing Type (%s): %s", d.Id(), err)
	}
	return diags
}
