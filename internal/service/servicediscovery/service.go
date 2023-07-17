// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_service_discovery_service", name="Service")
// @Tags(identifierAttribute="arn")
func ResourceService() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceCreate,
		ReadWithoutTimeout:   resourceServiceRead,
		UpdateWithoutTimeout: resourceServiceUpdate,
		DeleteWithoutTimeout: resourceServiceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(servicediscovery.ServiceTypeOption_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	name := d.Get("name").(string)
	input := &servicediscovery.CreateServiceInput{
		CreatorRequestId: aws.String(id.UniqueId()),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
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

	if v, ok := d.GetOk("type"); ok {
		input.Type = aws.String(v.(string))
	}

	output, err := conn.CreateServiceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Service Discovery Service (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Service.Id))

	return resourceServiceRead(ctx, d, meta)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	service, err := FindServiceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Service Discovery Service (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(service.Arn)
	d.Set("arn", arn)
	d.Set("description", service.Description)
	if tfMap := flattenDNSConfig(service.DnsConfig); len(tfMap) > 0 {
		if err := d.Set("dns_config", []interface{}{tfMap}); err != nil {
			return diag.Errorf("setting dns_config: %s", err)
		}
	} else {
		d.Set("dns_config", nil)
	}
	if tfMap := flattenHealthCheckConfig(service.HealthCheckConfig); len(tfMap) > 0 {
		if err := d.Set("health_check_config", []interface{}{tfMap}); err != nil {
			return diag.Errorf("setting health_check_config: %s", err)
		}
	} else {
		d.Set("health_check_config", nil)
	}
	if tfMap := flattenHealthCheckCustomConfig(service.HealthCheckCustomConfig); len(tfMap) > 0 {
		if err := d.Set("health_check_custom_config", []interface{}{tfMap}); err != nil {
			return diag.Errorf("setting health_check_custom_config: %s", err)
		}
	} else {
		d.Set("health_check_custom_config", nil)
	}
	d.Set("name", service.Name)
	d.Set("namespace_id", service.NamespaceId)
	d.Set("type", service.Type)

	return nil
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

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

		output, err := conn.UpdateServiceWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Service Discovery Service (%s): %s", d.Id(), err)
		}

		if output != nil && output.OperationId != nil {
			if _, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId)); err != nil {
				return diag.Errorf("waiting for Service Discovery Service (%s) update: %s", d.Id(), err)
			}
		}
	}

	return resourceServiceRead(ctx, d, meta)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	if d.Get("force_destroy").(bool) {
		var deletionErrs *multierror.Error
		input := &servicediscovery.ListInstancesInput{
			ServiceId: aws.String(d.Id()),
		}

		err := conn.ListInstancesPagesWithContext(ctx, input, func(page *servicediscovery.ListInstancesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, instance := range page.Instances {
				err := deregisterInstance(ctx, conn, d.Id(), aws.StringValue(instance.Id))

				if err != nil {
					log.Printf("[ERROR] %s", err)
					deletionErrs = multierror.Append(deletionErrs, err)

					continue
				}
			}

			return !lastPage
		})

		if err != nil {
			deletionErrs = multierror.Append(deletionErrs, fmt.Errorf("listing Service Discovery Instances: %w", err))
		}

		err = deletionErrs.ErrorOrNil()

		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[INFO] Deleting Service Discovery Service: %s", d.Id())
	_, err := conn.DeleteServiceWithContext(ctx, &servicediscovery.DeleteServiceInput{
		Id: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeServiceNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Service Discovery Service (%s): %s", d.Id(), err)
	}

	return nil
}

func expandDNSConfig(tfMap map[string]interface{}) *servicediscovery.DnsConfig {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &servicediscovery.DnsConfig{}

	if v, ok := tfMap["dns_records"].([]interface{}); ok && len(v) > 0 {
		apiObject.DnsRecords = expandDNSRecords(v)
	}

	if v, ok := tfMap["namespace_id"].(string); ok && v != "" {
		apiObject.NamespaceId = aws.String(v)
	}

	if v, ok := tfMap["routing_policy"].(string); ok && v != "" {
		apiObject.RoutingPolicy = aws.String(v)
	}

	return apiObject
}

func expandDNSConfigChange(tfMap map[string]interface{}) *servicediscovery.DnsConfigChange {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &servicediscovery.DnsConfigChange{}

	if v, ok := tfMap["dns_records"].([]interface{}); ok && len(v) > 0 {
		apiObject.DnsRecords = expandDNSRecords(v)
	}

	return apiObject
}

func expandDNSRecord(tfMap map[string]interface{}) *servicediscovery.DnsRecord {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &servicediscovery.DnsRecord{}

	if v, ok := tfMap["ttl"].(int); ok {
		apiObject.TTL = aws.Int64(int64(v))
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandDNSRecords(tfList []interface{}) []*servicediscovery.DnsRecord {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*servicediscovery.DnsRecord

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDNSRecord(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenDNSConfig(apiObject *servicediscovery.DnsConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DnsRecords; v != nil {
		tfMap["dns_records"] = flattenDNSRecords(v)
	}

	if v := apiObject.NamespaceId; v != nil {
		tfMap["namespace_id"] = aws.StringValue(v)
	}

	if v := apiObject.RoutingPolicy; v != nil {
		tfMap["routing_policy"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenDNSRecord(apiObject *servicediscovery.DnsRecord) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.TTL; v != nil {
		tfMap["ttl"] = aws.Int64Value(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenDNSRecords(apiObjects []*servicediscovery.DnsRecord) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDNSRecord(apiObject))
	}

	return tfList
}

func expandHealthCheckConfig(tfMap map[string]interface{}) *servicediscovery.HealthCheckConfig {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &servicediscovery.HealthCheckConfig{}

	if v, ok := tfMap["failure_threshold"].(int); ok && v != 0 {
		apiObject.FailureThreshold = aws.Int64(int64(v))
	}

	if v, ok := tfMap["resource_path"].(string); ok && v != "" {
		apiObject.ResourcePath = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func flattenHealthCheckConfig(apiObject *servicediscovery.HealthCheckConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FailureThreshold; v != nil {
		tfMap["failure_threshold"] = aws.Int64Value(v)
	}

	if v := apiObject.ResourcePath; v != nil {
		tfMap["resource_path"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func expandHealthCheckCustomConfig(tfMap map[string]interface{}) *servicediscovery.HealthCheckCustomConfig {
	if len(tfMap) < 1 {
		return nil
	}

	apiObject := &servicediscovery.HealthCheckCustomConfig{}

	if v, ok := tfMap["failure_threshold"].(int); ok && v != 0 {
		apiObject.FailureThreshold = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenHealthCheckCustomConfig(apiObject *servicediscovery.HealthCheckCustomConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FailureThreshold; v != nil {
		tfMap["failure_threshold"] = aws.Int64Value(v)
	}

	return tfMap
}

func deregisterInstance(ctx context.Context, conn *servicediscovery.ServiceDiscovery, serviceID, instanceID string) error {
	input := &servicediscovery.DeregisterInstanceInput{
		InstanceId: aws.String(instanceID),
		ServiceId:  aws.String(serviceID),
	}

	log.Printf("[INFO] Deregistering Service Discovery Service (%s) Instance: %s", serviceID, instanceID)
	output, err := conn.DeregisterInstanceWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("deregistering Service Discovery Service (%s) Instance (%s): %w", serviceID, instanceID, err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId)); err != nil {
			return fmt.Errorf("waiting for Service Discovery Service (%s) Instance (%s) delete: %w", serviceID, instanceID, err)
		}
	}

	return nil
}
