// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecs_task_set", name="Task Set")
// @Tags(identifierAttribute="arn")
func resourceTaskSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTaskSetCreate,
		ReadWithoutTimeout:   resourceTaskSetRead,
		UpdateWithoutTimeout: resourceTaskSetUpdate,
		DeleteWithoutTimeout: resourceTaskSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCapacityProviderStrategy: {
				Type:          schema.TypeSet,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"launch_type"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
						},
						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrWeight: {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
						},
					},
				},
			},
			"cluster": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrExternalID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			names.AttrForceDelete: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"launch_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.LaunchType](),
				ConflictsWith:    []string{names.AttrCapacityProviderStrategy},
			},
			// If you are using the CodeDeploy or an external deployment controller,
			// multiple target groups are not supported.
			// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/register-multiple-targetgroups.html
			"load_balancer": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"container_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"load_balancer_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"target_group_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrNetworkConfiguration: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							MaxItems: 5,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							MaxItems: 16,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"scale": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrUnit: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.ScaleUnitPercent,
							ValidateDiagFunc: enum.Validate[awstypes.ScaleUnit](),
						},
						names.AttrValue: {
							Type:         schema.TypeFloat,
							Optional:     true,
							ValidateFunc: validation.FloatBetween(0.0, 100.0),
						},
					},
				},
			},
			"service": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_registries": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"container_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						names.AttrPort: {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"registry_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"stability_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"task_set_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wait_until_stable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"wait_until_stable_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "10m",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					duration, err := time.ParseDuration(value)
					if err != nil {
						errors = append(errors, fmt.Errorf(
							"%q cannot be parsed as a duration: %w", k, err))
					}
					if duration < 0 {
						errors = append(errors, fmt.Errorf(
							"%q must be greater than zero", k))
					}
					return
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTaskSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	cluster := d.Get("cluster").(string)
	service := d.Get("service").(string)
	input := &ecs.CreateTaskSetInput{
		ClientToken:    aws.String(id.UniqueId()),
		Cluster:        aws.String(cluster),
		Service:        aws.String(service),
		Tags:           getTagsIn(ctx),
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
	}

	if v, ok := d.GetOk(names.AttrCapacityProviderStrategy); ok && v.(*schema.Set).Len() > 0 {
		input.CapacityProviderStrategy = expandCapacityProviderStrategyItems(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrExternalID); ok {
		input.ExternalId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("launch_type"); ok {
		input.LaunchType = awstypes.LaunchType(v.(string))
	}

	if v, ok := d.GetOk("load_balancer"); ok && v.(*schema.Set).Len() > 0 {
		input.LoadBalancers = expandTaskSetLoadBalancers(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrNetworkConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.NetworkConfiguration = expandNetworkConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("platform_version"); ok {
		input.PlatformVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("scale"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Scale = expandScale(v.([]interface{}))
	}

	if v, ok := d.GetOk("service_registries"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ServiceRegistries = expandServiceRegistries(v.([]interface{}))
	}

	output, err := retryTaskSetCreate(ctx, conn, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = retryTaskSetCreate(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS TaskSet: %s", err)
	}

	taskSetID := aws.ToString(output.TaskSet.Id)
	d.SetId(taskSetCreateResourceID(taskSetID, service, cluster))

	if d.Get("wait_until_stable").(bool) {
		timeout, _ := time.ParseDuration(d.Get("wait_until_stable_timeout").(string))
		if _, err := waitTaskSetStable(ctx, conn, taskSetID, service, cluster, timeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ECS Task Set (%s) create: %s", d.Id(), err)
		}
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, aws.ToString(output.TaskSet.TaskSetArn), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceTaskSetRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECS Task Set (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTaskSetRead(ctx, d, meta)...)
}

func resourceTaskSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	taskSetID, service, cluster, err := taskSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	taskSet, err := findTaskSetByThreePartKey(ctx, conn, taskSetID, service, cluster)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECS Task Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, taskSet.TaskSetArn)
	if err := d.Set(names.AttrCapacityProviderStrategy, flattenCapacityProviderStrategyItems(taskSet.CapacityProviderStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting capacity_provider_strategy: %s", err)
	}
	d.Set("cluster", cluster)
	d.Set(names.AttrExternalID, taskSet.ExternalId)
	d.Set("launch_type", taskSet.LaunchType)
	if err := d.Set("load_balancer", flattenTaskSetLoadBalancers(taskSet.LoadBalancers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting load_balancer: %s", err)
	}
	if err := d.Set(names.AttrNetworkConfiguration, flattenNetworkConfiguration(taskSet.NetworkConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_configuration: %s", err)
	}
	d.Set("platform_version", taskSet.PlatformVersion)
	if err := d.Set("scale", flattenScale(taskSet.Scale)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting scale: %s", err)
	}
	d.Set("service", service)
	if err := d.Set("service_registries", flattenServiceRegistries(taskSet.ServiceRegistries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting service_registries: %s", err)
	}
	d.Set("stability_status", taskSet.StabilityStatus)
	d.Set(names.AttrStatus, taskSet.Status)
	d.Set("task_definition", taskSet.TaskDefinition)
	d.Set("task_set_id", taskSet.Id)

	setTagsOut(ctx, taskSet.Tags)

	return diags
}

func resourceTaskSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		taskSetID, service, cluster, err := taskSetParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &ecs.UpdateTaskSetInput{
			Cluster: aws.String(cluster),
			Scale:   expandScale(d.Get("scale").([]interface{})),
			Service: aws.String(service),
			TaskSet: aws.String(taskSetID),
		}

		_, err = conn.UpdateTaskSet(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECS Task Set (%s): %s", d.Id(), err)
		}

		if d.Get("wait_until_stable").(bool) {
			timeout, _ := time.ParseDuration(d.Get("wait_until_stable_timeout").(string))
			if _, err := waitTaskSetStable(ctx, conn, taskSetID, service, cluster, timeout); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for ECS Task Set (%s) update: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceTaskSetRead(ctx, d, meta)...)
}

func resourceTaskSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	taskSetID, service, cluster, err := taskSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting ECS Task Set: %s", d.Id())
	_, err = conn.DeleteTaskSet(ctx, &ecs.DeleteTaskSetInput{
		Cluster: aws.String(cluster),
		Force:   aws.Bool(d.Get(names.AttrForceDelete).(bool)),
		Service: aws.String(service),
		TaskSet: aws.String(taskSetID),
	})

	if errs.IsA[*awstypes.TaskSetNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Task Set (%s): %s", d.Id(), err)
	}

	if _, err := waitTaskSetDeleted(ctx, conn, taskSetID, service, cluster); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Task Set (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const taskSetResourceIDSeparator = ","

func taskSetCreateResourceID(taskSetID, service, cluster string) string {
	parts := []string{taskSetID, service, cluster}
	id := strings.Join(parts, taskSetResourceIDSeparator)

	return id
}

func taskSetParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, taskSetResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TASK_SET_ID%[2]sSERVICE%[2]sCLUSTER", id, taskSetResourceIDSeparator)
}

func retryTaskSetCreate(ctx context.Context, conn *ecs.Client, input *ecs.CreateTaskSetInput) (*ecs.CreateTaskSetOutput, error) {
	const (
		taskSetCreateTimeout = 10 * time.Minute
		timeout              = propagationTimeout + taskSetCreateTimeout
	)
	outputRaw, err := tfresource.RetryWhen(ctx, timeout,
		func() (interface{}, error) {
			return conn.CreateTaskSet(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*awstypes.ClusterNotFoundException](err) || errs.IsA[*awstypes.ServiceNotFoundException](err) || errs.IsA[*awstypes.TaskSetNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "does not have an associated load balancer") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return nil, err
	}

	return outputRaw.(*ecs.CreateTaskSetOutput), nil
}

func findTaskSet(ctx context.Context, conn *ecs.Client, input *ecs.DescribeTaskSetsInput) (*awstypes.TaskSet, error) {
	output, err := findTaskSets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTaskSets(ctx context.Context, conn *ecs.Client, input *ecs.DescribeTaskSetsInput) ([]awstypes.TaskSet, error) {
	output, err := conn.DescribeTaskSets(ctx, input)

	if errs.IsA[*awstypes.ClusterNotFoundException](err) || errs.IsA[*awstypes.ServiceNotFoundException](err) || errs.IsA[*awstypes.TaskSetNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TaskSets, nil
}

func findTaskSetByThreePartKey(ctx context.Context, conn *ecs.Client, taskSetID, service, cluster string) (*awstypes.TaskSet, error) {
	input := &ecs.DescribeTaskSetsInput{
		Cluster:  aws.String(cluster),
		Include:  []awstypes.TaskSetField{awstypes.TaskSetFieldTags},
		Service:  aws.String(service),
		TaskSets: []string{taskSetID},
	}

	output, err := findTaskSet(ctx, conn, input)

	// Some partitions (i.e., ISO) may not support tagging, giving error.
	if errs.IsUnsupportedOperationInPartitionError(partitionFromConn(conn), err) {
		input.Include = nil

		output, err = findTaskSet(ctx, conn, input)
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findTaskSetNoTagsByThreePartKey(ctx context.Context, conn *ecs.Client, taskSetID, service, cluster string) (*awstypes.TaskSet, error) {
	input := &ecs.DescribeTaskSetsInput{
		Cluster:  aws.String(cluster),
		Service:  aws.String(service),
		TaskSets: []string{taskSetID},
	}

	return findTaskSet(ctx, conn, input)
}

func statusTaskSetStability(ctx context.Context, conn *ecs.Client, taskSetID, service, cluster string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTaskSetNoTagsByThreePartKey(ctx, conn, taskSetID, service, cluster)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.StabilityStatus), err
	}
}

func statusTaskSet(ctx context.Context, conn *ecs.Client, taskSetID, service, cluster string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTaskSetNoTagsByThreePartKey(ctx, conn, taskSetID, service, cluster)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), err
	}
}

const (
	taskSetStatusActive   = "ACTIVE"
	taskSetStatusDraining = "DRAINING"
	taskSetStatusPrimary  = "PRIMARY"
)

// Does not return tags.
func waitTaskSetStable(ctx context.Context, conn *ecs.Client, taskSetID, service, cluster string, timeout time.Duration) (*awstypes.TaskSet, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StabilityStatusStabilizing),
		Target:  enum.Slice(awstypes.StabilityStatusSteadyState),
		Refresh: statusTaskSetStability(ctx, conn, taskSetID, service, cluster),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TaskSet); ok {
		return output, err
	}

	return nil, err
}

// Does not return tags.
func waitTaskSetDeleted(ctx context.Context, conn *ecs.Client, taskSetID, service, cluster string) (*awstypes.TaskSet, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{taskSetStatusActive, taskSetStatusPrimary, taskSetStatusDraining},
		Target:  []string{},
		Refresh: statusTaskSet(ctx, conn, taskSetID, service, cluster),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TaskSet); ok {
		return output, err
	}

	return nil, err
}
