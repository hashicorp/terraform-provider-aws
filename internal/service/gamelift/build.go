// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"
	"log"

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
func ResourceBuild() *schema.Resource {
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
			names.AttrVersion: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBuildCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	input := gamelift.CreateBuildInput{
		Name:            aws.String(d.Get(names.AttrName).(string)),
		OperatingSystem: awstypes.OperatingSystem(d.Get("operating_system").(string)),
		StorageLocation: expandStorageLocation(d.Get("storage_location").([]interface{})),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrVersion); ok {
		input.Version = aws.String(v.(string))
	}

	log.Printf("[INFO] Creating GameLift Build: %+v", input)
	var out *gamelift.CreateBuildOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		out, err = conn.CreateBuild(ctx, &input)
		if err != nil {
			if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Provided build is not accessible.") ||
				errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "GameLift cannot assume the role") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.CreateBuild(ctx, &input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift build client: %s", err)
	}

	d.SetId(aws.ToString(out.Build.BuildId))

	if _, err := waitBuildReady(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Build (%s) to ready: %s", d.Id(), err)
	}

	return append(diags, resourceBuildRead(ctx, d, meta)...)
}

func resourceBuildRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	build, err := FindBuildByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Build (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Build (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, build.Name)
	d.Set("operating_system", build.OperatingSystem)
	d.Set(names.AttrVersion, build.Version)

	arn := aws.ToString(build.BuildArn)
	d.Set(names.AttrARN, arn)

	return diags
}

func resourceBuildUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		log.Printf("[INFO] Updating GameLift Build: %s", d.Id())
		input := gamelift.UpdateBuildInput{
			BuildId: aws.String(d.Id()),
			Name:    aws.String(d.Get(names.AttrName).(string)),
		}
		if v, ok := d.GetOk(names.AttrVersion); ok {
			input.Version = aws.String(v.(string))
		}

		_, err := conn.UpdateBuild(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift build client: %s", err)
		}
	}

	return append(diags, resourceBuildRead(ctx, d, meta)...)
}

func resourceBuildDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Deleting GameLift Build: %s", d.Id())
	_, err := conn.DeleteBuild(ctx, &gamelift.DeleteBuildInput{
		BuildId: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Build Client (%s): %s", d.Id(), err)
	}
	return diags
}

func expandStorageLocation(cfg []interface{}) *awstypes.S3Location {
	loc := cfg[0].(map[string]interface{})

	location := &awstypes.S3Location{
		Bucket:  aws.String(loc[names.AttrBucket].(string)),
		Key:     aws.String(loc[names.AttrKey].(string)),
		RoleArn: aws.String(loc[names.AttrRoleARN].(string)),
	}

	if v, ok := loc["object_version"].(string); ok && v != "" {
		location.ObjectVersion = aws.String(v)
	}

	return location
}
