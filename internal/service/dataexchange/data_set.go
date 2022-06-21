package dataexchange

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dataexchange"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDataSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceDataSetCreate,
		Read:   resourceDataSetRead,
		Update: resourceDataSetUpdate,
		Delete: resourceDataSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(dataexchange.AssetType_Values(), false),
			},
			"description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 16348),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDataSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataExchangeConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &dataexchange.CreateDataSetInput{
		Name:        aws.String(d.Get("name").(string)),
		AssetType:   aws.String(d.Get("asset_type").(string)),
		Description: aws.String(d.Get("description").(string)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateDataSet(input)
	if err != nil {
		return fmt.Errorf("Error creating DataExchange DataSet: %w", err)
	}

	d.SetId(aws.StringValue(out.Id))

	return resourceDataSetRead(d, meta)
}

func resourceDataSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataExchangeConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dataSet, err := FindDataSetById(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataExchange DataSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataExchange DataSet (%s): %w", d.Id(), err)
	}

	d.Set("asset_type", dataSet.AssetType)
	d.Set("name", dataSet.Name)
	d.Set("description", dataSet.Description)
	d.Set("arn", dataSet.Arn)

	tags := KeyValueTags(dataSet.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceDataSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataExchangeConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &dataexchange.UpdateDataSetInput{
			DataSetId: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		log.Printf("[DEBUG] Updating DataExchange DataSet: %s", d.Id())
		_, err := conn.UpdateDataSet(input)
		if err != nil {
			return fmt.Errorf("Error Updating DataExchange DataSet: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DataExchange DataSet (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceDataSetRead(d, meta)
}

func resourceDataSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataExchangeConn

	input := &dataexchange.DeleteDataSetInput{
		DataSetId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataExchange DataSet: %s", d.Id())
	_, err := conn.DeleteDataSet(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, dataexchange.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting DataExchange DataSet: %w", err)
	}

	return nil
}
