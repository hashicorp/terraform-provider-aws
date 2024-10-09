// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloud9

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloud9"
	"github.com/aws/aws-sdk-go-v2/service/cloud9/types"
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

// @SDKResource("aws_cloud9_environment_ec2", name="Environment EC2")
// @Tags(identifierAttribute="arn")
func resourceEnvironmentEC2() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEnvironmentEC2Create,
		ReadWithoutTimeout:   resourceEnvironmentEC2Read,
		UpdateWithoutTimeout: resourceEnvironmentEC2Update,
		DeleteWithoutTimeout: resourceEnvironmentEC2Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic_stop_time_minutes": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtMost(20160),
			},
			"connection_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.ConnectionTypeConnectSsh,
				ValidateDiagFunc: enum.Validate[types.ConnectionType](),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 200),
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"amazonlinux-1-x86_64",
					"amazonlinux-2-x86_64",
					"amazonlinux-2023-x86_64",
					"ubuntu-18.04-x86_64",
					"ubuntu-22.04-x86_64",
					"resolve:ssm:/aws/service/cloud9/amis/amazonlinux-1-x86_64",
					"resolve:ssm:/aws/service/cloud9/amis/amazonlinux-2-x86_64",
					"resolve:ssm:/aws/service/cloud9/amis/amazonlinux-2023-x86_64",
					"resolve:ssm:/aws/service/cloud9/amis/ubuntu-18.04-x86_64",
					"resolve:ssm:/aws/service/cloud9/amis/ubuntu-22.04-x86_64",
				}, false),
			},
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 60),
			},
			"owner_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEnvironmentEC2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Client(ctx)

	name := d.Get(names.AttrName).(string)
	input := &cloud9.CreateEnvironmentEC2Input{
		ClientRequestToken: aws.String(id.UniqueId()),
		ConnectionType:     types.ConnectionType(d.Get("connection_type").(string)),
		ImageId:            aws.String(d.Get("image_id").(string)),
		InstanceType:       aws.String(d.Get(names.AttrInstanceType).(string)),
		Name:               aws.String(name),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("automatic_stop_time_minutes"); ok {
		input.AutomaticStopTimeMinutes = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("owner_arn"); ok {
		input.OwnerArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSubnetID); ok {
		input.SubnetId = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*types.NotFoundException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateEnvironmentEC2(ctx, input)
	}, "User")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cloud9 EC2 Environment (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*cloud9.CreateEnvironmentEC2Output).EnvironmentId))

	if _, err := waitEnvironmentReady(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Cloud9 EC2 Environment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEnvironmentEC2Read(ctx, d, meta)...)
}

func resourceEnvironmentEC2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Client(ctx)

	env, err := findEnvironmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cloud9 EC2 Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cloud9 EC2 Environment (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, env.Arn)
	d.Set("connection_type", env.ConnectionType)
	d.Set(names.AttrDescription, env.Description)
	d.Set(names.AttrName, env.Name)
	d.Set("owner_arn", env.OwnerArn)
	d.Set(names.AttrType, env.Type)

	return diags
}

func resourceEnvironmentEC2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := cloud9.UpdateEnvironmentInput{
			Description:   aws.String(d.Get(names.AttrDescription).(string)),
			EnvironmentId: aws.String(d.Id()),
			Name:          aws.String(d.Get(names.AttrName).(string)),
		}

		_, err := conn.UpdateEnvironment(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Cloud9 EC2 Environment (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceEnvironmentEC2Read(ctx, d, meta)...)
}

func resourceEnvironmentEC2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Client(ctx)

	log.Printf("[INFO] Deleting Cloud9 EC2 Environment: %s", d.Id())
	_, err := conn.DeleteEnvironment(ctx, &cloud9.DeleteEnvironmentInput{
		EnvironmentId: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cloud9 EC2 Environment (%s): %s", d.Id(), err)
	}

	if _, err := waitEnvironmentDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Cloud9 EC2 Environment (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findEnvironment(ctx context.Context, conn *cloud9.Client, input *cloud9.DescribeEnvironmentsInput) (*types.Environment, error) {
	output, err := findEnvironments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	environment, err := tfresource.AssertSingleValueResult(output)

	if err != nil {
		return nil, err
	}

	if environment.Lifecycle == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return environment, nil
}

func findEnvironments(ctx context.Context, conn *cloud9.Client, input *cloud9.DescribeEnvironmentsInput) ([]types.Environment, error) {
	output, err := conn.DescribeEnvironments(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
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

	return output.Environments, nil
}

func findEnvironmentByID(ctx context.Context, conn *cloud9.Client, id string) (*types.Environment, error) {
	input := &cloud9.DescribeEnvironmentsInput{
		EnvironmentIds: []string{id},
	}

	output, err := findEnvironment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.Id) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusEnvironmentStatus(ctx context.Context, conn *cloud9.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findEnvironmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Lifecycle.Status), nil
	}
}

func waitEnvironmentReady(ctx context.Context, conn *cloud9.Client, id string) (*types.Environment, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EnvironmentLifecycleStatusCreating),
		Target:  enum.Slice(types.EnvironmentLifecycleStatusCreated),
		Refresh: statusEnvironmentStatus(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Environment); ok {
		if lifecycle := output.Lifecycle; lifecycle.Status == types.EnvironmentLifecycleStatusCreateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(lifecycle.Reason)))
		}

		return output, err
	}

	return nil, err
}

func waitEnvironmentDeleted(ctx context.Context, conn *cloud9.Client, id string) (*types.Environment, error) {
	const (
		timeout = 20 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EnvironmentLifecycleStatusDeleting),
		Target:  []string{},
		Refresh: statusEnvironmentStatus(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Environment); ok {
		if lifecycle := output.Lifecycle; lifecycle.Status == types.EnvironmentLifecycleStatusDeleteFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(lifecycle.Reason)))
		}

		return output, err
	}

	return nil, err
}
