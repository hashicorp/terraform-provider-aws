// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package scheduler

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modification_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), `The name must consist of alphanumerics, hyphens, and underscores.`),
				)),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 64-id.UniqueIDSuffixLength),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), `The name must consist of alphanumerics, hyphens, and underscores.`),
				)),
			},
			names.AttrState: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))

	in := &scheduler.CreateScheduleGroupInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	out, err := conn.CreateScheduleGroup(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.Scheduler, create.ErrActionCreating, ResNameScheduleGroup, name, err)
	}

	if out == nil || out.ScheduleGroupArn == nil {
		return create.AppendDiagError(diags, names.Scheduler, create.ErrActionCreating, ResNameScheduleGroup, name, errors.New("empty output"))
	}

	d.SetId(name)

	if _, err := waitScheduleGroupActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.Scheduler, create.ErrActionWaitingForCreation, ResNameScheduleGroup, d.Id(), err)
	}

	return append(diags, resourceScheduleGroupRead(ctx, d, meta)...)
}

func resourceScheduleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	out, err := findScheduleGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Scheduler Schedule Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Scheduler, create.ErrActionReading, ResNameScheduleGroup, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrCreationDate, aws.ToTime(out.CreationDate).Format(time.RFC3339))
	d.Set("last_modification_date", aws.ToTime(out.LastModificationDate).Format(time.RFC3339))
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(out.Name)))
	d.Set(names.AttrState, out.State)

	return diags
}

func resourceScheduleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceScheduleGroupRead(ctx, d, meta)
}

func resourceScheduleGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	log.Printf("[INFO] Deleting EventBridge Scheduler ScheduleGroup %s", d.Id())

	_, err := conn.DeleteScheduleGroup(ctx, &scheduler.DeleteScheduleGroupInput{
		Name: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.Scheduler, create.ErrActionDeleting, ResNameScheduleGroup, d.Id(), err)
	}

	if _, err := waitScheduleGroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.Scheduler, create.ErrActionWaitingForDeletion, ResNameScheduleGroup, d.Id(), err)
	}

	return diags
}
