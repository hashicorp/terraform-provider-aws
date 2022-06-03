package servicediscovery

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceHTTPNamespace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHTTPNamespaceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"http_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validNamespaceName,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceHTTPNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	input := &servicediscovery.ListNamespacesInput{}

	var filters []*servicediscovery.NamespaceFilter

	filter := &servicediscovery.NamespaceFilter{
		Condition: aws.String(servicediscovery.FilterConditionEq),
		Name:      aws.String(servicediscovery.NamespaceFilterNameType),
		Values:    []*string{aws.String("HTTP")},
	}

	filters = append(filters, filter)

	input.Filters = filters

	namespaceIds := make([]string, 0)

	err := conn.ListNamespacesPages(input, func(page *servicediscovery.ListNamespacesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, namespace := range page.Namespaces {
			if namespace == nil {
				continue
			}

			if name == aws.StringValue(namespace.Name) {
				namespaceIds = append(namespaceIds, aws.StringValue(namespace.Id))
			}
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Service Discovery Namespaces: %w", err)
	}

	if len(namespaceIds) == 0 {
		return fmt.Errorf("no matching Service Discovery Namespace found")
	}

	if len(namespaceIds) != 1 {
		return fmt.Errorf("search returned %d Service Discovery Namespaces, please revise so only one is returned", len(namespaceIds))
	}

	d.SetId(namespaceIds[0])

	resp, err := conn.GetNamespace(&servicediscovery.GetNamespaceInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeNamespaceNotFound) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Service Discovery HTTP Namespace (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(resp.Namespace.Arn)
	d.Set("arn", arn)
	d.Set("description", resp.Namespace.Description)
	d.Set("http_name", resp.Namespace.Properties.HttpProperties.HttpName)
	d.Set("name", resp.Namespace.Name)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}
