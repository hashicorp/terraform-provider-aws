package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAPIKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAPIKeyCreate,
		Read:   resourceAPIKeyRead,
		Update: resourceAPIKeyUpdate,
		Delete: resourceAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"value": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(30, 128),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAPIKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	log.Printf("[DEBUG] Creating API Gateway API Key")

	apiKey, err := conn.CreateApiKey(&apigateway.CreateApiKeyInput{
		Name:        aws.String(d.Get("name").(string)),
		Description: aws.String(d.Get("description").(string)),
		Enabled:     aws.Bool(d.Get("enabled").(bool)),
		Value:       aws.String(d.Get("value").(string)),
		Tags:        tags.IgnoreAws().ApigatewayTags(),
	})
	if err != nil {
		return fmt.Errorf("Error creating API Gateway API Key: %s", err)
	}

	d.SetId(aws.StringValue(apiKey.Id))

	return resourceAPIKeyRead(d, meta)
}

func resourceAPIKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading API Gateway API Key: %s", d.Id())

	apiKey, err := conn.GetApiKey(&apigateway.GetApiKeyInput{
		ApiKey:       aws.String(d.Id()),
		IncludeValue: aws.Bool(true),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] API Gateway API Key (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	tags := tftags.ApigatewayKeyValueTags(apiKey.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/apikeys/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	d.Set("name", apiKey.Name)
	d.Set("description", apiKey.Description)
	d.Set("enabled", apiKey.Enabled)
	d.Set("value", apiKey.Value)

	if err := d.Set("created_date", apiKey.CreatedDate.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting created_date: %s", err)
	}

	if err := d.Set("last_updated_date", apiKey.LastUpdatedDate.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting last_updated_date: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayApiKeyUpdateOperations(d *schema.ResourceData) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)
	if d.HasChange("enabled") {
		isEnabled := "false"
		if d.Get("enabled").(bool) {
			isEnabled = "true"
		}
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/enabled"),
			Value: aws.String(isEnabled),
		})
	}

	if d.HasChange("description") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}

	return operations
}

func resourceAPIKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	log.Printf("[DEBUG] Updating API Gateway API Key: %s", d.Id())

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.ApigatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	_, err := conn.UpdateApiKey(&apigateway.UpdateApiKeyInput{
		ApiKey:          aws.String(d.Id()),
		PatchOperations: resourceAwsApiGatewayApiKeyUpdateOperations(d),
	})
	if err != nil {
		return err
	}

	return resourceAPIKeyRead(d, meta)
}

func resourceAPIKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	log.Printf("[DEBUG] Deleting API Gateway API Key: %s", d.Id())

	_, err := conn.DeleteApiKey(&apigateway.DeleteApiKeyInput{
		ApiKey: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting API Gateway API Key (%s): %s", d.Id(), err)
	}

	return nil
}
