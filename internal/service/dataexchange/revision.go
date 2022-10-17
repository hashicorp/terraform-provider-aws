package dataexchange

import (
	"fmt"
	"log"
	"strings"

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

func ResourceRevision() *schema.Resource {
	return &schema.Resource{
		Create: resourceRevisionCreate,
		Read:   resourceRevisionRead,
		Update: resourceRevisionUpdate,
		Delete: resourceRevisionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"comment": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16348),
			},
			"data_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"revision_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRevisionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataExchangeConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &dataexchange.CreateRevisionInput{
		DataSetId: aws.String(d.Get("data_set_id").(string)),
		Comment:   aws.String(d.Get("comment").(string)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateRevision(input)
	if err != nil {
		return fmt.Errorf("Error creating DataExchange Revision: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(out.DataSetId), aws.StringValue(out.Id)))

	return resourceRevisionRead(d, meta)
}

func resourceRevisionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataExchangeConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dataSetId, revisionId, err := RevisionParseResourceID(d.Id())
	if err != nil {
		return err
	}

	revision, err := FindRevisionById(conn, dataSetId, revisionId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataExchange Revision (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataExchange Revision (%s): %w", d.Id(), err)
	}

	d.Set("data_set_id", revision.DataSetId)
	d.Set("comment", revision.Comment)
	d.Set("arn", revision.Arn)
	d.Set("revision_id", revision.Id)

	tags := KeyValueTags(revision.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceRevisionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataExchangeConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &dataexchange.UpdateRevisionInput{
			RevisionId: aws.String(d.Get("revision_id").(string)),
			DataSetId:  aws.String(d.Get("data_set_id").(string)),
		}

		if d.HasChange("comment") {
			input.Comment = aws.String(d.Get("comment").(string))
		}

		log.Printf("[DEBUG] Updating DataExchange Revision: %s", d.Id())
		_, err := conn.UpdateRevision(input)
		if err != nil {
			return fmt.Errorf("Error Updating DataExchange Revision: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DataExchange Revision (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceRevisionRead(d, meta)
}

func resourceRevisionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataExchangeConn

	input := &dataexchange.DeleteRevisionInput{
		RevisionId: aws.String(d.Get("revision_id").(string)),
		DataSetId:  aws.String(d.Get("data_set_id").(string)),
	}

	log.Printf("[DEBUG] Deleting DataExchange Revision: %s", d.Id())
	_, err := conn.DeleteRevision(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, dataexchange.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting DataExchange Revision: %w", err)
	}

	return nil
}

func RevisionParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%s), expected DATA-SET_ID:REVISION-ID", id)
}
