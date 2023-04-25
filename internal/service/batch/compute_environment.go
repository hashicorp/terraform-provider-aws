package batch

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_batch_compute_environment", name="Compute Environment")
// @Tags(identifierAttribute="arn")
func ResourceComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceComputeEnvironmentCreate,
		ReadWithoutTimeout:   resourceComputeEnvironmentRead,
		UpdateWithoutTimeout: resourceComputeEnvironmentUpdate,
		DeleteWithoutTimeout: resourceComputeEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			resourceComputeEnvironmentCustomizeDiff,
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_environment_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name_prefix"},
				ValidateFunc:  validName,
			},
			"compute_environment_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name"},
				ValidateFunc:  validPrefix,
			},
			"compute_resources": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							StateFunc: func(val interface{}) string {
								return strings.ToUpper(val.(string))
							},
							ValidateFunc: validation.StringInSlice(batch.CRAllocationStrategy_Values(), true),
						},
						"bid_percentage": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"desired_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"ec2_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_id_override": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"image_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
						"ec2_key_pair": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"image_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"instance_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"instance_type": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"launch_template": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"launch_template_id": {
										Type:          schema.TypeString,
										Optional:      true,
										ForceNew:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_name"},
									},
									"launch_template_name": {
										Type:          schema.TypeString,
										Optional:      true,
										ForceNew:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_id"},
									},
									"version": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"max_vcpus": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"min_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"spot_iam_fleet_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"tags": tftags.TagsSchemaForceNew(),
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							StateFunc: func(val interface{}) string {
								return strings.ToUpper(val.(string))
							},
							ValidateFunc: validation.StringInSlice(batch.CRType_Values(), true),
						},
					},
				},
			},
			"ecs_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"eks_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"eks_cluster_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"kubernetes_namespace": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"service_role": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateFunc: validation.StringInSlice(batch.CEState_Values(), true),
				Default:      batch.CEStateEnabled,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateFunc: validation.StringInSlice(batch.CEType_Values(), true),
			},
		},
	}
}

func resourceComputeEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn()

	computeEnvironmentName := create.Name(d.Get("compute_environment_name").(string), d.Get("compute_environment_name_prefix").(string))
	computeEnvironmentType := d.Get("type").(string)
	input := &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(computeEnvironmentName),
		ServiceRole:            aws.String(d.Get("service_role").(string)),
		Tags:                   GetTagsIn(ctx),
		Type:                   aws.String(computeEnvironmentType),
	}

	if v, ok := d.GetOk("compute_resources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ComputeResources = expandComputeResource(ctx, v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("eks_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EksConfiguration = expandEKSConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("state"); ok {
		input.State = aws.String(v.(string))
	}

	output, err := conn.CreateComputeEnvironmentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Batch Compute Environment (%s): %s", computeEnvironmentName, err)
	}

	d.SetId(aws.StringValue(output.ComputeEnvironmentName))

	if _, err := waitComputeEnvironmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Batch Compute Environment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceComputeEnvironmentRead(ctx, d, meta)...)
}

func resourceComputeEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn()

	computeEnvironment, err := FindComputeEnvironmentDetailByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Batch Compute Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Batch Compute Environment (%s): %s", d.Id(), err)
	}

	computeEnvironmentType := aws.StringValue(computeEnvironment.Type)

	d.Set("arn", computeEnvironment.ComputeEnvironmentArn)
	d.Set("compute_environment_name", computeEnvironment.ComputeEnvironmentName)
	d.Set("compute_environment_name_prefix", create.NamePrefixFromName(aws.StringValue(computeEnvironment.ComputeEnvironmentName)))
	d.Set("ecs_cluster_arn", computeEnvironment.EcsClusterArn)
	d.Set("service_role", computeEnvironment.ServiceRole)
	d.Set("state", computeEnvironment.State)
	d.Set("status", computeEnvironment.Status)
	d.Set("status_reason", computeEnvironment.StatusReason)
	d.Set("type", computeEnvironmentType)

	if computeEnvironment.ComputeResources != nil {
		if err := d.Set("compute_resources", []interface{}{flattenComputeResource(ctx, computeEnvironment.ComputeResources)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting compute_resources: %s", err)
		}
	} else {
		d.Set("compute_resources", nil)
	}

	if computeEnvironment.EksConfiguration != nil {
		if err := d.Set("eks_configuration", []interface{}{flattenEKSConfiguration(computeEnvironment.EksConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting eks_configuration: %s", err)
		}
	} else {
		d.Set("eks_configuration", nil)
	}

	SetTagsOut(ctx, computeEnvironment.Tags)

	return diags
}

func resourceComputeEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
		}

		if d.HasChange("service_role") {
			input.ServiceRole = aws.String(d.Get("service_role").(string))
		}

		if d.HasChange("state") {
			input.State = aws.String(d.Get("state").(string))
		}

		if computeEnvironmentType := strings.ToUpper(d.Get("type").(string)); computeEnvironmentType == batch.CETypeManaged {
			// "At least one compute-resources attribute must be specified"
			computeResourceUpdate := &batch.ComputeResourceUpdate{
				MaxvCpus: aws.Int64(int64(d.Get("compute_resources.0.max_vcpus").(int))),
			}

			if d.HasChange("compute_resources.0.desired_vcpus") {
				computeResourceUpdate.DesiredvCpus = aws.Int64(int64(d.Get("compute_resources.0.desired_vcpus").(int)))
			}

			if d.HasChange("compute_resources.0.min_vcpus") {
				computeResourceUpdate.MinvCpus = aws.Int64(int64(d.Get("compute_resources.0.min_vcpus").(int)))
			}

			if d.HasChange("compute_resources.0.security_group_ids") {
				computeResourceUpdate.SecurityGroupIds = flex.ExpandStringSet(d.Get("compute_resources.0.security_group_ids").(*schema.Set))
			}

			if d.HasChange("compute_resources.0.subnets") {
				computeResourceUpdate.Subnets = flex.ExpandStringSet(d.Get("compute_resources.0.subnets").(*schema.Set))
			}

			input.ComputeResources = computeResourceUpdate
		}

		log.Printf("[DEBUG] Updating Batch Compute Environment: %s", input)
		if _, err := conn.UpdateComputeEnvironmentWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Batch Compute Environment (%s): %s", d.Id(), err)
		}

		if _, err := waitComputeEnvironmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Batch Compute Environment (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceComputeEnvironmentRead(ctx, d, meta)...)
}

func resourceComputeEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn()

	log.Printf("[DEBUG] Disabling Batch Compute Environment: %s", d.Id())
	{
		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
			State:              aws.String(batch.CEStateDisabled),
		}

		if _, err := conn.UpdateComputeEnvironmentWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling Batch Compute Environment (%s): %s", d.Id(), err)
		}

		if _, err := waitComputeEnvironmentDisabled(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			log.Printf("[WARN] error waiting for Batch Compute Environment (%s) disable: %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting Batch Compute Environment: %s", d.Id())
	{
		input := &batch.DeleteComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
		}

		if _, err := conn.DeleteComputeEnvironmentWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Batch Compute Environment (%s): %s", d.Id(), err)
		}

		if _, err := waitComputeEnvironmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Batch Compute Environment (%s) delete: %s", d.Id(), err)
		}
	}

	return diags
}

func resourceComputeEnvironmentCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if computeEnvironmentType := strings.ToUpper(diff.Get("type").(string)); computeEnvironmentType == batch.CETypeUnmanaged {
		// UNMANAGED compute environments can have no compute_resources configured.
		if v, ok := diff.GetOk("compute_resources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			return fmt.Errorf("no `compute_resources` can be specified when `type` is %q", computeEnvironmentType)
		}
	}

	if diff.Id() != "" {
		// Update.

		computeResourceType := strings.ToUpper(diff.Get("compute_resources.0.type").(string))
		fargateComputeResources := false
		if computeResourceType == batch.CRTypeFargate || computeResourceType == batch.CRTypeFargateSpot {
			fargateComputeResources = true
		}

		if diff.HasChange("compute_resources.0.security_group_ids") && !fargateComputeResources {
			if err := diff.ForceNew("compute_resources.0.security_group_ids"); err != nil {
				return err
			}
		}

		if diff.HasChange("compute_resources.0.subnets") && !fargateComputeResources {
			if err := diff.ForceNew("compute_resources.0.subnets"); err != nil {
				return err
			}
		}
	}

	return nil
}

