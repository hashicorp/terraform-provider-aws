package servicediscovery

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceServiceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_records": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ttl": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"namespace_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"routing_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"health_check_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failure_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"resource_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"health_check_custom_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failure_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceServiceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	input := &servicediscovery.ListServicesInput{}

	var filters []*servicediscovery.ServiceFilter

	filter := &servicediscovery.ServiceFilter{
		Condition: aws.String(servicediscovery.FilterConditionEq),
		Name:      aws.String(servicediscovery.ServiceFilterNameNamespaceId),
		Values:    []*string{aws.String(d.Get("namespace_id").(string))},
	}

	filters = append(filters, filter)

	input.Filters = filters

	serviceIds := make([]string, 0)

	err := conn.ListServicesPages(input, func(page *servicediscovery.ListServicesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, service := range page.Services {
			if service == nil {
				continue
			}

			if name == aws.StringValue(service.Name) {
				serviceIds = append(serviceIds, aws.StringValue(service.Id))
			}
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Service Discovery Services: %w", err)
	}

	if len(serviceIds) == 0 {
		return fmt.Errorf("no matching Service Discovery Service found")
	}

	if len(serviceIds) != 1 {
		return fmt.Errorf("search returned %d Service Discovery Services, please revise so only one is returned", len(serviceIds))
	}

	d.SetId(serviceIds[0])

	service, err := FindServiceByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Service Discovery Service (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(service.Arn)
	d.Set("arn", arn)
	d.Set("description", service.Description)
	if err := d.Set("dns_config", flattenDNSConfig(service.DnsConfig)); err != nil {
		return fmt.Errorf("error setting dns_config: %w", err)
	}
	if err := d.Set("health_check_config", flattenHealthCheckConfig(service.HealthCheckConfig)); err != nil {
		return fmt.Errorf("error setting health_check_config: %w", err)
	}
	if err := d.Set("health_check_custom_config", flattenHealthCheckCustomConfig(service.HealthCheckCustomConfig)); err != nil {
		return fmt.Errorf("error setting health_check_custom_config: %w", err)
	}
	d.Set("name", service.Name)
	d.Set("namespace_id", service.NamespaceId)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %w", arn, err)
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
