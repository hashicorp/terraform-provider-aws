// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_provider_strategy": {
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
						"weight": {
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
			"external_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"force_delete": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"launch_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ValidateFunc:  validation.StringInSlice(ecs.LaunchType_Values(), false),
				ConflictsWith: []string{"capacity_provider_strategy"},
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
			"network_configuration": {
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
						"security_groups": {
							Type:     schema.TypeSet,
							MaxItems: 5,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnets": {
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
						"unit": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      ecs.ScaleUnitPercent,
							ValidateFunc: validation.StringInSlice(ecs.ScaleUnit_Values(), false),
						},
						"value": {
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
						"port": {
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
			"status": {
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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	cluster := d.Get("cluster").(string)
	service := d.Get("service").(string)
	input := &ecs.CreateTaskSetInput{
		ClientToken:    aws.String(id.UniqueId()),
		Cluster:        aws.String(cluster),
		Service:        aws.String(service),
		Tags:           getTagsIn(ctx),
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
	}

	if v, ok := d.GetOk("capacity_provider_strategy"); ok && v.(*schema.Set).Len() > 0 {
		input.CapacityProviderStrategy = expandCapacityProviderStrategy(v.(*schema.Set))
	}

	if v, ok := d.GetOk("external_id"); ok {
		input.ExternalId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("launch_type"); ok {
		input.LaunchType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("load_balancer"); ok && v.(*schema.Set).Len() > 0 {
		input.LoadBalancers = expandTaskSetLoadBalancers(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("network_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
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
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = retryTaskSetCreate(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS TaskSet: %s", err)
	}

	taskSetId := aws.StringValue(output.TaskSet.Id)

	d.SetId(fmt.Sprintf("%s,%s,%s", taskSetId, service, cluster))

	if d.Get("wait_until_stable").(bool) {
		timeout, _ := time.ParseDuration(d.Get("wait_until_stable_timeout").(string))
		if err := waitTaskSetStable(ctx, conn, timeout, taskSetId, service, cluster); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ECS Task Set (%s) create: %s", d.Id(), err)
		}
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, aws.StringValue(output.TaskSet.TaskSetArn), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	taskSetId, service, cluster, err := TaskSetParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Set (%s): %s", d.Id(), err)
	}

	input := &ecs.DescribeTaskSetsInput{
		Cluster:  aws.String(cluster),
		Include:  aws.StringSlice([]string{ecs.TaskSetFieldTags}),
		Service:  aws.String(service),
		TaskSets: aws.StringSlice([]string{taskSetId}),
	}

	out, err := conn.DescribeTaskSetsWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException, ecs.ErrCodeServiceNotFoundException, ecs.ErrCodeTaskSetNotFoundException) {
		log.Printf("[WARN] ECS Task Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] ECS tagging failed describing Task Set (%s) with tags: %s; retrying without tags", d.Id(), err)

		input.Include = nil
		out, err = conn.DescribeTaskSetsWithContext(ctx, input)
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

	d.Set("arn", taskSet.TaskSetArn)
	d.Set("cluster", cluster)
	d.Set("launch_type", taskSet.LaunchType)
	d.Set("platform_version", taskSet.PlatformVersion)
	d.Set("external_id", taskSet.ExternalId)
	d.Set("service", service)
	d.Set("status", taskSet.Status)
	d.Set("stability_status", taskSet.StabilityStatus)
	d.Set("task_definition", taskSet.TaskDefinition)
	d.Set("task_set_id", taskSet.Id)

	if err := d.Set("capacity_provider_strategy", flattenCapacityProviderStrategy(taskSet.CapacityProviderStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting capacity_provider_strategy: %s", err)
	}

	if err := d.Set("load_balancer", flattenTaskSetLoadBalancers(taskSet.LoadBalancers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting load_balancer: %s", err)
	}

	if err := d.Set("network_configuration", flattenNetworkConfiguration(taskSet.NetworkConfiguration)); err != nil {
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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
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

		_, err = conn.UpdateTaskSetWithContext(ctx, input)

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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	taskSetId, service, cluster, err := TaskSetParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Task Set (%s): %s", d.Id(), err)
	}

	input := &ecs.DeleteTaskSetInput{
		Cluster: aws.String(cluster),
		Service: aws.String(service),
		TaskSet: aws.String(taskSetId),
		Force:   aws.Bool(d.Get("force_delete").(bool)),
	}

	_, err = conn.DeleteTaskSetWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ecs.ErrCodeTaskSetNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Task Set (%s): %s", d.Id(), err)
	}

	if err := waitTaskSetDeleted(ctx, conn, taskSetId, service, cluster); err != nil {
		if tfawserr.ErrCodeEquals(err, ecs.ErrCodeTaskSetNotFoundException) {
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

func retryTaskSetCreate(ctx context.Context, conn *ecs.ECS, input *ecs.CreateTaskSetInput) (*ecs.CreateTaskSetOutput, error) {
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout+taskSetCreateTimeout,
		func() (interface{}, error) {
			return conn.CreateTaskSetWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException, ecs.ErrCodeServiceNotFoundException, ecs.ErrCodeTaskSetNotFoundException) ||
				tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "does not have an associated load balancer") {
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
