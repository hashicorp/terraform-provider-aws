package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsTimestreamWriteDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsTimestreamWriteDatabaseCreate,
		Read:   resourceAwsTimestreamWriteDatabaseRead,
		Update: resourceAwsTimestreamWriteDatabaseUpdate,
		Delete: resourceAwsTimestreamWriteDatabaseDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsTimestreamWriteDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).timestreamwriteconn

	input := &timestreamwrite.CreateDatabaseInput{
		DatabaseName: aws.String(d.Get("database_name").(string)),
	}
	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if attr, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(attr.(map[string]interface{})).IgnoreAws().TimestreamwriteTags()
	}

	_, err := conn.CreateDatabase(input)
	if err != nil {
		return err
	}

	d.SetId(d.Get("database_name").(string))

	return resourceAwsTimestreamWriteDatabaseRead(d, meta)
}

func resourceAwsTimestreamWriteDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).timestreamwriteconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeDatabase(&timestreamwrite.DescribeDatabaseInput{
		DatabaseName: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, timestreamwrite.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Timestream Database %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("database_name", resp.Database.DatabaseName)
	d.Set("kms_key_id", resp.Database.KmsKeyId)
	d.Set("arn", resp.Database.Arn)

	arn := aws.StringValue(resp.Database.Arn)

	tags, err := keyvaluetags.TimestreamwriteListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Timestream Database (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsTimestreamWriteDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).timestreamwriteconn

	if d.HasChange("kms_key_id") {
		input := &timestreamwrite.UpdateDatabaseInput{
			DatabaseName: aws.String(d.Id()),
			KmsKeyId:     aws.String(d.Get("kms_key_id").(string)),
		}

		_, err := conn.UpdateDatabase(input)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.TimestreamwriteUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Timesteram Database (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsTimestreamWriteDatabaseRead(d, meta)
}

func resourceAwsTimestreamWriteDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).timestreamwriteconn

	input := &timestreamwrite.DeleteDatabaseInput{
		DatabaseName: aws.String(d.Id()),
	}

	_, err := conn.DeleteDatabase(input)
	if err != nil {
		if isAWSErr(err, timestreamwrite.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}
