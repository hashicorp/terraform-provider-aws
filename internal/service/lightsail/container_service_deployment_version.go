// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_container_service_deployment_version")
func ResourceContainerServiceDeploymentVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContainerServiceDeploymentVersionCreate,
		ReadWithoutTimeout:   resourceContainerServiceDeploymentVersionRead,
		DeleteWithoutTimeout: resourceContainerServiceDeploymentVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"container": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MaxItems: 53,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"image": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"command": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						names.AttrEnvironment: {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ports": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(flattenContainerServiceProtocolValues(types.ContainerServiceProtocol("").Values()), false),
							}},
					},
				},
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_endpoint": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"container_port": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						names.AttrHealthCheck: {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"healthy_threshold": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
										Default:  2,
									},
									"interval_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										Default:      5,
										ValidateFunc: validation.IntBetween(5, 300),
									},
									names.AttrPath: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Default:  "/",
									},
									"success_codes": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Default:  "200-499",
									},
									"timeout_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										Default:      2,
										ValidateFunc: validation.IntBetween(2, 60),
									},
									"unhealthy_threshold": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
										Default:  2,
									},
								},
							},
						},
					},
				},
			},
			names.AttrServiceName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceContainerServiceDeploymentVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	serviceName := d.Get(names.AttrServiceName).(string)

	input := lightsail.CreateContainerServiceDeploymentInput{
		ServiceName: aws.String(serviceName),
	}

	if v, ok := d.GetOk("container"); ok && v.(*schema.Set).Len() > 0 {
		input.Containers = expandContainerServiceDeploymentContainers(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("public_endpoint"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.PublicEndpoint = expandContainerServiceDeploymentPublicEndpoint(v.([]interface{}))
	}

	output, err := conn.CreateContainerServiceDeployment(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lightsail Container Service (%s) Deployment Version: %s", serviceName, err)
	}

	if output == nil || output.ContainerService == nil || output.ContainerService.NextDeployment == nil {
		return sdkdiag.AppendErrorf(diags, "creating Lightsail Container Service (%s) Deployment Version: empty output", serviceName)
	}

	version := int(aws.ToInt32(output.ContainerService.NextDeployment.Version))

	d.SetId(fmt.Sprintf("%s/%d", serviceName, version))

	if err := waitContainerServiceDeploymentVersionActive(ctx, conn, serviceName, version, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lightsail Container Service (%s) Deployment Version (%d): %s", serviceName, version, err)
	}

	return append(diags, resourceContainerServiceDeploymentVersionRead(ctx, d, meta)...)
}

func resourceContainerServiceDeploymentVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	serviceName, version, err := ContainerServiceDeploymentVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	deployment, err := FindContainerServiceDeploymentByVersion(ctx, conn, serviceName, version)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lightsail Container Service (%s) Deployment Version (%d) not found, removing from state", serviceName, version)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lightsail Container Service (%s) Deployment Version (%d): %s", serviceName, version, err)
	}

	d.Set(names.AttrCreatedAt, aws.ToTime(deployment.CreatedAt).Format(time.RFC3339))
	d.Set(names.AttrServiceName, serviceName)
	d.Set(names.AttrState, deployment.State)
	d.Set(names.AttrVersion, deployment.Version)

	if err := d.Set("container", flattenContainerServiceDeploymentContainers(deployment.Containers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting container for Lightsail Container Service (%s) Deployment Version (%d): %s", serviceName, version, err)
	}

	if err := d.Set("public_endpoint", flattenContainerServiceDeploymentPublicEndpoint(deployment.PublicEndpoint)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting public_endpoint for Lightsail Container Service (%s) Deployment Version (%d): %s", serviceName, version, err)
	}

	return diags
}

func resourceContainerServiceDeploymentVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[WARN] Cannot destroy Lightsail Container Service Deployment Version. Terraform will remove this resource from the state file, however resources may remain.")
	return nil // nosemgrep:ci.semgrep.pluginsdk.return-diags-not-nil
}

func ContainerServiceDeploymentVersionParseResourceID(id string) (string, int, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", 0, fmt.Errorf("unexpected format for ID (%[1]s), expected SERVICE_NAME/VERSION", id)
	}

	version, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, err
	}

	return parts[0], version, nil
}

func expandContainerServiceDeploymentContainers(tfList []interface{}) map[string]types.Container {
	if len(tfList) == 0 {
		return map[string]types.Container{}
	}

	result := make(map[string]types.Container)

	for _, tfListRaw := range tfList {
		tfMap, ok := tfListRaw.(map[string]interface{})
		if !ok {
			continue
		}

		containerName := tfMap["container_name"].(string)

		container := types.Container{
			Image: aws.String(tfMap["image"].(string)),
		}

		if v, ok := tfMap["command"].([]interface{}); ok && len(v) > 0 {
			container.Command = aws.ToStringSlice(flex.ExpandStringList(v))
		}

		if v, ok := tfMap[names.AttrEnvironment].(map[string]interface{}); ok && len(v) > 0 {
			container.Environment = aws.ToStringMap(flex.ExpandStringMap(v))
		}

		if v, ok := tfMap["ports"].(map[string]interface{}); ok && len(v) > 0 {
			container.Ports = expandContainerServiceProtocol(v)
		}

		result[containerName] = container
	}

	return result
}

