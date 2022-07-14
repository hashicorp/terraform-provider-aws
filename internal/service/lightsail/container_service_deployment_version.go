package lightsail

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceContainerServiceDeploymentVersion() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceContainerServiceDeploymentVersionCreate,
		ReadContext:   resourceContainerServiceDeploymentVersionRead,
		DeleteContext: resourceContainerServiceDeploymentVersionDelete,
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
						"environment": {
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
								ValidateFunc: validation.StringInSlice(lightsail.ContainerServiceProtocol_Values(), false),
							}},
					},
				},
			},
			"created_at": {
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
						"health_check": {
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
									"path": {
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
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceContainerServiceDeploymentVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn
	serviceName := d.Get("service_name").(string)

	input := &lightsail.CreateContainerServiceDeploymentInput{
		ServiceName: aws.String(serviceName),
	}

	if v, ok := d.GetOk("container"); ok && v.(*schema.Set).Len() > 0 {
		input.Containers = expandContainerServiceDeploymentContainers(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("public_endpoint"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.PublicEndpoint = expandContainerServiceDeploymentPublicEndpoint(v.([]interface{}))
	}

	output, err := conn.CreateContainerServiceDeploymentWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("error creating Lightsail Container Service (%s) Deployment Version: %s", serviceName, err)
	}

	if output == nil || output.ContainerService == nil || output.ContainerService.NextDeployment == nil {
		return diag.Errorf("error creating Lightsail Container Service (%s) Deployment Version: empty output", serviceName)
	}

	version := int(aws.Int64Value(output.ContainerService.NextDeployment.Version))

	d.SetId(fmt.Sprintf("%s/%d", serviceName, version))

	if err := waitContainerServiceDeploymentVersionActive(ctx, conn, serviceName, version, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for Lightsail Container Service (%s) Deployment Version (%d): %s", serviceName, version, err)
	}

	return resourceContainerServiceDeploymentVersionRead(ctx, d, meta)
}

func resourceContainerServiceDeploymentVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn

	serviceName, version, err := ContainerServiceDeploymentVersionParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err := FindContainerServiceDeploymentByVersion(ctx, conn, serviceName, version)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lightsail Container Service (%s) Deployment Version (%d) not found, removing from state", serviceName, version)
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Lightsail Container Service (%s) Deployment Version (%d): %s", serviceName, version, err)
	}

	d.Set("created_at", aws.TimeValue(deployment.CreatedAt).Format(time.RFC3339))
	d.Set("service_name", serviceName)
	d.Set("state", deployment.State)
	d.Set("version", deployment.Version)

	if err := d.Set("container", flattenContainerServiceDeploymentContainers(deployment.Containers)); err != nil {
		return diag.Errorf("error setting container for Lightsail Container Service (%s) Deployment Version (%d): %s", serviceName, version, err)
	}

	if err := d.Set("public_endpoint", flattenContainerServiceDeploymentPublicEndpoint(deployment.PublicEndpoint)); err != nil {
		return diag.Errorf("error setting public_endpoint for Lightsail Container Service (%s) Deployment Version (%d): %s", serviceName, version, err)
	}

	return nil
}

func resourceContainerServiceDeploymentVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[WARN] Cannot destroy Lightsail Container Service Deployment Version. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
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

func expandContainerServiceDeploymentContainers(tfList []interface{}) map[string]*lightsail.Container {
	if len(tfList) == 0 {
		return map[string]*lightsail.Container{}
	}

	result := make(map[string]*lightsail.Container)

	for _, tfListRaw := range tfList {
		tfMap, ok := tfListRaw.(map[string]interface{})
		if !ok {
			continue
		}

		containerName := tfMap["container_name"].(string)

		container := &lightsail.Container{
			Image: aws.String(tfMap["image"].(string)),
		}

		if v, ok := tfMap["command"].([]interface{}); ok && len(v) > 0 {
			container.Command = flex.ExpandStringList(v)
		}

		if v, ok := tfMap["environment"].(map[string]interface{}); ok && len(v) > 0 {
			container.Environment = flex.ExpandStringMap(v)
		}

		if v, ok := tfMap["ports"].(map[string]interface{}); ok && len(v) > 0 {
			container.Ports = flex.ExpandStringMap(v)
		}

		result[containerName] = container
	}

	return result
}

func expandContainerServiceDeploymentPublicEndpoint(tfList []interface{}) *lightsail.EndpointRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	endpoint := &lightsail.EndpointRequest{
		ContainerName: aws.String(tfMap["container_name"].(string)),
		ContainerPort: aws.Int64(int64(tfMap["container_port"].(int))),
	}

	if v, ok := tfMap["health_check"].([]interface{}); ok && len(v) > 0 {
		endpoint.HealthCheck = expandContainerServiceDeploymentPublicEndpointHealthCheck(v)
	}

	return endpoint
}

func expandContainerServiceDeploymentPublicEndpointHealthCheck(tfList []interface{}) *lightsail.ContainerServiceHealthCheckConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	healthCheck := &lightsail.ContainerServiceHealthCheckConfig{
		HealthyThreshold:   aws.Int64(int64(tfMap["healthy_threshold"].(int))),
		IntervalSeconds:    aws.Int64(int64(tfMap["interval_seconds"].(int))),
		Path:               aws.String(tfMap["path"].(string)),
		SuccessCodes:       aws.String(tfMap["success_codes"].(string)),
		TimeoutSeconds:     aws.Int64(int64(tfMap["timeout_seconds"].(int))),
		UnhealthyThreshold: aws.Int64(int64(tfMap["unhealthy_threshold"].(int))),
	}

	return healthCheck
}

func flattenContainerServiceDeploymentContainers(containers map[string]*lightsail.Container) []interface{} {
	if len(containers) == 0 {
		return nil
	}

	var rawContainers []interface{}
	for containerName, container := range containers {
		rawContainer := map[string]interface{}{
			"container_name": containerName,
			"image":          aws.StringValue(container.Image),
			"command":        aws.StringValueSlice(container.Command),
			"environment":    aws.StringValueMap(container.Environment),
			"ports":          aws.StringValueMap(container.Ports),
		}

		rawContainers = append(rawContainers, rawContainer)
	}

	return rawContainers
}

func flattenContainerServiceDeploymentPublicEndpoint(endpoint *lightsail.ContainerServiceEndpoint) []interface{} {
	if endpoint == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"container_name": aws.StringValue(endpoint.ContainerName),
			"container_port": int(aws.Int64Value(endpoint.ContainerPort)),
			"health_check":   flattenContainerServiceDeploymentPublicEndpointHealthCheck(endpoint.HealthCheck),
		},
	}
}

func flattenContainerServiceDeploymentPublicEndpointHealthCheck(healthCheck *lightsail.ContainerServiceHealthCheckConfig) []interface{} {
	if healthCheck == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"healthy_threshold":   int(aws.Int64Value(healthCheck.HealthyThreshold)),
			"interval_seconds":    int(aws.Int64Value(healthCheck.IntervalSeconds)),
			"path":                aws.StringValue(healthCheck.Path),
			"success_codes":       aws.StringValue(healthCheck.SuccessCodes),
			"timeout_seconds":     int(aws.Int64Value(healthCheck.TimeoutSeconds)),
			"unhealthy_threshold": int(aws.Int64Value(healthCheck.UnhealthyThreshold)),
		},
	}
}
