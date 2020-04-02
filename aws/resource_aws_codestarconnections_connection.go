package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsCodeStarConnectionsConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeStarConnectionsConnectionCreate,
		Read:   resourceAwsCodeStarConnectionsConnectionRead,
		Delete: resourceAwsCodeStarConnectionsConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"connection_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"connection_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"provider_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					codestarconnections.ProviderTypeBitbucket,
				}, false),
			},
		},
	}
}

func resourceAwsCodeStarConnectionsConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	params := &codestarconnections.CreateConnectionInput{
		ConnectionName: aws.String(d.Get("connection_name").(string)),
		ProviderType:   aws.String(d.Get("provider_type").(string)),
	}

	res, err := conn.CreateConnection(params)
	if err != nil {
		return fmt.Errorf("error creating codestar connection: %s", err)
	}

	d.SetId(aws.StringValue(res.ConnectionArn))

	return resourceAwsCodeStarConnectionsConnectionRead(d, meta)
}

func resourceAwsCodeStarConnectionsConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	rule, err := conn.GetConnection(&codestarconnections.GetConnectionInput{
		ConnectionArn: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, codestarconnections.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] codestar connection (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading codestar connection: %s", err)
	}

	d.SetId(aws.StringValue(rule.Connection.ConnectionArn))
	d.Set("arn", rule.Connection.ConnectionArn)
	d.Set("connection_arn", rule.Connection.ConnectionArn)
	d.Set("connection_name", rule.Connection.ConnectionName)
	d.Set("connection_status", rule.Connection.ConnectionStatus)
	d.Set("provider_type", rule.Connection.ProviderType)

	return nil
}

func resourceAwsCodeStarConnectionsConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	_, err := conn.DeleteConnection(&codestarconnections.DeleteConnectionInput{
		ConnectionArn: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting codestar connection: %s", err)
	}

	return nil
}
