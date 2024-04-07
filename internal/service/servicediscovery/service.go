// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.RecordType](),
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
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          awstypes.RoutingPolicyMultivalue,
							ValidateDiagFunc: enum.Validate[awstypes.RoutingPolicy](),
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
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.HealthCheckType](),
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ServiceTypeOption](),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

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
		input.Type = awstypes.ServiceTypeOption(v.(string))
	}

	output, err := conn.CreateService(ctx, input)

	if err != nil {
		return diag.Errorf("creating Service Discovery Service (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Service.Id))

	return resourceServiceRead(ctx, d, meta)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	service, err := FindServiceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Service Discovery Service (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(service.Arn)
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
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &servicediscovery.UpdateServiceInput{
			Id: aws.String(d.Id()),
			Service: &awstypes.ServiceChange{
				Description: aws.String(d.Get("description").(string)),
			},
		}

		if v, ok := d.GetOk("dns_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Service.DnsConfig = expandDNSConfigChange(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("health_check_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Service.HealthCheckConfig = expandHealthCheckConfig(v.([]interface{})[0].(map[string]interface{}))
		}

		output, err := conn.UpdateService(ctx, input)

		if err != nil {
			return diag.Errorf("updating Service Discovery Service (%s): %s", d.Id(), err)
		}

		if output != nil && output.OperationId != nil {
			if _, err := WaitOperationSuccess(ctx, conn, aws.ToString(output.OperationId)); err != nil {
				return diag.Errorf("waiting for Service Discovery Service (%s) update: %s", d.Id(), err)
			}
		}
	}

	return resourceServiceRead(ctx, d, meta)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	if d.Get("force_destroy").(bool) {
		var errs []error
		input := &servicediscovery.ListInstancesInput{
			ServiceId: aws.String(d.Id()),
		}

		pages := servicediscovery.NewListInstancesPaginator(conn, input)

		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil {
				errs = append(errs, fmt.Errorf("listing Service Discovery Instances: %w", err))
			}

			for _, instance := range page.Instances {
				err := deregisterInstance(ctx, conn, d.Id(), aws.ToString(instance.Id))

				if err != nil {
					errs = append(errs, err)

					continue
				}
			}
		}

		err := errors.Join(errs...)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[INFO] Deleting Service Discovery Service: %s", d.Id())
	_, err := conn.DeleteService(ctx, &servicediscovery.DeleteServiceInput{
		Id: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ServiceNotFound](err) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Service Discovery Service (%s): %s", d.Id(), err)
	}

	return nil
}

func expandDNSConfig(tfMap map[string]interface{}) *awstypes.DnsConfig {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &awstypes.DnsConfig{}

	if v, ok := tfMap["dns_records"].([]interface{}); ok && len(v) > 0 {
		apiObject.DnsRecords = expandDNSRecords(v)
	}

	if v, ok := tfMap["namespace_id"].(string); ok && v != "" {
		apiObject.NamespaceId = aws.String(v)
	}

	if v, ok := tfMap["routing_policy"].(string); ok && v != "" {
		apiObject.RoutingPolicy = awstypes.RoutingPolicy(v)
	}

	return apiObject
}

func expandDNSConfigChange(tfMap map[string]interface{}) *awstypes.DnsConfigChange {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &awstypes.DnsConfigChange{}

	if v, ok := tfMap["dns_records"].([]interface{}); ok && len(v) > 0 {
		apiObject.DnsRecords = expandDNSRecords(v)
	}

	return apiObject
}

func expandDNSRecord(tfMap map[string]interface{}) awstypes.DnsRecord {
	apiObject := awstypes.DnsRecord{}

	if v, ok := tfMap["ttl"].(int); ok {
		apiObject.TTL = aws.Int64(int64(v))
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = awstypes.RecordType(v)
	}

	return apiObject
}

func expandDNSRecords(tfList []interface{}) []awstypes.DnsRecord {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DnsRecord

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDNSRecord(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenDNSConfig(apiObject *awstypes.DnsConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DnsRecords; v != nil {
		tfMap["dns_records"] = flattenDNSRecords(v)
	}

	if v := apiObject.NamespaceId; v != nil {
		tfMap["namespace_id"] = aws.ToString(v)
	}

	tfMap["routing_policy"] = string(apiObject.RoutingPolicy)

	return tfMap
}

func flattenDNSRecord(apiObject awstypes.DnsRecord) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.TTL; v != nil {
		tfMap["ttl"] = aws.ToInt64(v)
	}

	tfMap["type"] = string(apiObject.Type)

	return tfMap
}

func flattenDNSRecords(apiObjects []awstypes.DnsRecord) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenDNSRecord(apiObject))
	}

	return tfList
}

func expandHealthCheckConfig(tfMap map[string]interface{}) *awstypes.HealthCheckConfig {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &awstypes.HealthCheckConfig{}

	if v, ok := tfMap["failure_threshold"].(int); ok && v != 0 {
		apiObject.FailureThreshold = aws.Int32(int32(v))
	}

	if v, ok := tfMap["resource_path"].(string); ok && v != "" {
		apiObject.ResourcePath = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = awstypes.HealthCheckType(v)
	}

	return apiObject
}

func flattenHealthCheckConfig(apiObject *awstypes.HealthCheckConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FailureThreshold; v != nil {
		tfMap["failure_threshold"] = aws.ToInt32(v)
	}

	if v := apiObject.ResourcePath; v != nil {
		tfMap["resource_path"] = aws.ToString(v)
	}

	tfMap["type"] = string(apiObject.Type)

	return tfMap
}

func expandHealthCheckCustomConfig(tfMap map[string]interface{}) *awstypes.HealthCheckCustomConfig {
	if len(tfMap) < 1 {
		return nil
	}

	apiObject := &awstypes.HealthCheckCustomConfig{}

	if v, ok := tfMap["failure_threshold"].(int); ok && v != 0 {
		apiObject.FailureThreshold = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenHealthCheckCustomConfig(apiObject *awstypes.HealthCheckCustomConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FailureThreshold; v != nil {
		tfMap["failure_threshold"] = aws.ToInt32(v)
	}

	return tfMap
}

func deregisterInstance(ctx context.Context, conn *servicediscovery.Client, serviceID, instanceID string) error {
	input := &servicediscovery.DeregisterInstanceInput{
		InstanceId: aws.String(instanceID),
		ServiceId:  aws.String(serviceID),
	}

	log.Printf("[INFO] Deregistering Service Discovery Service (%s) Instance: %s", serviceID, instanceID)
	output, err := conn.DeregisterInstance(ctx, input)

	if err != nil {
		return fmt.Errorf("deregistering Service Discovery Service (%s) Instance (%s): %w", serviceID, instanceID, err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := WaitOperationSuccess(ctx, conn, aws.ToString(output.OperationId)); err != nil {
			return fmt.Errorf("waiting for Service Discovery Service (%s) Instance (%s) delete: %w", serviceID, instanceID, err)
		}
	}

	return nil
}
