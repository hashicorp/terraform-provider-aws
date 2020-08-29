package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func dataSourceServiceDiscoveryDnsNamespace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceServiceDiscoveryDnsNamespaceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					servicediscovery.NamespaceTypeDnsPublic,
					servicediscovery.NamespaceTypeDnsPrivate,
				}, false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceServiceDiscoveryDnsNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn

	name := d.Get("name").(string)
	input := &servicediscovery.ListNamespacesInput{}

	var filters []*servicediscovery.NamespaceFilter

	filter := &servicediscovery.NamespaceFilter{
		Condition: aws.String("EQ"),
		Name:      aws.String("TYPE"),
		Values:    []*string{aws.String(d.Get("type").(string))},
	}

	filters = append(filters, filter)

	input.Filters = filters

	namespaceIds := make([]string, 0)
	if err := conn.ListNamespacesPages(input, func(res *servicediscovery.ListNamespacesOutput, lastPage bool) bool {
		for _, namespace := range res.Namespaces {
			if name == aws.StringValue(namespace.Name) {
				namespaceIds = append(namespaceIds, aws.StringValue(namespace.Id))
			}
		}
		return !lastPage
	}); err != nil {
		return err
	}

	if namespaceIds == nil || len(namespaceIds) == 0 {
		return fmt.Errorf("no matching Namespace found")
	}
	if len(namespaceIds) > 1 {
		return fmt.Errorf("multiple Namespaces matched; use additional constraints to reduce matches to a single Namespace")
	}

	d.SetId(namespaceIds[0])

	req := &servicediscovery.GetNamespaceInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetNamespace(req)
	if err != nil {
		return err
	}

	d.Set("name", resp.Namespace.Name)
	d.Set("description", resp.Namespace.Description)
	d.Set("arn", resp.Namespace.Arn)
	if resp.Namespace.Properties != nil {
		d.Set("hosted_zone", resp.Namespace.Properties.DnsProperties.HostedZoneId)
	}
	return nil

}
