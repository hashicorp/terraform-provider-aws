// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrserverless

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emrserverless"
	"github.com/aws/aws-sdk-go-v2/service/emrserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_emrserverless_application", name="Application")
// @Tags(identifierAttribute="arn")
func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationCreate,
		ReadWithoutTimeout:   resourceApplicationRead,
		UpdateWithoutTimeout: resourceApplicationUpdate,
		DeleteWithoutTimeout: resourceApplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"architecture": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.ArchitectureX8664,
				ValidateDiagFunc: enum.Validate[types.Architecture](),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_start_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"auto_stop_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"idle_timeout_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      15,
							ValidateFunc: validation.IntBetween(1, 10080),
						},
					},
				},
			},
			"image_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_uri": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"initial_capacity": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"initial_capacity_config": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"worker_configuration": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu": {
													Type:     schema.TypeString,
													Required: true,
												},
												"disk": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"memory": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"worker_count": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 1000000),
									},
								},
							},
						},
						"initial_capacity_type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"interactive_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"livy_endpoint_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"studio_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"maximum_capacity": {
				Type:             schema.TypeList,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:     schema.TypeString,
							Required: true,
						},
						"disk": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"memory": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			names.AttrNetworkConfiguration: {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"release_label": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRServerlessClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &emrserverless.CreateApplicationInput{
		ClientToken:  aws.String(id.UniqueId()),
		ReleaseLabel: aws.String(d.Get("release_label").(string)),
		Name:         aws.String(name),
		Tags:         getTagsIn(ctx),
		Type:         aws.String(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk("architecture"); ok {
		input.Architecture = types.Architecture(v.(string))
	}

	if v, ok := d.GetOk("auto_start_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AutoStartConfiguration = expandAutoStartConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("auto_stop_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AutoStopConfiguration = expandAutoStopConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("image_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ImageConfiguration = expandImageConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("initial_capacity"); ok && v.(*schema.Set).Len() > 0 {
		input.InitialCapacity = expandInitialCapacity(v.(*schema.Set))
	}

	if v, ok := d.GetOk("interactive_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.InteractiveConfiguration = expandInteractiveConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("maximum_capacity"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.MaximumCapacity = expandMaximumCapacity(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrNetworkConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.NetworkConfiguration = expandNetworkConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateApplication(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EMR Serveless Application (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ApplicationId))

	if _, err := waitApplicationCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EMR Serveless Application (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRServerlessClient(ctx)

	application, err := findApplicationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Serverless Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Serverless Application (%s): %s", d.Id(), err)
	}

	d.Set("architecture", application.Architecture)
	d.Set(names.AttrARN, application.Arn)
	d.Set(names.AttrName, application.Name)
	d.Set("release_label", application.ReleaseLabel)
	d.Set(names.AttrType, strings.ToLower(aws.ToString(application.Type)))

	if err := d.Set("auto_start_configuration", []interface{}{flattenAutoStartConfig(application.AutoStartConfiguration)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting auto_start_configuration: %s", err)
	}

	if err := d.Set("auto_stop_configuration", []interface{}{flattenAutoStopConfig(application.AutoStopConfiguration)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting auto_stop_configuration: %s", err)
	}

	if err := d.Set("image_configuration", flattenImageConfiguration(application.ImageConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting image_configuration: %s", err)
	}

	if err := d.Set("initial_capacity", flattenInitialCapacity(application.InitialCapacity)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting initial_capacity: %s", err)
	}

	if err := d.Set("interactive_configuration", []interface{}{flattenInteractiveConfiguration(application.InteractiveConfiguration)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting interactive_configuration: %s", err)
	}

	if err := d.Set("maximum_capacity", []interface{}{flattenMaximumCapacity(application.MaximumCapacity)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting maximum_capacity: %s", err)
	}

	if err := d.Set(names.AttrNetworkConfiguration, []interface{}{flattenNetworkConfiguration(application.NetworkConfiguration)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_configuration: %s", err)
	}

	setTagsOut(ctx, application.Tags)

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRServerlessClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &emrserverless.UpdateApplicationInput{
			ApplicationId: aws.String(d.Id()),
			ClientToken:   aws.String(id.UniqueId()),
		}

		if v, ok := d.GetOk("architecture"); ok {
			input.Architecture = types.Architecture(v.(string))
		}

		if v, ok := d.GetOk("auto_start_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.AutoStartConfiguration = expandAutoStartConfig(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("auto_stop_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.AutoStopConfiguration = expandAutoStopConfig(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("image_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ImageConfiguration = expandImageConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("initial_capacity"); ok && v.(*schema.Set).Len() > 0 {
			input.InitialCapacity = expandInitialCapacity(v.(*schema.Set))
		}

		if v, ok := d.GetOk("interactive_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.InteractiveConfiguration = expandInteractiveConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("maximum_capacity"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.MaximumCapacity = expandMaximumCapacity(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk(names.AttrNetworkConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.NetworkConfiguration = expandNetworkConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("release_label"); ok {
			input.ReleaseLabel = aws.String(v.(string))
		}

		_, err := conn.UpdateApplication(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Serveless Application (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRServerlessClient(ctx)

	log.Printf("[INFO] Deleting EMR Serverless Application: %s", d.Id())
	_, err := conn.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
		ApplicationId: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EMR Serverless Application (%s): %s", d.Id(), err)
	}

	if _, err := waitApplicationTerminated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EMR Serveless Application (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findApplicationByID(ctx context.Context, conn *emrserverless.Client, id string) (*types.Application, error) {
	input := &emrserverless.GetApplicationInput{
		ApplicationId: aws.String(id),
	}

	output, err := conn.GetApplication(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Application == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if output.Application.State == types.ApplicationStateTerminated {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Application, nil
}

func statusApplication(ctx context.Context, conn *emrserverless.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findApplicationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitApplicationCreated(ctx context.Context, conn *emrserverless.Client, id string) (*types.Application, error) {
	const (
		timeout    = 75 * time.Minute
		minTimeout = 10 * time.Second
		delay      = 30 * time.Second
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ApplicationStateCreating),
		Target:     enum.Slice(types.ApplicationStateCreated),
		Refresh:    statusApplication(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: minTimeout,
		Delay:      delay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Application); ok {
		if stateChangeReason := output.StateDetails; stateChangeReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateChangeReason)))
		}

		return output, err
	}

	return nil, err
}

func waitApplicationTerminated(ctx context.Context, conn *emrserverless.Client, id string) (*types.Application, error) {
	const (
		timeout    = 20 * time.Minute
		minTimeout = 10 * time.Second
		delay      = 30 * time.Second
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Values[types.ApplicationState](),
		Target:     []string{},
		Refresh:    statusApplication(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: minTimeout,
		Delay:      delay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Application); ok {
		if stateChangeReason := output.StateDetails; stateChangeReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateChangeReason)))
		}

		return output, err
	}

	return nil, err
}

func expandAutoStartConfig(tfMap map[string]interface{}) *types.AutoStartConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AutoStartConfig{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func flattenAutoStartConfig(apiObject *types.AutoStartConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	return tfMap
}

func expandAutoStopConfig(tfMap map[string]interface{}) *types.AutoStopConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AutoStopConfig{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["idle_timeout_minutes"].(int); ok {
		apiObject.IdleTimeoutMinutes = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenAutoStopConfig(apiObject *types.AutoStopConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.IdleTimeoutMinutes; v != nil {
		tfMap["idle_timeout_minutes"] = aws.ToInt32(v)
	}

	return tfMap
}

func expandInteractiveConfiguration(tfMap map[string]interface{}) *types.InteractiveConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.InteractiveConfiguration{}

	if v, ok := tfMap["livy_endpoint_enabled"].(bool); ok {
		apiObject.LivyEndpointEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["studio_enabled"].(bool); ok {
		apiObject.StudioEnabled = aws.Bool(v)
	}

	return apiObject
}

func flattenInteractiveConfiguration(apiObject *types.InteractiveConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LivyEndpointEnabled; v != nil {
		tfMap["livy_endpoint_enabled"] = aws.ToBool(v)
	}

	if v := apiObject.StudioEnabled; v != nil {
		tfMap["studio_enabled"] = aws.ToBool(v)
	}

	return tfMap
}

func expandMaximumCapacity(tfMap map[string]interface{}) *types.MaximumAllowedResources {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.MaximumAllowedResources{}

	if v, ok := tfMap["cpu"].(string); ok && v != "" {
		apiObject.Cpu = aws.String(v)
	}

	if v, ok := tfMap["disk"].(string); ok && v != "" {
		apiObject.Disk = aws.String(v)
	}

	if v, ok := tfMap["memory"].(string); ok && v != "" {
		apiObject.Memory = aws.String(v)
	}

	return apiObject
}

func flattenMaximumCapacity(apiObject *types.MaximumAllowedResources) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Cpu; v != nil {
		tfMap["cpu"] = aws.ToString(v)
	}

	if v := apiObject.Disk; v != nil {
		tfMap["disk"] = aws.ToString(v)
	}

	if v := apiObject.Memory; v != nil {
		tfMap["memory"] = aws.ToString(v)
	}

	return tfMap
}

func expandNetworkConfiguration(tfMap map[string]interface{}) *types.NetworkConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.NetworkConfiguration{}

	if v, ok := tfMap[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func flattenNetworkConfiguration(apiObject *types.NetworkConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap[names.AttrSecurityGroupIDs] = flex.FlattenStringValueSet(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = flex.FlattenStringValueSet(v)
	}

	return tfMap
}

func expandImageConfiguration(tfMap map[string]interface{}) *types.ImageConfigurationInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ImageConfigurationInput{}

	if v, ok := tfMap["image_uri"].(string); ok && v != "" {
		apiObject.ImageUri = aws.String(v)
	}

	return apiObject
}

func flattenImageConfiguration(apiObject *types.ImageConfiguration) []interface{} {
	if apiObject == nil || apiObject.ImageUri == nil {
		return nil
	}

	var tfList []interface{}

	if v := apiObject.ImageUri; v != nil {
		tfList = append(tfList, map[string]interface{}{
			"image_uri": aws.ToString(v),
		})
	}

	return tfList
}

func expandInitialCapacity(tfMap *schema.Set) map[string]types.InitialCapacityConfig {
	if tfMap == nil {
		return nil
	}

	configs := make(map[string]types.InitialCapacityConfig)

	for _, tfMapRaw := range tfMap.List() {
		config, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		if v, ok := config["initial_capacity_type"].(string); ok && v != "" {
			if conf, ok := config["initial_capacity_config"].([]interface{}); ok && len(conf) > 0 {
				configs[v] = expandInitialCapacityConfig(conf[0].(map[string]interface{}))
			}
		}
	}

	return configs
}

func flattenInitialCapacity(apiObject map[string]types.InitialCapacityConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}

	for capacityType, config := range apiObject {
		tfList = append(tfList, map[string]interface{}{
			"initial_capacity_type":   capacityType,
			"initial_capacity_config": []interface{}{flattenInitialCapacityConfig(&config)},
		})
	}

	return tfList
}

func expandInitialCapacityConfig(tfMap map[string]interface{}) types.InitialCapacityConfig {
	apiObject := types.InitialCapacityConfig{}

	if v, ok := tfMap["worker_count"].(int); ok {
		apiObject.WorkerCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["worker_configuration"].([]interface{}); ok && v[0] != nil {
		apiObject.WorkerConfiguration = expandWorkerResourceConfig(v[0].(map[string]interface{}))
	}

	return apiObject
}

func flattenInitialCapacityConfig(apiObject *types.InitialCapacityConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"worker_count": apiObject.WorkerCount,
	}

	if v := apiObject.WorkerConfiguration; v != nil {
		tfMap["worker_configuration"] = []interface{}{flattenWorkerResourceConfig(v)}
	}

	return tfMap
}

func expandWorkerResourceConfig(tfMap map[string]interface{}) *types.WorkerResourceConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.WorkerResourceConfig{}

	if v, ok := tfMap["cpu"].(string); ok && v != "" {
		apiObject.Cpu = aws.String(v)
	}

	if v, ok := tfMap["disk"].(string); ok && v != "" {
		apiObject.Disk = aws.String(v)
	}

	if v, ok := tfMap["memory"].(string); ok && v != "" {
		apiObject.Memory = aws.String(v)
	}

	return apiObject
}

func flattenWorkerResourceConfig(apiObject *types.WorkerResourceConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Cpu; v != nil {
		tfMap["cpu"] = aws.ToString(v)
	}

	if v := apiObject.Disk; v != nil {
		tfMap["disk"] = aws.ToString(v)
	}

	if v := apiObject.Memory; v != nil {
		tfMap["memory"] = aws.ToString(v)
	}

	return tfMap
}
