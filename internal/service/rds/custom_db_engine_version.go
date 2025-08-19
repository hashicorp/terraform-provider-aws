// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfio "github.com/hashicorp/terraform-provider-aws/internal/io"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_custom_db_engine_version", name="Custom DB Engine Version")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceCustomDBEngineVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomDBEngineVersionCreate,
		ReadWithoutTimeout:   resourceCustomDBEngineVersionRead,
		UpdateWithoutTimeout: resourceCustomDBEngineVersionUpdate,
		DeleteWithoutTimeout: resourceCustomDBEngineVersionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(240 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_installation_files_s3_bucket_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			"database_installation_files_s3_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"db_parameter_group_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(fmt.Sprintf(`^%s.*$`, InstanceEngineCustomPrefix)), fmt.Sprintf("must begin with %s", InstanceEngineCustomPrefix)),
					validation.StringLenBetween(1, 35),
				),
			},
			names.AttrEngineVersion: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 60),
			},
			"filename": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"manifest"},
			},
			//API returns created image_id of the newly created image.
			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"major_engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"manifest": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringIsJSON,
					validation.StringLenBetween(1, 100000),
				),
				ConflictsWith:    []string{"filename"},
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			//API returns manifest with service added additions, non-determinestic.
			"manifest_computed": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"manifest_hash": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.CustomEngineVersionStatus](),
			},
			// Allow CEV creation from a source AMI ID.
			// implicit state passthrough, virtual attribute
			"source_image_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCustomDBEngineVersionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	const (
		mutexKey = `aws_rds_custom_engine_version`
	)
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds.CreateCustomDBEngineVersionInput{
		Engine:        aws.String(d.Get(names.AttrEngine).(string)),
		EngineVersion: aws.String(d.Get(names.AttrEngineVersion).(string)),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("database_installation_files_s3_bucket_name"); ok {
		input.DatabaseInstallationFilesS3BucketName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("database_installation_files_s3_prefix"); ok {
		input.DatabaseInstallationFilesS3Prefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("filename"); ok {
		// Grab an exclusive lock so that we're only reading one contact flow into
		// memory at a time.
		// See https://github.com/hashicorp/terraform/issues/9364
		conns.GlobalMutexKV.Lock(mutexKey)
		defer conns.GlobalMutexKV.Unlock(mutexKey)

		file, err := tfio.ReadFileContents(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.Manifest = aws.String(string(file))
	} else if v, ok := d.GetOk("manifest"); ok {
		input.Manifest = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KMSKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_image_id"); ok {
		input.ImageId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("filename"); ok {
		filename := v.(string)
		// Grab an exclusive lock so that we're only reading one contact flow into
		// memory at a time.
		// See https://github.com/hashicorp/terraform/issues/9364
		conns.GlobalMutexKV.Lock(mutexKey)
		defer conns.GlobalMutexKV.Unlock(mutexKey)

		file, err := tfio.ReadFileContents(filename)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.Manifest = aws.String(string(file))
	} else if v, ok := d.GetOk("manifest"); ok {
		input.Manifest = aws.String(v.(string))
	}

	output, err := conn.CreateCustomDBEngineVersion(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Custom DB Engine Version: %s", err)
	}

	engine, engineVersion := aws.ToString(output.Engine), aws.ToString(output.EngineVersion)
	d.SetId(customEngineDBVersionResourceID(engine, engineVersion))

	if _, err := waitCustomDBEngineVersionCreated(ctx, conn, engine, engineVersion, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Custom DB Engine Version (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCustomDBEngineVersionRead(ctx, d, meta)...)
}

func resourceCustomDBEngineVersionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	engine, engineVersion, err := customEngineDBVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := findCustomDBEngineVersionByTwoPartKey(ctx, conn, engine, engineVersion)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Custom DB Engine Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Custom DB Engine Version (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.DBEngineVersionArn)
	if out.CreateTime != nil {
		d.Set(names.AttrCreateTime, out.CreateTime.Format(time.RFC3339))
	}
	d.Set("database_installation_files_s3_bucket_name", out.DatabaseInstallationFilesS3BucketName)
	d.Set("database_installation_files_s3_prefix", out.DatabaseInstallationFilesS3Prefix)
	d.Set("db_parameter_group_family", out.DBParameterGroupFamily)
	d.Set(names.AttrDescription, out.DBEngineVersionDescription)
	d.Set(names.AttrEngine, out.Engine)
	d.Set(names.AttrEngineVersion, out.EngineVersion)
	d.Set("image_id", out.Image.ImageId)
	d.Set(names.AttrKMSKeyID, out.KMSKeyId)
	d.Set("major_engine_version", out.MajorEngineVersion)
	d.Set("manifest_computed", out.CustomDBEngineVersionManifest)
	d.Set(names.AttrStatus, out.Status)

	setTagsOut(ctx, out.TagList)

	return diags
}

func resourceCustomDBEngineVersionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	engine, engineVersion, err := customEngineDBVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &rds.ModifyCustomDBEngineVersionInput{
			Engine:        aws.String(engine),
			EngineVersion: aws.String(engineVersion),
		}

		if d.HasChanges(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChanges(names.AttrStatus) {
			input.Status = types.CustomEngineVersionStatus(d.Get(names.AttrStatus).(string))
		}

		_, err := conn.ModifyCustomDBEngineVersion(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Custom DB Engine Version (%s): %s", d.Id(), err)
		}

		if _, err := waitCustomDBEngineVersionUpdated(ctx, conn, engine, engineVersion, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Custom DB Engine Version (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceCustomDBEngineVersionRead(ctx, d, meta)...)
}

func resourceCustomDBEngineVersionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	engine, engineVersion, err := customEngineDBVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting RDS Custom DB Engine Version: %s", d.Id())
	_, err = conn.DeleteCustomDBEngineVersion(ctx, &rds.DeleteCustomDBEngineVersionInput{
		Engine:        aws.String(engine),
		EngineVersion: aws.String(engineVersion),
	})

	if errs.IsA[*types.CustomDBEngineVersionNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Custom DB Engine Version (%s): %s", d.Id(), err)
	}

	if _, err := waitCustomDBEngineVersionDeleted(ctx, conn, engine, engineVersion, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Custom DB Engine Version (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const customEngineDBVersionResourceIDSeparator = ":"

func customEngineDBVersionResourceID(engine, engineVersion string) string {
	parts := []string{engine, engineVersion}
	id := strings.Join(parts, customEngineDBVersionResourceIDSeparator)

	return id
}

func customEngineDBVersionParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, customEngineDBVersionResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected engine%[2]sengineversion", id, customEngineDBVersionResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findCustomDBEngineVersionByTwoPartKey(ctx context.Context, conn *rds.Client, engine, engineVersion string) (*types.DBEngineVersion, error) {
	input := &rds.DescribeDBEngineVersionsInput{
		Engine:        aws.String(engine),
		EngineVersion: aws.String(engineVersion),
		IncludeAll:    aws.Bool(true), // Required to return CEVs that are in `creating` state.
	}

	return findDBEngineVersion(ctx, conn, input, tfslices.PredicateTrue[*types.DBEngineVersion]())
}

func findDBEngineVersion(ctx context.Context, conn *rds.Client, input *rds.DescribeDBEngineVersionsInput, filter tfslices.Predicate[*types.DBEngineVersion]) (*types.DBEngineVersion, error) {
	output, err := findDBEngineVersions(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBEngineVersions(ctx context.Context, conn *rds.Client, input *rds.DescribeDBEngineVersionsInput, filter tfslices.Predicate[*types.DBEngineVersion]) ([]types.DBEngineVersion, error) {
	var output []types.DBEngineVersion

	pages := rds.NewDescribeDBEngineVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.CustomDBEngineVersionNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBEngineVersions {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

const (
	statusAvailable         = "available"
	statusCreating          = "creating"
	statusDeleting          = "deleting"
	statusDeprecated        = "deprecated"
	statusFailed            = "failed"
	statusPendingValidation = "pending-validation" // Custom for SQL Server, ready for validation by an instance
)

func statusDBEngineVersion(ctx context.Context, conn *rds.Client, engine, engineVersion string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findCustomDBEngineVersionByTwoPartKey(ctx, conn, engine, engineVersion)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitCustomDBEngineVersionCreated(ctx context.Context, conn *rds.Client, engine, engineVersion string, timeout time.Duration) (*types.DBEngineVersion, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusCreating},
		Target:                    []string{statusAvailable, statusPendingValidation},
		Refresh:                   statusDBEngineVersion(ctx, conn, engine, engineVersion),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBEngineVersion); ok {
		return output, err
	}

	return nil, err
}

func waitCustomDBEngineVersionUpdated(ctx context.Context, conn *rds.Client, engine, engineVersion string, timeout time.Duration) (*types.DBEngineVersion, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusAvailable},
		Target:  []string{statusAvailable, statusPendingValidation},
		Refresh: statusDBEngineVersion(ctx, conn, engine, engineVersion),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBEngineVersion); ok {
		return output, err
	}

	return nil, err
}

func waitCustomDBEngineVersionDeleted(ctx context.Context, conn *rds.Client, engine, engineVersion string, timeout time.Duration) (*types.DBEngineVersion, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting},
		Target:  []string{},
		Refresh: statusDBEngineVersion(ctx, conn, engine, engineVersion),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBEngineVersion); ok {
		return output, err
	}

	return nil, err
}
