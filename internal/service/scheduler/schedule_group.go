// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package scheduler

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_scheduler_schedule_group", name="Schedule Group")
// @Tags(identifierAttribute="arn")
func ResourceScheduleGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScheduleGroupCreate,
		ReadWithoutTimeout:   resourceScheduleGroupRead,
		UpdateWithoutTimeout: resourceScheduleGroupUpdate,
		DeleteWithoutTimeout: resourceScheduleGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modification_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_.]+$`), `The name must consist of alphanumerics, hyphens, and underscores.`),
				)),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64-id.UniqueIDSuffixLength),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-_.]+$`), `The name must consist of alphanumerics, hyphens, and underscores.`),
				)),
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	ResNameScheduleGroup = "Schedule Group"
)

func resourceScheduleGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))

	in := &scheduler.CreateScheduleGroupInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	out, err := conn.CreateScheduleGroup(ctx, in)
	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionCreating, ResNameScheduleGroup, name, err)
	}

	if out == nil || out.ScheduleGroupArn == nil {
		return create.DiagError(names.Scheduler, create.ErrActionCreating, ResNameScheduleGroup, name, errors.New("empty output"))
	}

	d.SetId(name)

	if _, err := waitScheduleGroupActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionWaitingForCreation, ResNameScheduleGroup, d.Id(), err)
	}

	return resourceScheduleGroupRead(ctx, d, meta)
}

func resourceScheduleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	out, err := findScheduleGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Scheduler Schedule Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionReading, ResNameScheduleGroup, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("creation_date", aws.ToTime(out.CreationDate).Format(time.RFC3339))
	d.Set("last_modification_date", aws.ToTime(out.LastModificationDate).Format(time.RFC3339))
	d.Set("name", out.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.ToString(out.Name)))
	d.Set("state", out.State)

	return nil
}

func resourceScheduleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceScheduleGroupRead(ctx, d, meta)
}

func resourceScheduleGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	log.Printf("[INFO] Deleting EventBridge Scheduler ScheduleGroup %s", d.Id())

	_, err := conn.DeleteScheduleGroup(ctx, &scheduler.DeleteScheduleGroupInput{
		Name: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.Scheduler, create.ErrActionDeleting, ResNameScheduleGroup, d.Id(), err)
	}

	if _, err := waitScheduleGroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.Scheduler, create.ErrActionWaitingForDeletion, ResNameScheduleGroup, d.Id(), err)
	}

	return nil
}
