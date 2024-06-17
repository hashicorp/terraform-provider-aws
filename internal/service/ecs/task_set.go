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
func ResourceTaskSet() *schema.Resource {
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
		input.CapacityProviderStrategy = expandCapacityProviderStrategy(v.(*schema.Set))
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

	taskSetId := aws.ToString(output.TaskSet.Id)

	d.SetId(fmt.Sprintf("%s,%s,%s", taskSetId, service, cluster))

	if d.Get("wait_until_stable").(bool) {
		timeout, _ := time.ParseDuration(d.Get("wait_until_stable_timeout").(string))
		if err := waitTaskSetStable(ctx, conn, timeout, taskSetId, service, cluster); err != nil {
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
	partition := meta.(*conns.AWSClient).Partition

	taskSetId, service, cluster, err := TaskSetParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Set (%s): %s", d.Id(), err)
	}

	input := &ecs.DescribeTaskSetsInput{
		Cluster:  aws.String(cluster),
		Include:  []awstypes.TaskSetField{awstypes.TaskSetFieldTags},
		Service:  aws.String(service),
		TaskSets: []string{taskSetId},
	}

	out, err := conn.DescribeTaskSets(ctx, input)

	if !d.IsNewResource() && (errs.IsA[*awstypes.ClusterNotFoundException](err) || errs.IsA[*awstypes.ServiceNotFoundException](err) || errs.IsA[*awstypes.TaskSetNotFoundException](err)) {
		log.Printf("[WARN] ECS Task Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if errs.IsUnsupportedOperationInPartitionError(partition, err) {
		log.Printf("[WARN] ECS tagging failed describing Task Set (%s) with tags: %s; retrying without tags", d.Id(), err)

		input.Include = nil
		out, err = conn.DescribeTaskSets(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Set (%s): %s", d.Id(), err)
	}

	if out == nil || len(out.TaskSets) == 0 {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading ECS Task Set (%s): empty output after creation", d.Id())
		}
		log.Printf("[WARN] ECS Task Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	taskSet := out.TaskSets[0]

	d.Set(names.AttrARN, taskSet.TaskSetArn)
	d.Set("cluster", cluster)
	d.Set("launch_type", taskSet.LaunchType)
	d.Set("platform_version", taskSet.PlatformVersion)
	d.Set(names.AttrExternalID, taskSet.ExternalId)
	d.Set("service", service)
	d.Set(names.AttrStatus, taskSet.Status)
	d.Set("stability_status", taskSet.StabilityStatus)
	d.Set("task_definition", taskSet.TaskDefinition)
	d.Set("task_set_id", taskSet.Id)

	if err := d.Set(names.AttrCapacityProviderStrategy, flattenCapacityProviderStrategy(taskSet.CapacityProviderStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting capacity_provider_strategy: %s", err)
	}

	if err := d.Set("load_balancer", flattenTaskSetLoadBalancers(taskSet.LoadBalancers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting load_balancer: %s", err)
	}

	if err := d.Set(names.AttrNetworkConfiguration, flattenNetworkConfiguration(taskSet.NetworkConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_configuration: %s", err)
	}

	if err := d.Set("scale", flattenScale(taskSet.Scale)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting scale: %s", err)
	}

	if err := d.Set("service_registries", flattenServiceRegistries(taskSet.ServiceRegistries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting service_registries: %s", err)
	}

	setTagsOut(ctx, taskSet.Tags)

	return diags
}

func resourceTaskSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		taskSetId, service, cluster, err := TaskSetParseID(d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECS Task Set (%s): %s", d.Id(), err)
		}

		input := &ecs.UpdateTaskSetInput{
			Cluster: aws.String(cluster),
			Service: aws.String(service),
			TaskSet: aws.String(taskSetId),
			Scale:   expandScale(d.Get("scale").([]interface{})),
		}

		_, err = conn.UpdateTaskSet(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECS Task Set (%s): %s", d.Id(), err)
		}

		if d.Get("wait_until_stable").(bool) {
			timeout, _ := time.ParseDuration(d.Get("wait_until_stable_timeout").(string))
			if err := waitTaskSetStable(ctx, conn, timeout, taskSetId, service, cluster); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for ECS Task Set (%s) to be stable after update: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceTaskSetRead(ctx, d, meta)...)
}

func resourceTaskSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	taskSetId, service, cluster, err := TaskSetParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Task Set (%s): %s", d.Id(), err)
	}

	input := &ecs.DeleteTaskSetInput{
		Cluster: aws.String(cluster),
		Service: aws.String(service),
		TaskSet: aws.String(taskSetId),
		Force:   aws.Bool(d.Get(names.AttrForceDelete).(bool)),
	}

	_, err = conn.DeleteTaskSet(ctx, input)

	if errs.IsA[*awstypes.TaskSetNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Task Set (%s): %s", d.Id(), err)
	}

	if err := waitTaskSetDeleted(ctx, conn, taskSetId, service, cluster); err != nil {
		if errs.IsA[*awstypes.TaskSetNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting ECS Task Set (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func TaskSetParseID(id string) (string, string, string, error) {
	parts := strings.Split(id, ",")

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected TASK_SET_ID,SERVICE,CLUSTER", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func retryTaskSetCreate(ctx context.Context, conn *ecs.Client, input *ecs.CreateTaskSetInput) (*ecs.CreateTaskSetOutput, error) {
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout+taskSetCreateTimeout,
		func() (interface{}, error) {
			return conn.CreateTaskSet(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*awstypes.ClusterNotFoundException](err) || errs.IsA[*awstypes.ServiceNotFoundException](err) || errs.IsA[*awstypes.TaskSetNotFoundException](err) ||
				errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "does not have an associated load balancer") {
				return true, err
			}
			return false, err
		},
	)

	output, ok := outputRaw.(*ecs.CreateTaskSetOutput)
	if !ok || output == nil || output.TaskSet == nil {
		return nil, fmt.Errorf("creating ECS TaskSet: empty output")
	}

	return output, err
}
