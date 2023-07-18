// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_batch_job_queue", name="Job Queue")
// @Tags(identifierAttribute="arn")
func ResourceJobQueue() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceJobQueueCreate,
		ReadWithoutTimeout:   resourceJobQueueRead,
		UpdateWithoutTimeout: resourceJobQueueUpdate,
		DeleteWithoutTimeout: resourceJobQueueDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("arn", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"compute_environments": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},
			"priority": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"scheduling_policy_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"state": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{batch.JQStateDisabled, batch.JQStateEnabled}, true),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceJobQueueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	input := batch.CreateJobQueueInput{
		ComputeEnvironmentOrder: createComputeEnvironmentOrder(d.Get("compute_environments").([]interface{})),
		JobQueueName:            aws.String(d.Get("name").(string)),
		Priority:                aws.Int64(int64(d.Get("priority").(int))),
		State:                   aws.String(d.Get("state").(string)),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("scheduling_policy_arn"); ok {
		input.SchedulingPolicyArn = aws.String(v.(string))
	}

	name := d.Get("name").(string)
	out, err := conn.CreateJobQueueWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "%s %q", err, name)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{batch.JQStatusCreating, batch.JQStatusUpdating},
		Target:     []string{batch.JQStatusValid},
		Refresh:    jobQueueRefreshStatusFunc(ctx, conn, name),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for JobQueue state to be \"VALID\": %s", err)
	}

	arn := aws.StringValue(out.JobQueueArn)
	log.Printf("[DEBUG] JobQueue created: %s", arn)
	d.SetId(arn)

	return append(diags, resourceJobQueueRead(ctx, d, meta)...)
}

func resourceJobQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	jq, err := GetJobQueue(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Batch Job Queue (%s): %s", d.Get("name").(string), err)
	}
	if jq == nil {
		log.Printf("[WARN] Batch Job Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("arn", jq.JobQueueArn)

	computeEnvironments := make([]string, 0, len(jq.ComputeEnvironmentOrder))

	sort.Slice(jq.ComputeEnvironmentOrder, func(i, j int) bool {
		return aws.Int64Value(jq.ComputeEnvironmentOrder[i].Order) < aws.Int64Value(jq.ComputeEnvironmentOrder[j].Order)
	})

	for _, computeEnvironmentOrder := range jq.ComputeEnvironmentOrder {
		computeEnvironments = append(computeEnvironments, aws.StringValue(computeEnvironmentOrder.ComputeEnvironment))
	}

	if err := d.Set("compute_environments", computeEnvironments); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting compute_environments: %s", err)
	}

	d.Set("name", jq.JobQueueName)
	d.Set("priority", jq.Priority)
	d.Set("scheduling_policy_arn", jq.SchedulingPolicyArn)
	d.Set("state", jq.State)

	setTagsOut(ctx, jq.Tags)

	return diags
}

func resourceJobQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	if d.HasChanges("compute_environments", "priority", "scheduling_policy_arn", "state") {
		name := d.Get("name").(string)
		updateInput := &batch.UpdateJobQueueInput{
			ComputeEnvironmentOrder: createComputeEnvironmentOrder(d.Get("compute_environments").([]interface{})),
			JobQueue:                aws.String(name),
			Priority:                aws.Int64(int64(d.Get("priority").(int))),
			State:                   aws.String(d.Get("state").(string)),
		}
		// After a job queue is created, you can replace but can't remove the fair share scheduling policy
		// https://docs.aws.amazon.com/sdk-for-go/api/service/batch/#CreateJobQueueInput
		if d.HasChange("scheduling_policy_arn") {
			if v, ok := d.GetOk("scheduling_policy_arn"); ok {
				updateInput.SchedulingPolicyArn = aws.String(v.(string))
			} else {
				return sdkdiag.AppendErrorf(diags, "Cannot remove the fair share scheduling policy")
			}
		} else {
			// if a queue is a FIFO queue, SchedulingPolicyArn should not be set. Error is "Only fairshare queue can have scheduling policy"
			// hence, check for scheduling_policy_arn and set it in the inputs only if it exists already
			if v, ok := d.GetOk("scheduling_policy_arn"); ok {
				updateInput.SchedulingPolicyArn = aws.String(v.(string))
			}
		}

		_, err := conn.UpdateJobQueueWithContext(ctx, updateInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Batch Job Queue (%s): %s", d.Get("name").(string), err)
		}
		stateConf := &retry.StateChangeConf{
			Pending:    []string{batch.JQStatusUpdating},
			Target:     []string{batch.JQStatusValid},
			Refresh:    jobQueueRefreshStatusFunc(ctx, conn, name),
			Timeout:    10 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Batch Job Queue (%s): waiting for completion: %s", d.Get("name").(string), err)
		}
	}

	return append(diags, resourceJobQueueRead(ctx, d, meta)...)
}

func resourceJobQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn(ctx)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Disabling Batch Job Queue: %s", name)
	err := DisableJobQueue(ctx, name, conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Batch Job Queue (%s): %s", name, err)
	}

	log.Printf("[DEBUG] Deleting Batch Job Queue: %s", name)
	err = DeleteJobQueue(ctx, name, conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Batch Job Queue (%s): %s", name, err)
	}

	return diags
}

func createComputeEnvironmentOrder(order []interface{}) (envs []*batch.ComputeEnvironmentOrder) {
	for i, env := range order {
		envs = append(envs, &batch.ComputeEnvironmentOrder{
			Order:              aws.Int64(int64(i)),
			ComputeEnvironment: aws.String(env.(string)),
		})
	}
	return
}

func DeleteJobQueue(ctx context.Context, jobQueue string, conn *batch.Batch) error {
	_, err := conn.DeleteJobQueueWithContext(ctx, &batch.DeleteJobQueueInput{
		JobQueue: aws.String(jobQueue),
	})
	if err != nil {
		return err
	}

	stateChangeConf := &retry.StateChangeConf{
		Pending:    []string{batch.JQStateDisabled, batch.JQStatusDeleting},
		Target:     []string{batch.JQStatusDeleted},
		Refresh:    jobQueueRefreshStatusFunc(ctx, conn, jobQueue),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateChangeConf.WaitForStateContext(ctx)
	return err
}

func DisableJobQueue(ctx context.Context, jobQueue string, conn *batch.Batch) error {
	_, err := conn.UpdateJobQueueWithContext(ctx, &batch.UpdateJobQueueInput{
		JobQueue: aws.String(jobQueue),
		State:    aws.String(batch.JQStateDisabled),
	})
	if err != nil {
		return err
	}

	stateChangeConf := &retry.StateChangeConf{
		Pending:    []string{batch.JQStatusUpdating},
		Target:     []string{batch.JQStatusValid},
		Refresh:    jobQueueRefreshStatusFunc(ctx, conn, jobQueue),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateChangeConf.WaitForStateContext(ctx)
	return err
}

func GetJobQueue(ctx context.Context, conn *batch.Batch, sn string) (*batch.JobQueueDetail, error) {
	describeOpts := &batch.DescribeJobQueuesInput{
		JobQueues: []*string{aws.String(sn)},
	}
	resp, err := conn.DescribeJobQueuesWithContext(ctx, describeOpts)
	if err != nil {
		return nil, err
	}

	numJobQueues := len(resp.JobQueues)
	switch {
	case numJobQueues == 0:
		return nil, nil
	case numJobQueues == 1:
		return resp.JobQueues[0], nil
	case numJobQueues > 1:
		return nil, fmt.Errorf("Multiple Job Queues with name %s", sn)
	}
	return nil, nil
}

func jobQueueRefreshStatusFunc(ctx context.Context, conn *batch.Batch, sn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ce, err := GetJobQueue(ctx, conn, sn)
		if err != nil {
			return nil, "failed", err
		}
		if ce == nil {
			return 42, batch.JQStatusDeleted, nil
		}
		return ce, *ce.Status, nil
	}
}
