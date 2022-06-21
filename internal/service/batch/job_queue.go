package batch

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceJobQueue() *schema.Resource {
	return &schema.Resource{
		Create: resourceJobQueueCreate,
		Read:   resourceJobQueueRead,
		Update: resourceJobQueueUpdate,
		Delete: resourceJobQueueDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceJobQueueCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BatchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	input := batch.CreateJobQueueInput{
		ComputeEnvironmentOrder: createComputeEnvironmentOrder(d.Get("compute_environments").([]interface{})),
		JobQueueName:            aws.String(d.Get("name").(string)),
		Priority:                aws.Int64(int64(d.Get("priority").(int))),
		State:                   aws.String(d.Get("state").(string)),
	}

	if v, ok := d.GetOk("scheduling_policy_arn"); ok {
		input.SchedulingPolicyArn = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	name := d.Get("name").(string)
	out, err := conn.CreateJobQueue(&input)
	if err != nil {
		return fmt.Errorf("%s %q", err, name)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{batch.JQStatusCreating, batch.JQStatusUpdating},
		Target:     []string{batch.JQStatusValid},
		Refresh:    jobQueueRefreshStatusFunc(conn, name),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for JobQueue state to be \"VALID\": %s", err)
	}

	arn := aws.StringValue(out.JobQueueArn)
	log.Printf("[DEBUG] JobQueue created: %s", arn)
	d.SetId(arn)

	return resourceJobQueueRead(d, meta)
}

func resourceJobQueueRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BatchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	jq, err := GetJobQueue(conn, d.Id())
	if err != nil {
		return err
	}
	if jq == nil {
		log.Printf("[WARN] Batch Job Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
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
		return fmt.Errorf("error setting compute_environments: %s", err)
	}

	d.Set("name", jq.JobQueueName)
	d.Set("priority", jq.Priority)
	d.Set("scheduling_policy_arn", jq.SchedulingPolicyArn)
	d.Set("state", jq.State)

	tags := KeyValueTags(jq.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceJobQueueUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BatchConn

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
				return fmt.Errorf("Cannot remove the fair share scheduling policy")
			}
		} else {
			// if a queue is a FIFO queue, SchedulingPolicyArn should not be set. Error is "Only fairshare queue can have scheduling policy"
			// hence, check for scheduling_policy_arn and set it in the inputs only if it exists already
			if v, ok := d.GetOk("scheduling_policy_arn"); ok {
				updateInput.SchedulingPolicyArn = aws.String(v.(string))
			}
		}

		_, err := conn.UpdateJobQueue(updateInput)
		if err != nil {
			return err
		}
		stateConf := &resource.StateChangeConf{
			Pending:    []string{batch.JQStatusUpdating},
			Target:     []string{batch.JQStatusValid},
			Refresh:    jobQueueRefreshStatusFunc(conn, name),
			Timeout:    10 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceJobQueueRead(d, meta)
}

func resourceJobQueueDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BatchConn
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Disabling Batch Job Queue %s", name)
	err := DisableJobQueue(name, conn)
	if err != nil {
		return fmt.Errorf("error disabling Batch Job Queue (%s): %s", name, err)
	}

	log.Printf("[DEBUG] Deleting Batch Job Queue %s", name)
	err = DeleteJobQueue(name, conn)
	if err != nil {
		return fmt.Errorf("error deleting Batch Job Queue (%s): %s", name, err)
	}

	return nil
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

func DeleteJobQueue(jobQueue string, conn *batch.Batch) error {
	_, err := conn.DeleteJobQueue(&batch.DeleteJobQueueInput{
		JobQueue: aws.String(jobQueue),
	})
	if err != nil {
		return err
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending:    []string{batch.JQStateDisabled, batch.JQStatusDeleting},
		Target:     []string{batch.JQStatusDeleted},
		Refresh:    jobQueueRefreshStatusFunc(conn, jobQueue),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateChangeConf.WaitForState()
	return err
}

func DisableJobQueue(jobQueue string, conn *batch.Batch) error {
	_, err := conn.UpdateJobQueue(&batch.UpdateJobQueueInput{
		JobQueue: aws.String(jobQueue),
		State:    aws.String(batch.JQStateDisabled),
	})
	if err != nil {
		return err
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending:    []string{batch.JQStatusUpdating},
		Target:     []string{batch.JQStatusValid},
		Refresh:    jobQueueRefreshStatusFunc(conn, jobQueue),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateChangeConf.WaitForState()
	return err
}

func GetJobQueue(conn *batch.Batch, sn string) (*batch.JobQueueDetail, error) {
	describeOpts := &batch.DescribeJobQueuesInput{
		JobQueues: []*string{aws.String(sn)},
	}
	resp, err := conn.DescribeJobQueues(describeOpts)
	if err != nil {
		return nil, err
	}

	numJobQueues := len(resp.JobQueues)
	switch {
	case numJobQueues == 0:
		log.Printf("[DEBUG] Job Queue %q is already gone", sn)
		return nil, nil
	case numJobQueues == 1:
		return resp.JobQueues[0], nil
	case numJobQueues > 1:
		return nil, fmt.Errorf("Multiple Job Queues with name %s", sn)
	}
	return nil, nil
}

func jobQueueRefreshStatusFunc(conn *batch.Batch, sn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ce, err := GetJobQueue(conn, sn)
		if err != nil {
			return nil, "failed", err
		}
		if ce == nil {
			return 42, batch.JQStatusDeleted, nil
		}
		return ce, *ce.Status, nil
	}
}
