package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
)

func resourceAwsQuickSightDataSource() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsQuickSightDataSourceCreate,
		Read:   resourceAwsQuickSightDataSourceRead,
		Update: resourceAwsQuickSightDataSourceUpdate,
		Delete: resourceAwsQuickSightDataSourceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			// OK
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// OK
			"aws_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			// TODO
			"credentials": {
			},

			// TODO
			"id": {
			},

			// TODO
			"name": {
			},

			// TODO
			"parameters": {
			},

			// TODO
			"permissions": {
			},

			// TODO
			"ssl_properties": {
			},

			// TODO
			"tags": {
			},

			// TODO
			"type": {
			},

			// TODO
			"vpc_connection_properties": {
			},

			// TODO: this is a tough one...
			"creation_stauts": {
				Type:     schema.TypeString,
			},
		},
	}
}

func resourceAwsQuickSightDataSourceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID := meta.(*AWSClient).accountid
	namespace := d.Get("namespace").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountID = v.(string)
	}

	createOpts := &quicksight.CreateDataSourceInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		DataSourceName:    aws.String(d.Get("group_name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		createOpts.Description = aws.String(v.(string))
	}

	resp, err := conn.CreateDataSource(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating QuickSight DataSource: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", awsAccountID, namespace, aws.StringValue(resp.DataSource.DataSourceName)))

	return resourceAwsQuickSightDataSourceRead(d, meta)
}

func resourceAwsQuickSightDataSourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, groupName, err := resourceAwsQuickSightDataSourceParseID(d.Id())
	if err != nil {
		return err
	}

	descOpts := &quicksight.DescribeDataSourceInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		DataSourceName:    aws.String(groupName),
	}

	resp, err := conn.DescribeDataSource(descOpts)
	if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] QuickSight DataSource %s is already gone", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing QuickSight DataSource (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.DataSource.Arn)
	d.Set("aws_account_id", awsAccountID)
	d.Set("group_name", resp.DataSource.DataSourceName)
	d.Set("description", resp.DataSource.Description)
	d.Set("namespace", namespace)

	return nil
}

func resourceAwsQuickSightDataSourceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, groupName, err := resourceAwsQuickSightDataSourceParseID(d.Id())
	if err != nil {
		return err
	}

	updateOpts := &quicksight.UpdateDataSourceInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		DataSourceName:    aws.String(groupName),
	}

	if v, ok := d.GetOk("description"); ok {
		updateOpts.Description = aws.String(v.(string))
	}

	_, err = conn.UpdateDataSource(updateOpts)
	if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] QuickSight DataSource %s is already gone", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error updating QuickSight DataSource %s: %s", d.Id(), err)
	}

	return resourceAwsQuickSightDataSourceRead(d, meta)
}

func resourceAwsQuickSightDataSourceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, groupName, err := resourceAwsQuickSightDataSourceParseID(d.Id())
	if err != nil {
		return err
	}

	deleteOpts := &quicksight.DeleteDataSourceInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		DataSourceName:    aws.String(groupName),
	}

	if _, err := conn.DeleteDataSource(deleteOpts); err != nil {
		if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting QuickSight DataSource %s: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsQuickSightDataSourceParseID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/DATA_SOURCE_ID", id)
	}
	return parts[0], parts[1], parts[2], nil
}
