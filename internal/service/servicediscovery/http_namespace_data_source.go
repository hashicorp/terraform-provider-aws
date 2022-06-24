package servicediscovery

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
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
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	nsSummary, err := findNamespaceByNameAndType(conn, name, servicediscovery.NamespaceTypeHttp)

	if err != nil {
		return fmt.Errorf("reading Service Discovery HTTP Namespace (%s): %w", name, err)
	}

	namespaceID := aws.StringValue(nsSummary.Id)

	ns, err := FindNamespaceByID(conn, namespaceID)

	if err != nil {
		return fmt.Errorf("reading Service Discovery HTTP Namespace (%s): %w", namespaceID, err)
	}

	d.SetId(namespaceID)
	arn := aws.StringValue(ns.Arn)
	d.Set("arn", arn)
	d.Set("description", ns.Description)
	d.Set("http_name", ns.Properties.HttpProperties.HttpName)
	d.Set("name", ns.Name)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("listing tags for Service Discovery HTTP Namespace (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	return nil
}
