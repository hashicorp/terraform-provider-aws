package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
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

			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"provider_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(codestarconnections.ProviderType_Values(), false),
			},

			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAwsCodeStarConnectionsConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	params := &codestarconnections.CreateConnectionInput{
		ConnectionName: aws.String(d.Get("name").(string)),
		ProviderType:   aws.String(d.Get("provider_type").(string)),
	}

	if v, ok := d.GetOk("tags"); ok {
		params.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CodestarconnectionsTags()
	}

	resp, err := conn.CreateConnection(params)
	if err != nil {
		return fmt.Errorf("error creating CodeStar connection: %w", err)
	}

	d.SetId(aws.StringValue(resp.ConnectionArn))

	return resourceAwsCodeStarConnectionsConnectionRead(d, meta)
}

func resourceAwsCodeStarConnectionsConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetConnection(&codestarconnections.GetConnectionInput{
		ConnectionArn: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] CodeStar connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading CodeStar connection: %w", err)
	}

	if resp == nil || resp.Connection == nil {
		return fmt.Errorf("error reading CodeStar connection (%s): empty response", d.Id())
	}

	arn := aws.StringValue(resp.Connection.ConnectionArn)
	d.SetId(arn)
	d.Set("arn", resp.Connection.ConnectionArn)
	d.Set("name", resp.Connection.ConnectionName)
	d.Set("connection_status", resp.Connection.ConnectionStatus)
	d.Set("provider_type", resp.Connection.ProviderType)

	tags, err := keyvaluetags.CodestarconnectionsListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for CodeStar connection (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags for CodeStar connection (%s): %w", arn, err)
	}

	return nil
}

func resourceAwsCodeStarConnectionsConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	_, err := conn.DeleteConnection(&codestarconnections.DeleteConnectionInput{
		ConnectionArn: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting CodeStar connection: %w", err)
	}

	return nil
}
