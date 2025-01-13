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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_service_discovery_service", name="Service")
// @Tags(identifierAttribute="arn")
func resourceService() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceCreate,
		ReadWithoutTimeout:   resourceServiceRead,
		UpdateWithoutTimeout: resourceServiceUpdate,
		DeleteWithoutTimeout: resourceServiceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
									names.AttrType: {
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
			names.AttrForceDestroy: {
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
						names.AttrType: {
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
							Type:       schema.TypeInt,
							Optional:   true,
							ForceNew:   true,
							Deprecated: `The attribute "failure_threshold" is now unsupported in the AWS API and is always set to 1. The attribute will be removed in a future major version.`,
						},
					},
				},
			},
			names.AttrName: {
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
			names.AttrType: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &servicediscovery.CreateServiceInput{
		CreatorRequestId: aws.String(id.UniqueId()),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
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

	if v, ok := d.GetOk(names.AttrType); ok {
		input.Type = awstypes.ServiceTypeOption(v.(string))
	}

	output, err := conn.CreateService(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Discovery Service (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Service.Id))

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	service, err := findServiceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Discovery Service (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(service.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, service.Description)
	if tfMap := flattenDNSConfig(service.DnsConfig); len(tfMap) > 0 {
		if err := d.Set("dns_config", []interface{}{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dns_config: %s", err)
		}
	} else {
		d.Set("dns_config", nil)
	}
	if tfMap := flattenHealthCheckConfig(service.HealthCheckConfig); len(tfMap) > 0 {
		if err := d.Set("health_check_config", []interface{}{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting health_check_config: %s", err)
		}
	} else {
		d.Set("health_check_config", nil)
	}
	if tfMap := flattenHealthCheckCustomConfig(service.HealthCheckCustomConfig); len(tfMap) > 0 {
		if err := d.Set("health_check_custom_config", []interface{}{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting health_check_custom_config: %s", err)
		}
	} else {
		d.Set("health_check_custom_config", nil)
	}
	d.Set(names.AttrName, service.Name)
	d.Set("namespace_id", service.NamespaceId)
	d.Set(names.AttrType, service.Type)

	return diags
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &servicediscovery.UpdateServiceInput{
			Id: aws.String(d.Id()),
			Service: &awstypes.ServiceChange{
				Description: aws.String(d.Get(names.AttrDescription).(string)),
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
			return sdkdiag.AppendErrorf(diags, "updating Service Discovery Service (%s): %s", d.Id(), err)
		}

		if output != nil && output.OperationId != nil {
			if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Service Discovery Service (%s) update: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	if d.Get(names.AttrForceDestroy).(bool) {
		var errs []error
		input := &servicediscovery.ListInstancesInput{
			ServiceId: aws.String(d.Id()),
		}

		pages := servicediscovery.NewListInstancesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil {
				errs = append(errs, fmt.Errorf("listing Service Discovery Instances: %w", err))
				break
			}

			for _, v := range page.Instances {
				err := deregisterInstance(ctx, conn, d.Id(), aws.ToString(v.Id))

				if err != nil {
					errs = append(errs, err)
					continue
				}
			}
		}

		if err := errors.Join(errs...); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting Service Discovery Service: %s", d.Id())
	_, err := conn.DeleteService(ctx, &servicediscovery.DeleteServiceInput{
		Id: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ServiceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Discovery Service (%s): %s", d.Id(), err)
	}

	return diags
}

func findService(ctx context.Context, conn *servicediscovery.Client, input *servicediscovery.ListServicesInput, filter tfslices.Predicate[*awstypes.ServiceSummary]) (*awstypes.ServiceSummary, error) {
	output, err := findServices(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findServices(ctx context.Context, conn *servicediscovery.Client, input *servicediscovery.ListServicesInput, filter tfslices.Predicate[*awstypes.ServiceSummary]) ([]awstypes.ServiceSummary, error) {
	var output []awstypes.ServiceSummary

	pages := servicediscovery.NewListServicesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Services {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findServiceByNameAndNamespaceID(ctx context.Context, conn *servicediscovery.Client, name, namespaceID string) (*awstypes.ServiceSummary, error) {
	input := &servicediscovery.ListServicesInput{
		Filters: []awstypes.ServiceFilter{{
			Condition: awstypes.FilterConditionEq,
			Name:      awstypes.ServiceFilterNameNamespaceId,
			Values:    []string{namespaceID},
		}},
	}

	return findService(ctx, conn, input, func(v *awstypes.ServiceSummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findServiceByID(ctx context.Context, conn *servicediscovery.Client, id string) (*awstypes.Service, error) {
	input := &servicediscovery.GetServiceInput{
		Id: aws.String(id),
	}

	output, err := conn.GetService(ctx, input)

	if errs.IsA[*awstypes.ServiceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Service == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Service, nil
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

func expandDNSRecord(tfMap map[string]interface{}) *awstypes.DnsRecord {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &awstypes.DnsRecord{}

	if v, ok := tfMap["ttl"].(int); ok {
		apiObject.TTL = aws.Int64(int64(v))
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
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

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
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

	if v := apiObject.RoutingPolicy; v != "" {
		tfMap["routing_policy"] = v
	}

	return tfMap
}

func flattenDNSRecord(apiObject *awstypes.DnsRecord) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.TTL; v != nil {
		tfMap["ttl"] = aws.ToInt64(v)
	}

	if v := apiObject.Type; v != "" {
		tfMap[names.AttrType] = v
	}

	return tfMap
}

func flattenDNSRecords(apiObjects []awstypes.DnsRecord) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenDNSRecord(&apiObject))
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

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
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

	if v := apiObject.Type; v != "" {
		tfMap[names.AttrType] = v
	}

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