func expandContainerServiceProtocol(tfMap map[string]interface{}) map[string]types.ContainerServiceProtocol {
	if tfMap == nil {
		return nil
	}

	apiObject := map[string]types.ContainerServiceProtocol{}

	for k, v := range tfMap {
		switch v {
		case "HTTP":
			apiObject[k] = types.ContainerServiceProtocolHttp
		case "HTTPS":
			apiObject[k] = types.ContainerServiceProtocolHttps
		case "TCP":
			apiObject[k] = types.ContainerServiceProtocolTcp
		case "UDP":
			apiObject[k] = types.ContainerServiceProtocolUdp
		}
	}

	return apiObject
}

func expandContainerServiceDeploymentPublicEndpoint(tfList []interface{}) *types.EndpointRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	endpoint := &types.EndpointRequest{
		ContainerName: aws.String(tfMap["container_name"].(string)),
		ContainerPort: aws.Int32(int32(tfMap["container_port"].(int))),
	}

	if v, ok := tfMap[names.AttrHealthCheck].([]interface{}); ok && len(v) > 0 {
		endpoint.HealthCheck = expandContainerServiceDeploymentPublicEndpointHealthCheck(v)
	}

	return endpoint
}

func expandContainerServiceDeploymentPublicEndpointHealthCheck(tfList []interface{}) *types.ContainerServiceHealthCheckConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	healthCheck := &types.ContainerServiceHealthCheckConfig{
		HealthyThreshold:   aws.Int32(int32(tfMap["healthy_threshold"].(int))),
		IntervalSeconds:    aws.Int32(int32(tfMap["interval_seconds"].(int))),
		Path:               aws.String(tfMap[names.AttrPath].(string)),
		SuccessCodes:       aws.String(tfMap["success_codes"].(string)),
		TimeoutSeconds:     aws.Int32(int32(tfMap["timeout_seconds"].(int))),
		UnhealthyThreshold: aws.Int32(int32(tfMap["unhealthy_threshold"].(int))),
	}

	return healthCheck
}

func flattenContainerServiceDeploymentContainers(containers map[string]types.Container) []interface{} {
	if len(containers) == 0 {
		return nil
	}

	var rawContainers []interface{}
	for containerName, container := range containers {
		rawContainer := map[string]interface{}{
			"container_name":      containerName,
			"image":               aws.ToString(container.Image),
			"command":             container.Command,
			names.AttrEnvironment: container.Environment,
			"ports":               container.Ports,
		}

		rawContainers = append(rawContainers, rawContainer)
	}

	return rawContainers
}

func flattenContainerServiceDeploymentPublicEndpoint(endpoint *types.ContainerServiceEndpoint) []interface{} {
	if endpoint == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"container_name":      aws.ToString(endpoint.ContainerName),
			"container_port":      int(aws.ToInt32(endpoint.ContainerPort)),
			names.AttrHealthCheck: flattenContainerServiceDeploymentPublicEndpointHealthCheck(endpoint.HealthCheck),
		},
	}
}

func flattenContainerServiceDeploymentPublicEndpointHealthCheck(healthCheck *types.ContainerServiceHealthCheckConfig) []interface{} {
	if healthCheck == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"healthy_threshold":   int(aws.ToInt32(healthCheck.HealthyThreshold)),
			"interval_seconds":    int(aws.ToInt32(healthCheck.IntervalSeconds)),
			names.AttrPath:        aws.ToString(healthCheck.Path),
			"success_codes":       aws.ToString(healthCheck.SuccessCodes),
			"timeout_seconds":     int(aws.ToInt32(healthCheck.TimeoutSeconds)),
			"unhealthy_threshold": int(aws.ToInt32(healthCheck.UnhealthyThreshold)),
		},
	}
}

func flattenContainerServiceProtocolValues(t []types.ContainerServiceProtocol) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func FindContainerServiceDeploymentByVersion(ctx context.Context, conn *lightsail.Client, serviceName string, version int) (*types.ContainerServiceDeployment, error) {
	input := &lightsail.GetContainerServiceDeploymentsInput{
		ServiceName: aws.String(serviceName),
	}

	output, err := conn.GetContainerServiceDeployments(ctx, input)

	if IsANotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Deployments) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	var result types.ContainerServiceDeployment

	for _, deployment := range output.Deployments {
		if reflect.DeepEqual(deployment, types.ContainerServiceDeployment{}) {
			continue
		}

		if int(aws.ToInt32(deployment.Version)) == version {
			result = deployment
			break
		}
	}

	if reflect.DeepEqual(result, types.ContainerServiceDeployment{}) {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return &result, nil
}
