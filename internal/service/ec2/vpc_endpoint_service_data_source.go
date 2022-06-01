package ec2

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceVPCEndpointService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVPCEndpointServiceRead,

		Schema: map[string]*schema.Schema{
			"acceptance_required": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zones": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},
			"base_endpoint_dns_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},
			"manages_vpc_endpoints": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"service_name"},
			},
			"service_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"service"},
			},
			"service_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ec2.ServiceType_Values(), false),
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_endpoint_policy_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
		},
	}
}

func dataSourceVPCEndpointServiceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	filters, filtersOk := d.GetOk("filter")
	tags, tagsOk := d.GetOk("tags")

	var serviceName string
	serviceNameOk := false
	if v, ok := d.GetOk("service_name"); ok {
		serviceName = v.(string)
		serviceNameOk = true
	} else if v, ok := d.GetOk("service"); ok {
		serviceName = fmt.Sprintf("com.amazonaws.%s.%s", meta.(*conns.AWSClient).Region, v.(string))
		serviceNameOk = true
	}

	req := &ec2.DescribeVpcEndpointServicesInput{}
	if filtersOk {
		req.Filters = BuildFiltersDataSource(filters.(*schema.Set))
	}
	if serviceNameOk {
		req.ServiceNames = aws.StringSlice([]string{serviceName})
	}

	if v, ok := d.GetOk("service_type"); ok {
		req.Filters = append(req.Filters, &ec2.Filter{
			Name:   aws.String("service-type"),
			Values: aws.StringSlice([]string{v.(string)}),
		})
	}

	if tagsOk {
		req.Filters = append(req.Filters, tagFiltersFromMap(tags.(map[string]interface{}))...)
	}

	log.Printf("[DEBUG] Reading VPC Endpoint Service: %s", req)
	resp, err := conn.DescribeVpcEndpointServices(req)
	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint Service (%s): %w", serviceName, err)
	}

	if resp == nil || (len(resp.ServiceNames) == 0 && len(resp.ServiceDetails) == 0) {
		return fmt.Errorf("no matching VPC Endpoint Service found")
	}

	// Note: AWS Commercial now returns a response with `ServiceNames` and
	// `ServiceDetails`, but GovCloud responses only include `ServiceNames`
	if len(resp.ServiceDetails) == 0 {
		// GovCloud doesn't respect the filter.
		names := aws.StringValueSlice(resp.ServiceNames)
		for _, name := range names {
			if name == serviceName {
				d.SetId(strconv.Itoa(create.StringHashcode(name)))
				d.Set("service_name", name)
				return nil
			}
		}

		return fmt.Errorf("no matching VPC Endpoint Service found")
	}

	if len(resp.ServiceDetails) > 1 {
		return fmt.Errorf("multiple VPC Endpoint Services matched; use additional constraints to reduce matches to a single VPC Endpoint Service")
	}

	sd := resp.ServiceDetails[0]
	serviceId := aws.StringValue(sd.ServiceId)
	serviceName = aws.StringValue(sd.ServiceName)

	d.SetId(strconv.Itoa(create.StringHashcode(serviceName)))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpc-endpoint-service/%s", serviceId),
	}.String()
	d.Set("arn", arn)

	d.Set("acceptance_required", sd.AcceptanceRequired)
	err = d.Set("availability_zones", flex.FlattenStringSet(sd.AvailabilityZones))
	if err != nil {
		return fmt.Errorf("error setting availability_zones: %w", err)
	}
	err = d.Set("base_endpoint_dns_names", flex.FlattenStringSet(sd.BaseEndpointDnsNames))
	if err != nil {
		return fmt.Errorf("error setting base_endpoint_dns_names: %w", err)
	}
	d.Set("manages_vpc_endpoints", sd.ManagesVpcEndpoints)
	d.Set("owner", sd.Owner)
	d.Set("private_dns_name", sd.PrivateDnsName)
	d.Set("service_id", serviceId)
	d.Set("service_name", serviceName)
	d.Set("service_type", sd.ServiceType[0].ServiceType)
	err = d.Set("tags", KeyValueTags(sd.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())
	if err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}
	d.Set("vpc_endpoint_policy_supported", sd.VpcEndpointPolicySupported)

	return nil
}