func FindComputeEnvironmentDetailByName(ctx context.Context, conn *batch.Batch, name string) (*batch.ComputeEnvironmentDetail, error) {
	input := &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: aws.StringSlice([]string{name}),
	}

	output, err := findComputeEnvironmentDetail(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status); status == batch.CEStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findComputeEnvironmentDetail(ctx context.Context, conn *batch.Batch, input *batch.DescribeComputeEnvironmentsInput) (*batch.ComputeEnvironmentDetail, error) {
	output, err := conn.DescribeComputeEnvironmentsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ComputeEnvironments) == 0 || output.ComputeEnvironments[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ComputeEnvironments); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.ComputeEnvironments[0], nil
}

func statusComputeEnvironment(ctx context.Context, conn *batch.Batch, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		computeEnvironmentDetail, err := FindComputeEnvironmentDetailByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return computeEnvironmentDetail, aws.StringValue(computeEnvironmentDetail.Status), nil
	}
}

func waitComputeEnvironmentCreated(ctx context.Context, conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{batch.CEStatusCreating},
		Target:  []string{batch.CEStatusValid},
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		if status := aws.StringValue(output.Status); status == batch.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitComputeEnvironmentDeleted(ctx context.Context, conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{batch.CEStatusDeleting},
		Target:  []string{},
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		if status := aws.StringValue(output.Status); status == batch.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitComputeEnvironmentDisabled(ctx context.Context, conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{batch.CEStatusUpdating},
		Target:  []string{batch.CEStatusValid},
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		if status := aws.StringValue(output.Status); status == batch.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitComputeEnvironmentUpdated(ctx context.Context, conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{batch.CEStatusUpdating},
		Target:  []string{batch.CEStatusValid},
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		return v, err
	}

	return nil, err
}

func expandComputeResource(ctx context.Context, tfMap map[string]interface{}) *batch.ComputeResource {
	if tfMap == nil {
		return nil
	}

	var computeResourceType string

	if v, ok := tfMap["type"].(string); ok && v != "" {
		computeResourceType = v
	}

	apiObject := &batch.ComputeResource{}

	if v, ok := tfMap["allocation_strategy"].(string); ok && v != "" {
		apiObject.AllocationStrategy = aws.String(v)
	}

	if v, ok := tfMap["bid_percentage"].(int); ok && v != 0 {
		apiObject.BidPercentage = aws.Int64(int64(v))
	}

	if v, ok := tfMap["desired_vcpus"].(int); ok && v != 0 {
		apiObject.DesiredvCpus = aws.Int64(int64(v))
	}

	if v, ok := tfMap["ec2_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.Ec2Configuration = expandEC2Configurations(v)
	}

	if v, ok := tfMap["ec2_key_pair"].(string); ok && v != "" {
		apiObject.Ec2KeyPair = aws.String(v)
	}

	if v, ok := tfMap["image_id"].(string); ok && v != "" {
		apiObject.ImageId = aws.String(v)
	}

	if v, ok := tfMap["instance_role"].(string); ok && v != "" {
		apiObject.InstanceRole = aws.String(v)
	}

	if v, ok := tfMap["instance_type"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.InstanceTypes = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["launch_template"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.LaunchTemplate = expandLaunchTemplateSpecification(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["max_vcpus"].(int); ok && v != 0 {
		apiObject.MaxvCpus = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min_vcpus"].(int); ok && v != 0 {
		apiObject.MinvCpus = aws.Int64(int64(v))
	} else if computeResourceType := strings.ToUpper(computeResourceType); computeResourceType == batch.CRTypeEc2 || computeResourceType == batch.CRTypeSpot {
		apiObject.MinvCpus = aws.Int64(0)
	}

	if v, ok := tfMap["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["spot_iam_fleet_role"].(string); ok && v != "" {
		apiObject.SpotIamFleetRole = aws.String(v)
	}

	if v, ok := tfMap["subnets"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["tags"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Tags = Tags(tftags.New(ctx, v).IgnoreAWS())
	}

	if computeResourceType != "" {
		apiObject.Type = aws.String(computeResourceType)
	}

	return apiObject
}

func expandEKSConfiguration(tfMap map[string]interface{}) *batch.EksConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.EksConfiguration{}

	if v, ok := tfMap["eks_cluster_arn"].(string); ok && v != "" {
		apiObject.EksClusterArn = aws.String(v)
	}

	if v, ok := tfMap["kubernetes_namespace"].(string); ok && v != "" {
		apiObject.KubernetesNamespace = aws.String(v)
	}

	return apiObject
}

func expandEC2Configuration(tfMap map[string]interface{}) *batch.Ec2Configuration {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.Ec2Configuration{}

	if v, ok := tfMap["image_id_override"].(string); ok && v != "" {
		apiObject.ImageIdOverride = aws.String(v)
	}

	if v, ok := tfMap["image_type"].(string); ok && v != "" {
		apiObject.ImageType = aws.String(v)
	}

	return apiObject
}

func expandEC2Configurations(tfList []interface{}) []*batch.Ec2Configuration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*batch.Ec2Configuration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEC2Configuration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateSpecification(tfMap map[string]interface{}) *batch.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.LaunchTemplateSpecification{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["launch_template_name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func flattenComputeResource(ctx context.Context, apiObject *batch.ComputeResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllocationStrategy; v != nil {
		tfMap["allocation_strategy"] = aws.StringValue(v)
	}

	if v := apiObject.BidPercentage; v != nil {
		tfMap["bid_percentage"] = aws.Int64Value(v)
	}

	if v := apiObject.DesiredvCpus; v != nil {
		tfMap["desired_vcpus"] = aws.Int64Value(v)
	}

	if v := apiObject.Ec2Configuration; v != nil {
		tfMap["ec2_configuration"] = flattenEC2Configurations(v)
	}

	if v := apiObject.Ec2KeyPair; v != nil {
		tfMap["ec2_key_pair"] = aws.StringValue(v)
	}

	if v := apiObject.ImageId; v != nil {
		tfMap["image_id"] = aws.StringValue(v)
	}

	if v := apiObject.InstanceRole; v != nil {
		tfMap["instance_role"] = aws.StringValue(v)
	}

	if v := apiObject.InstanceTypes; v != nil {
		tfMap["instance_type"] = aws.StringValueSlice(v)
	}

	if v := apiObject.LaunchTemplate; v != nil {
		tfMap["launch_template"] = []interface{}{flattenLaunchTemplateSpecification(v)}
	}

	if v := apiObject.MaxvCpus; v != nil {
		tfMap["max_vcpus"] = aws.Int64Value(v)
	}

	if v := apiObject.MinvCpus; v != nil {
		tfMap["min_vcpus"] = aws.Int64Value(v)
	}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap["security_group_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.SpotIamFleetRole; v != nil {
		tfMap["spot_iam_fleet_role"] = aws.StringValue(v)
	}

	if v := apiObject.Subnets; v != nil {
		tfMap["subnets"] = aws.StringValueSlice(v)
	}

	if v := apiObject.Tags; v != nil {
		tfMap["tags"] = KeyValueTags(ctx, v).IgnoreAWS().Map()
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEKSConfiguration(apiObject *batch.EksConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EksClusterArn; v != nil {
		tfMap["eks_cluster_arn"] = aws.StringValue(v)
	}

	if v := apiObject.KubernetesNamespace; v != nil {
		tfMap["kubernetes_namespace"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEC2Configuration(apiObject *batch.Ec2Configuration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ImageIdOverride; v != nil {
		tfMap["image_id_override"] = aws.StringValue(v)
	}

	if v := apiObject.ImageType; v != nil {
		tfMap["image_type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEC2Configurations(apiObjects []*batch.Ec2Configuration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenEC2Configuration(apiObject))
	}

	return tfList
}

func flattenLaunchTemplateSpecification(apiObject *batch.LaunchTemplateSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.StringValue(v)
	}

	if v := apiObject.LaunchTemplateName; v != nil {
		tfMap["launch_template_name"] = aws.StringValue(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap["version"] = aws.StringValue(v)
	}

	return tfMap
}
