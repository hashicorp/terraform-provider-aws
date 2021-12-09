package eks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
		CreateWithoutTimeout: resourceClusterRegistrationCreate,
		ReadWithoutTimeout:   resourceClusterRegistrationRead,
		DeleteWithoutTimeout: resourceClusterRegistrationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceClusterRegistrationImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			"connector_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
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
						"activation_code": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"activation_expiry": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchemaForceNew(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceClusterRegistrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)

	input := &eks.RegisterClusterInput{
		Name:            aws.String(name),
		ConnectorConfig: expandConnectorConfigRequest(d.Get("connector_config").([]interface{})),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating EKS Cluster Registration: %s", input)

	var output *eks.RegisterClusterOutput

	err := resource.Retry(tfiam.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.RegisterCluster(input)

		// InvalidRequestException: Not existing role: arn:aws:iam::12345678:role/xxx
		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidRequestException, "Not existing role") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.RegisterCluster(input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error registering EKS Cluster (%s): %w", name, err))
	}

	d.SetId(aws.StringValue(output.Cluster.Name))

	_, err = waitClusterRegistrationPending(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {

		return diag.FromErr(fmt.Errorf("unexpected EKS Cluster Registration (%s) state returned during creation: %s", d.Id(), err))
	}

	return resourceClusterRegistrationRead(ctx, d, meta)
}

func resourceClusterRegistrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	cluster, err := FindClusterByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Cluster-Registration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading EKS Add-On (%s): %w", d.Id(), err))
	}

	d.Set("arn", cluster.Arn)

	if err := d.Set("connector_config", flattenConnectorConfig(cluster.ConnectorConfig)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting connector config: %w", err))
	}

	d.Set("created_at", aws.TimeValue(cluster.CreatedAt).String())

	d.Set("name", cluster.Name)
	d.Set("status", cluster.Status)

	tags := KeyValueTags(cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceClusterRegistrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn

	log.Printf("[DEBUG] Deleting EKS Cluster Registration: %s", d.Id())

	_, err := conn.DeregisterClusterWithContext(ctx, &eks.DeregisterClusterInput{
		Name: aws.String(d.Id()),
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting EKS Cluster Registration (%s): %w", d.Id(), err))
	}

	return nil
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

func flattenConnectorConfig(apiObject *eks.ConnectorConfigResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"provider":          aws.StringValue(apiObject.Provider),
		"role_arn":          aws.StringValue(apiObject.RoleArn),
		"activation_code":   aws.StringValue(apiObject.ActivationCode),
		"activation_expiry": aws.TimeValue(apiObject.ActivationExpiry).Format(time.RFC3339),
	}

	return []interface{}{tfMap}
}

func resourceClusterRegistrationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).EKSConn

	cluster, err := FindClusterByName(conn, d.Id())
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, fmt.Errorf("EKS cluster (%s) not found", d.Id())
	}

	if connectorProvider := aws.StringValue(cluster.ConnectorConfig.Provider); connectorProvider == "" {
		return nil, fmt.Errorf("EKS cluster (%s) has not been registered", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
