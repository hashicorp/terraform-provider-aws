// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/mitchellh/go-homedir"
)

const scriptMutex = `aws_gamelift_script`

// @SDKResource("aws_gamelift_script", name="Script")
// @Tags(identifierAttribute="arn")
func ResourceScript() *schema.Resource {
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
				ForceNew:     true,
				Computed:     true,
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
			names.AttrVersion: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"zip_file": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"zip_file", "storage_location"},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceScriptCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	input := gamelift.CreateScriptInput{
		Name: aws.String(d.Get(names.AttrName).(string)),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("storage_location"); ok && len(v.([]interface{})) > 0 {
		input.StorageLocation = expandStorageLocation(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrVersion); ok {
		input.Version = aws.String(v.(string))
	}

	if v, ok := d.GetOk("zip_file"); ok {
		conns.GlobalMutexKV.Lock(scriptMutex)
		defer conns.GlobalMutexKV.Unlock(scriptMutex)

		file, err := loadFileContent(v.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "unable to load %q: %s", v.(string), err)
		}
		input.ZipFile = file
	}

	log.Printf("[INFO] Creating GameLift Script: %s", input)
	var out *gamelift.CreateScriptOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		out, err = conn.CreateScriptWithContext(ctx, &input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, "GameLift cannot assume the role") ||
				tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, "Provided resource is not accessible") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.CreateScriptWithContext(ctx, &input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift script client: %s", err)
	}

	d.SetId(aws.StringValue(out.Script.ScriptId))

	return append(diags, resourceScriptRead(ctx, d, meta)...)
}

func resourceScriptRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	log.Printf("[INFO] Reading GameLift Script: %s", d.Id())
	script, err := FindScriptByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Script (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Script (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, script.Name)
	d.Set(names.AttrVersion, script.Version)

	if err := d.Set("storage_location", flattenStorageLocation(script.StorageLocation)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_location: %s", err)
	}

	arn := aws.StringValue(script.ScriptArn)
	d.Set(names.AttrARN, arn)

	return diags
}

func resourceScriptUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		log.Printf("[INFO] Updating GameLift Script: %s", d.Id())
		input := gamelift.UpdateScriptInput{
			ScriptId: aws.String(d.Id()),
			Name:     aws.String(d.Get(names.AttrName).(string)),
		}

		if d.HasChange(names.AttrVersion) {
			if v, ok := d.GetOk(names.AttrVersion); ok {
				input.Version = aws.String(v.(string))
			}
		}

		if d.HasChange("storage_location") {
			if v, ok := d.GetOk("storage_location"); ok {
				input.StorageLocation = expandStorageLocation(v.([]interface{}))
			}
		}

		if d.HasChange("zip_file") {
			if v, ok := d.GetOk("zip_file"); ok {
				conns.GlobalMutexKV.Lock(scriptMutex)
				defer conns.GlobalMutexKV.Unlock(scriptMutex)

				file, err := loadFileContent(v.(string))
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "unable to load %q: %s", v.(string), err)
				}
				input.ZipFile = file
			}
		}

		_, err := conn.UpdateScriptWithContext(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Script: %s", err)
		}
	}

	return append(diags, resourceScriptRead(ctx, d, meta)...)
}

func resourceScriptDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	log.Printf("[INFO] Deleting GameLift Script: %s", d.Id())
	_, err := conn.DeleteScriptWithContext(ctx, &gamelift.DeleteScriptInput{
		ScriptId: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting GameLift script: %s", err)
	}

	return diags
}

func flattenStorageLocation(sl *gamelift.S3Location) []interface{} {
	if sl == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrBucket:  aws.StringValue(sl.Bucket),
		names.AttrKey:     aws.StringValue(sl.Key),
		names.AttrRoleARN: aws.StringValue(sl.RoleArn),
		"object_version":  aws.StringValue(sl.ObjectVersion),
	}

	return []interface{}{m}
}

// loadFileContent returns contents of a file in a given path
func loadFileContent(v string) ([]byte, error) {
	filename, err := homedir.Expand(v)
	if err != nil {
		return nil, err
	}
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return fileContent, nil
}
