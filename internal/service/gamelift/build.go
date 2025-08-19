// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

// @SDKResource("aws_gamelift_build", name="Build")
// @Tags(identifierAttribute="arn")
func resourceBuild() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBuildCreate,
		ReadWithoutTimeout:   resourceBuildRead,
		UpdateWithoutTimeout: resourceBuildUpdate,
		DeleteWithoutTimeout: resourceBuildDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"operating_system": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.OperatingSystem](),
			},
			"storage_location": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucket: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"object_version": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrRoleARN: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
		},
	}
}

func resourceBuildCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &gamelift.CreateBuildInput{
		Name:            aws.String(name),
		OperatingSystem: awstypes.OperatingSystem(d.Get("operating_system").(string)),
		StorageLocation: expandStorageLocation(d.Get("storage_location").([]any)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrVersion); ok {
		input.Version = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (any, error) {
			return conn.CreateBuild(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Provided build is not accessible.") ||
				errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "GameLift cannot assume the role") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Build (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*gamelift.CreateBuildOutput).Build.BuildId))

	if _, err := waitBuildReady(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Build (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBuildRead(ctx, d, meta)...)
}

func resourceBuildRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	build, err := findBuildByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Build (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Build (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, build.BuildArn)
	d.Set(names.AttrName, build.Name)
	d.Set("operating_system", build.OperatingSystem)
	d.Set(names.AttrVersion, build.Version)

	return diags
}

func resourceBuildUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &gamelift.UpdateBuildInput{
			BuildId: aws.String(d.Id()),
			Name:    aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk(names.AttrVersion); ok {
			input.Version = aws.String(v.(string))
		}

		_, err := conn.UpdateBuild(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Build (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceBuildRead(ctx, d, meta)...)
}

func resourceBuildDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Deleting GameLift Build: %s", d.Id())
	_, err := conn.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
		BuildId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Build (%s): %s", d.Id(), err)
	}

	return diags
}

func findBuildByID(ctx context.Context, conn *gamelift.Client, id string) (*awstypes.Build, error) {
	input := &gamelift.DescribeBuildInput{
		BuildId: aws.String(id),
	}

	output, err := conn.DescribeBuild(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Build == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Build, nil
}

func statusBuild(ctx context.Context, conn *gamelift.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findBuildByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitBuildReady(ctx context.Context, conn *gamelift.Client, id string) (*awstypes.Build, error) {
	const (
		timeout = 1 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BuildStatusInitialized),
		Target:  enum.Slice(awstypes.BuildStatusReady),
		Refresh: statusBuild(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Build); ok {
		return output, err
	}

	return nil, err
}

func expandStorageLocation(tfList []any) *awstypes.S3Location {
	tfMap := tfList[0].(map[string]any)

	apiObject := &awstypes.S3Location{
		Bucket:  aws.String(tfMap[names.AttrBucket].(string)),
		Key:     aws.String(tfMap[names.AttrKey].(string)),
		RoleArn: aws.String(tfMap[names.AttrRoleARN].(string)),
	}

	if v, ok := tfMap["object_version"].(string); ok && v != "" {
		apiObject.ObjectVersion = aws.String(v)
	}

	return apiObject
}
