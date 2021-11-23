package eks

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceClusterRegistration() *schema.Resource {
	return &schema.Resource{
		Create: resourceClusterRegistrationCreate,
		Read:   resourceClusterRegistrationRead,
		Update: resourceClusterRegistrionUpdate,
		Delete: resourceClusterRegistrationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"client_request_token": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validClusterName,
			},
			"connector_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(eks.ConnectorConfigProvider_Values(), false),
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceClusterRegistrationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)

	input := &eks.RegisterClusterInput{
		ClientRequestToken: aws.String(d.Get("client_request_token").(string)),
		Name:               aws.String(name),
		ConnectorConfig:    expandConnectorConfigRequest(d.Get("connector_config").([]interface{})),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating EKS Cluster Registration: %s", input)
	var output *eks.RegisterClusterOutput
	err := resource.Retry(tfiam.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.RegisterCluster(input)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.RegisterCluster(input)
	}

	if err != nil {
		return fmt.Errorf("error registering EKS Cluster (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.Cluster.Name))

	return resourceClusterRead(d, meta)
}

func expandConnectorConfigRequest(tfList []interface{}) *eks.ConnectorConfigRequest {
	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &eks.ConnectorConfigRequest{}

	if v, ok := tfMap["provider"].(string); ok && v != "" {
		apiObject.Provider = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}
