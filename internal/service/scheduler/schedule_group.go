// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	awstypes "github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_scheduler_schedule_group", name="Schedule Group")
// @Tags(identifierAttribute="arn")
func resourceScheduleGroup() *schema.Resource {
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
					validation.StringLenBetween(1, 64-sdkid.UniqueIDSuffixLength),
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

func resourceScheduleGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	name := create.Name(ctx, d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	in := scheduler.CreateScheduleGroupInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	_, err := conn.CreateScheduleGroup(ctx, &in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Scheduler Schedule Group (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitScheduleGroupActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Scheduler Schedule Group (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceScheduleGroupRead(ctx, d, meta)...)
}

func resourceScheduleGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	out, err := findScheduleGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EventBridge Scheduler Schedule Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Scheduler Schedule Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrCreationDate, aws.ToTime(out.CreationDate).Format(time.RFC3339))
	d.Set("last_modification_date", aws.ToTime(out.LastModificationDate).Format(time.RFC3339))
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(out.Name)))
	d.Set(names.AttrState, out.State)

	return diags
}

func resourceScheduleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceScheduleGroupRead(ctx, d, meta)
}

func resourceScheduleGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchedulerClient(ctx)

	log.Printf("[INFO] Deleting EventBridge Scheduler ScheduleGroup: %s", d.Id())
	in := scheduler.DeleteScheduleGroupInput{
		Name: aws.String(d.Id()),
	}
	_, err := conn.DeleteScheduleGroup(ctx, &in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Scheduler Schedule Group (%s): %s", d.Id(), err)
	}

	if _, err := waitScheduleGroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Scheduler Schedule Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findScheduleGroupByName(ctx context.Context, conn *scheduler.Client, name string) (*scheduler.GetScheduleGroupOutput, error) {
	in := scheduler.GetScheduleGroupInput{
		Name: aws.String(name),
	}

	return findScheduleGroup(ctx, conn, &in)
}

func findScheduleGroup(ctx context.Context, conn *scheduler.Client, input *scheduler.GetScheduleGroupInput) (*scheduler.GetScheduleGroupOutput, error) {
	output, err := conn.GetScheduleGroup(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Arn == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusScheduleGroup(conn *scheduler.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findScheduleGroupByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func waitScheduleGroupActive(ctx context.Context, conn *scheduler.Client, name string, timeout time.Duration) (*scheduler.GetScheduleGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.ScheduleGroupStateActive),
		Refresh:                   statusScheduleGroup(conn, name),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*scheduler.GetScheduleGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func waitScheduleGroupDeleted(ctx context.Context, conn *scheduler.Client, name string, timeout time.Duration) (*scheduler.GetScheduleGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScheduleGroupStateDeleting, awstypes.ScheduleGroupStateActive),
		Target:  []string{},
		Refresh: statusScheduleGroup(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*scheduler.GetScheduleGroupOutput); ok {
		return out, err
	}

	return nil, err
}
