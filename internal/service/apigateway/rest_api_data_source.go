package apigateway

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceRestAPI() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRestAPIRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"root_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_key_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"minimum_compression_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"binary_media_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"endpoint_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"types": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_endpoint_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRestAPIRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &apigateway.GetRestApisInput{}

	target := d.Get("name")
	var matchedApis []*apigateway.RestApi
	log.Printf("[DEBUG] Reading API Gateway REST APIs: %s", params)
	err := conn.GetRestApisPages(params, func(page *apigateway.GetRestApisOutput, lastPage bool) bool {
		for _, api := range page.Items {
			if aws.StringValue(api.Name) == target {
				matchedApis = append(matchedApis, api)
			}
		}
		return !lastPage
	})
	if err != nil {
		return fmt.Errorf("error describing API Gateway REST APIs: %w", err)
	}

	if len(matchedApis) == 0 {
		return fmt.Errorf("no REST APIs with name %q found in this region", target)
	}
	if len(matchedApis) > 1 {
		return fmt.Errorf("multiple REST APIs with name %q found in this region", target)
	}

	match := matchedApis[0]

	d.SetId(aws.StringValue(match.Id))

	restApiArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/restapis/%s", d.Id()),
	}.String()
	d.Set("arn", restApiArn)
	d.Set("description", match.Description)
	d.Set("policy", match.Policy)
	d.Set("api_key_source", match.ApiKeySource)
	d.Set("binary_media_types", match.BinaryMediaTypes)

	if match.MinimumCompressionSize == nil {
		d.Set("minimum_compression_size", -1)
	} else {
		d.Set("minimum_compression_size", match.MinimumCompressionSize)
	}

	if err := d.Set("endpoint_configuration", flattenEndpointConfiguration(match.EndpointConfiguration)); err != nil {
		return fmt.Errorf("error setting endpoint_configuration: %w", err)
	}

	if err := d.Set("tags", KeyValueTags(match.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	executionArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "execute-api",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  d.Id(),
	}.String()
	d.Set("execution_arn", executionArn)

	resourceParams := &apigateway.GetResourcesInput{
		RestApiId: aws.String(d.Id()),
	}

	err = conn.GetResourcesPages(resourceParams, func(page *apigateway.GetResourcesOutput, lastPage bool) bool {
		for _, item := range page.Items {
			if aws.StringValue(item.Path) == "/" {
				d.Set("root_resource_id", item.Id)
				return false
			}
		}
		return !lastPage
	})

	return err
}
