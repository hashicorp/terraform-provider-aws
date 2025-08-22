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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfio "github.com/hashicorp/terraform-provider-aws/internal/io"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const scriptMutex = `aws_gamelift_script`

// @SDKResource("aws_gamelift_script", name="Script")
// @Tags(identifierAttribute="arn")
func resourceScript() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScriptCreate,
		ReadWithoutTimeout:   resourceScriptRead,
		UpdateWithoutTimeout: resourceScriptUpdate,
		DeleteWithoutTimeout: resourceScriptDelete,

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
			"storage_location": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"zip_file", "storage_location"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucket: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						"object_version": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrRoleARN: {
							Type:         schema.TypeString,
							Required:     true,
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
			"zip_file": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"zip_file", "storage_location"},
			},
		},
	}
}

func resourceScriptCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &gamelift.CreateScriptInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("storage_location"); ok && len(v.([]any)) > 0 {
		input.StorageLocation = expandStorageLocation(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrVersion); ok {
		input.Version = aws.String(v.(string))
	}

	if v, ok := d.GetOk("zip_file"); ok {
		conns.GlobalMutexKV.Lock(scriptMutex)
		defer conns.GlobalMutexKV.Unlock(scriptMutex)

		file, err := tfio.ReadFileContents(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.ZipFile = file
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (any, error) {
			return conn.CreateScript(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "GameLift cannot assume the role") ||
				errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Provided resource is not accessible") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Script (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*gamelift.CreateScriptOutput).Script.ScriptId))

	return append(diags, resourceScriptRead(ctx, d, meta)...)
}

func resourceScriptRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	script, err := findScriptByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Script (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Script (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, script.ScriptArn)
	d.Set(names.AttrName, script.Name)
	if err := d.Set("storage_location", flattenStorageLocation(script.StorageLocation)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_location: %s", err)
	}
	d.Set(names.AttrVersion, script.Version)

	return diags
}

func resourceScriptUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &gamelift.UpdateScriptInput{
			Name:     aws.String(d.Get(names.AttrName).(string)),
			ScriptId: aws.String(d.Id()),
		}

		if d.HasChange("storage_location") {
			if v, ok := d.GetOk("storage_location"); ok {
				input.StorageLocation = expandStorageLocation(v.([]any))
			}
		}

		if d.HasChange(names.AttrVersion) {
			if v, ok := d.GetOk(names.AttrVersion); ok {
				input.Version = aws.String(v.(string))
			}
		}

		if d.HasChange("zip_file") {
			if v, ok := d.GetOk("zip_file"); ok {
				conns.GlobalMutexKV.Lock(scriptMutex)
				defer conns.GlobalMutexKV.Unlock(scriptMutex)

				file, err := tfio.ReadFileContents(v.(string))
				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				input.ZipFile = file
			}
		}

		_, err := conn.UpdateScript(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Script (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceScriptRead(ctx, d, meta)...)
}

func resourceScriptDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Deleting GameLift Script: %s", d.Id())
	_, err := conn.DeleteScript(ctx, &gamelift.DeleteScriptInput{
		ScriptId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Script (%s): %s", d.Id(), err)
	}

	return diags
}

func findScriptByID(ctx context.Context, conn *gamelift.Client, id string) (*awstypes.Script, error) {
	input := &gamelift.DescribeScriptInput{
		ScriptId: aws.String(id),
	}

	output, err := conn.DescribeScript(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Script == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Script, nil
}

func flattenStorageLocation(apiObject *awstypes.S3Location) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrBucket:  aws.ToString(apiObject.Bucket),
		names.AttrKey:     aws.ToString(apiObject.Key),
		"object_version":  aws.ToString(apiObject.ObjectVersion),
		names.AttrRoleARN: aws.ToString(apiObject.RoleArn),
	}

	return []any{tfMap}
}
