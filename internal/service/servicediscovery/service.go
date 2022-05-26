package servicediscovery

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceService() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreate,
		Read:   resourceServiceRead,
		Update: resourceServiceUpdate,
		Delete: resourceServiceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dns_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_records": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ttl": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(servicediscovery.RecordType_Values(), false),
									},
								},
							},
						},
						"namespace_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"routing_policy": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      servicediscovery.RoutingPolicyMultivalue,
							ValidateFunc: validation.StringInSlice(servicediscovery.RoutingPolicy_Values(), false),
						},
					},
				},
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"health_check_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failure_threshold": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"resource_path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(servicediscovery.HealthCheckType_Values(), false),
						},
					},
				},
			},
			"health_check_custom_config": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failure_threshold": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"namespace_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceServiceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &servicediscovery.CreateServiceInput{
		CreatorRequestId: aws.String(resource.UniqueId()),
		Name:             aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dns_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DnsConfig = expandDNSConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("health_check_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.HealthCheckConfig = expandHealthCheckConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("health_check_custom_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.HealthCheckCustomConfig = expandHealthCheckCustomConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("namespace_id"); ok {
		input.NamespaceId = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Service Discovery Service: %s", input)
	output, err := conn.CreateService(input)

	if err != nil {
		return fmt.Errorf("error creating Service Discovery Service (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.Service.Id))

	return resourceServiceRead(d, meta)
}

func resourceServiceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

func resourceServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &servicediscovery.UpdateServiceInput{
			Id: aws.String(d.Id()),
			Service: &servicediscovery.ServiceChange{
				Description: aws.String(d.Get("description").(string)),
			},
		}

		if v, ok := d.GetOk("dns_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Service.DnsConfig = expandDNSConfigChange(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("health_check_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Service.HealthCheckConfig = expandHealthCheckConfig(v.([]interface{})[0].(map[string]interface{}))
		}

		output, err := conn.UpdateService(input)

		if err != nil {
			return fmt.Errorf("error updating Service Discovery Service (%s): %w", d.Id(), err)
		}

		if output != nil && output.OperationId != nil {
			if _, err := WaitOperationSuccess(conn, aws.StringValue(output.OperationId)); err != nil {
				return fmt.Errorf("error waiting for Service Discovery Service (%s) update: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Service Discovery Service (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceServiceRead(d, meta)
}

func resourceServiceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn

	if d.Get("force_destroy").(bool) {
		input := &servicediscovery.ListInstancesInput{
			ServiceId: aws.String(d.Id()),
		}

		var deletionErrs *multierror.Error

		err := conn.ListInstancesPages(input, func(page *servicediscovery.ListInstancesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, instance := range page.Instances {
				err := deregisterInstance(conn, d.Id(), aws.StringValue(instance.Id))

				if err != nil {
					log.Printf("[ERROR] %s", err)
					deletionErrs = multierror.Append(deletionErrs, err)

					continue
				}
			}

			return !lastPage
		})

		if err != nil {
			deletionErrs = multierror.Append(deletionErrs, fmt.Errorf("error listing Service Discovery Instances: %w", err))
		}

		err = deletionErrs.ErrorOrNil()

		if err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Deleting Service Discovery Service: (%s)", d.Id())
	_, err := conn.DeleteService(&servicediscovery.DeleteServiceInput{
		Id: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeServiceNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Service Discovery Service (%s): %w", d.Id(), err)
	}

	return nil
}

func expandDNSConfig(configured map[string]interface{}) *servicediscovery.DnsConfig {
	result := &servicediscovery.DnsConfig{}

	result.NamespaceId = aws.String(configured["namespace_id"].(string))
	dnsRecords := configured["dns_records"].([]interface{})
	drs := make([]*servicediscovery.DnsRecord, len(dnsRecords))
	for i := range drs {
		raw := dnsRecords[i].(map[string]interface{})
		dr := &servicediscovery.DnsRecord{
			TTL:  aws.Int64(int64(raw["ttl"].(int))),
			Type: aws.String(raw["type"].(string)),
		}
		drs[i] = dr
	}
	result.DnsRecords = drs
	if v, ok := configured["routing_policy"]; ok && v != "" {
		result.RoutingPolicy = aws.String(v.(string))
	}

	return result
}

func flattenDNSConfig(config *servicediscovery.DnsConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{}

	if config.NamespaceId != nil {
		result["namespace_id"] = aws.StringValue(config.NamespaceId)
	}
	if config.RoutingPolicy != nil {
		result["routing_policy"] = aws.StringValue(config.RoutingPolicy)
	}
	if config.DnsRecords != nil {
		drs := make([]map[string]interface{}, 0)
		for _, v := range config.DnsRecords {
			dr := map[string]interface{}{}
			dr["ttl"] = aws.Int64Value(v.TTL)
			dr["type"] = aws.StringValue(v.Type)
			drs = append(drs, dr)
		}
		result["dns_records"] = drs
	}

	if len(result) < 1 {
		return nil
	}

	return []map[string]interface{}{result}
}

func expandDNSConfigChange(configured map[string]interface{}) *servicediscovery.DnsConfigChange {
	result := &servicediscovery.DnsConfigChange{}

	dnsRecords := configured["dns_records"].([]interface{})
	drs := make([]*servicediscovery.DnsRecord, len(dnsRecords))
	for i := range drs {
		raw := dnsRecords[i].(map[string]interface{})
		dr := &servicediscovery.DnsRecord{
			TTL:  aws.Int64(int64(raw["ttl"].(int))),
			Type: aws.String(raw["type"].(string)),
		}
		drs[i] = dr
	}
	result.DnsRecords = drs

	return result
}

func expandHealthCheckConfig(configured map[string]interface{}) *servicediscovery.HealthCheckConfig {
	if len(configured) < 1 {
		return nil
	}
	result := &servicediscovery.HealthCheckConfig{}

	if v, ok := configured["failure_threshold"]; ok && v.(int) != 0 {
		result.FailureThreshold = aws.Int64(int64(v.(int)))
	}
	if v, ok := configured["resource_path"]; ok && v.(string) != "" {
		result.ResourcePath = aws.String(v.(string))
	}
	if v, ok := configured["type"]; ok && v.(string) != "" {
		result.Type = aws.String(v.(string))
	}

	return result
}

func flattenHealthCheckConfig(config *servicediscovery.HealthCheckConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}
	result := map[string]interface{}{}

	if config.FailureThreshold != nil {
		result["failure_threshold"] = aws.Int64Value(config.FailureThreshold)
	}
	if config.ResourcePath != nil {
		result["resource_path"] = aws.StringValue(config.ResourcePath)
	}
	if config.Type != nil {
		result["type"] = aws.StringValue(config.Type)
	}

	if len(result) < 1 {
		return nil
	}

	return []map[string]interface{}{result}
}

func expandHealthCheckCustomConfig(configured map[string]interface{}) *servicediscovery.HealthCheckCustomConfig {
	if len(configured) < 1 {
		return nil
	}
	result := &servicediscovery.HealthCheckCustomConfig{}

	if v, ok := configured["failure_threshold"]; ok && v.(int) != 0 {
		result.FailureThreshold = aws.Int64(int64(v.(int)))
	}

	return result
}

func flattenHealthCheckCustomConfig(config *servicediscovery.HealthCheckCustomConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}
	result := map[string]interface{}{}

	if config.FailureThreshold != nil {
		result["failure_threshold"] = aws.Int64Value(config.FailureThreshold)
	}

	if len(result) < 1 {
		return nil
	}

	return []map[string]interface{}{result}
}

func deregisterInstance(conn *servicediscovery.ServiceDiscovery, serviceID, instanceID string) error {
	input := &servicediscovery.DeregisterInstanceInput{
		InstanceId: aws.String(instanceID),
		ServiceId:  aws.String(serviceID),
	}

	log.Printf("[INFO] Deregistering Service Discovery Service (%s) Instance: %s", serviceID, instanceID)
	output, err := conn.DeregisterInstance(input)

	if err != nil {
		return fmt.Errorf("error deregistering Service Discovery Service (%s) Instance (%s): %w", serviceID, instanceID, err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := WaitOperationSuccess(conn, aws.StringValue(output.OperationId)); err != nil {
			return fmt.Errorf("error waiting for Service Discovery Service (%s) Instance (%s) deregister: %w", serviceID, instanceID, err)
		}
	}

	return nil
}
