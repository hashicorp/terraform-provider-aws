// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/mitchellh/go-homedir"
)

const cevMutexKey = `aws_rds_custom_engine_version`

// @SDKResource("aws_rds_custom_db_engine_version", name="Custom DB Engine Version")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func ResourceCustomDBEngineVersion() *schema.Resource {
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
				StateFunc: func(v interface{}) string {
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(rds.CustomEngineVersionStatus_Values(), false),
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
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameCustomDBEngineVersion = "Custom DB Engine Version"
)

func resourceCustomDBEngineVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := rds.CreateCustomDBEngineVersionInput{
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

	if v, ok := d.GetOk("source_image_id"); ok {
		input.ImageId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KMSKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("filename"); ok {
		filename := v.(string)
		// Grab an exclusive lock so that we're only reading one contact flow into
		// memory at a time.
		// See https://github.com/hashicorp/terraform/issues/9364
		conns.GlobalMutexKV.Lock(cevMutexKey)
		defer conns.GlobalMutexKV.Unlock(cevMutexKey)
		file, err := resourceCustomDBEngineVersionLoadFileContent(filename)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "unable to load %q: %s", filename, err)
		}
		input.Manifest = aws.String(file)
	} else if v, ok := d.GetOk("manifest"); ok {
		input.Manifest = aws.String(v.(string))
	}

	output, err := conn.CreateCustomDBEngineVersionWithContext(ctx, &input)
	if err != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionCreating, ResNameCustomDBEngineVersion, fmt.Sprintf("%s:%s", aws.StringValue(output.Engine), aws.StringValue(output.EngineVersion)), err)
	}

	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(output.Engine), aws.StringValue(output.EngineVersion)))

	if _, err := waitCustomDBEngineVersionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionWaitingForCreation, ResNameCustomDBEngineVersion, d.Id(), err)
	}

	return append(diags, resourceCustomDBEngineVersionRead(ctx, d, meta)...)
}

func resourceCustomDBEngineVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	out, err := FindCustomDBEngineVersionByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS CustomDBEngineVersion (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionReading, ResNameCustomDBEngineVersion, d.Id(), err)
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

func resourceCustomDBEngineVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	if d.HasChangesExcept(names.AttrDescription, names.AttrStatus) {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionUpdating, ResNameCustomDBEngineVersion, d.Id(), errors.New("only description and status can be updated"))
	}

	update := false
	engine, engineVersion, e := customEngineVersionParseID(d.Id())
	if e != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionUpdating, ResNameCustomDBEngineVersion, d.Id(), e)
	}
	input := &rds.ModifyCustomDBEngineVersionInput{
		Engine:        aws.String(engine),
		EngineVersion: aws.String(engineVersion),
	}

	if d.HasChanges(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
		update = true
	}
	if d.HasChanges(names.AttrStatus) {
		input.Status = aws.String(d.Get(names.AttrStatus).(string))
		update = true
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating RDS CustomDBEngineVersion (%s): %#v", d.Id(), input)
	output, err := conn.ModifyCustomDBEngineVersionWithContext(ctx, input)
	if err != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionUpdating, ResNameCustomDBEngineVersion, d.Id(), err)
	}
	if output == nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionUpdating, ResNameCustomDBEngineVersion, d.Id(), errors.New("empty output"))
	}

	if _, err := waitCustomDBEngineVersionUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionWaitingForUpdate, ResNameCustomDBEngineVersion, d.Id(), err)
	}

	return append(diags, resourceCustomDBEngineVersionRead(ctx, d, meta)...)
}

func resourceCustomDBEngineVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	log.Printf("[INFO] Deleting RDS CustomDBEngineVersion %s", d.Id())

	engine, engineVersion, e := customEngineVersionParseID(d.Id())
	if e != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionUpdating, ResNameCustomDBEngineVersion, d.Id(), e)
	}
	_, err := conn.DeleteCustomDBEngineVersionWithContext(ctx, &rds.DeleteCustomDBEngineVersionInput{
		Engine:        aws.String(engine),
		EngineVersion: aws.String(engineVersion),
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeCustomDBEngineVersionNotFoundFault) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionDeleting, ResNameCustomDBEngineVersion, d.Id(), err)
	}

	if _, err := waitCustomDBEngineVersionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionWaitingForDeletion, ResNameCustomDBEngineVersion, d.Id(), err)
	}

	return diags
}

const (
	statusAvailable         = "available"
	statusCreating          = "creating"
	statusDeleting          = "deleting"
	statusDeprecated        = "deprecated"
	statusFailed            = "failed"
	statusPendingValidation = "pending-validation" // Custom for SQL Server, ready for validation by an instance
)

func waitCustomDBEngineVersionCreated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBEngineVersion, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusCreating},
		Target:                    []string{statusAvailable, statusPendingValidation},
		Refresh:                   statusCustomDBEngineVersion(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rds.DBEngineVersion); ok {
		return out, err
	}

	return nil, err
}

func waitCustomDBEngineVersionUpdated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBEngineVersion, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusAvailable},
		Target:                    []string{statusAvailable, statusPendingValidation},
		Refresh:                   statusCustomDBEngineVersion(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rds.DBEngineVersion); ok {
		return out, err
	}

	return nil, err
}

func waitCustomDBEngineVersionDeleted(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBEngineVersion, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting},
		Target:  []string{},
		Refresh: statusCustomDBEngineVersion(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rds.DBEngineVersion); ok {
		return out, err
	}

	return nil, err
}

func statusCustomDBEngineVersion(ctx context.Context, conn *rds.RDS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindCustomDBEngineVersionByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.StringValue(out.Status), nil
	}
}

func FindCustomDBEngineVersionByID(ctx context.Context, conn *rds.RDS, id string) (*rds.DBEngineVersion, error) {
	engine, engineVersion, e := customEngineVersionParseID(id)
	if e != nil {
		return nil, e
	}
	input := &rds.DescribeDBEngineVersionsInput{
		Engine:        aws.String(engine),
		EngineVersion: aws.String(engineVersion),
		IncludeAll:    aws.Bool(true), // Required to return CEVs that are in `creating` state
	}

	output, err := conn.DescribeDBEngineVersionsWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, rds.ErrCodeCustomDBEngineVersionNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}
	if output == nil || len(output.DBEngineVersions) == 0 {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output.DBEngineVersions[0], nil
}

func customEngineVersionParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected engine:engineversion", id)
	}

	return parts[0], parts[1], nil
}

func resourceCustomDBEngineVersionLoadFileContent(filename string) (string, error) {
	filename, err := homedir.Expand(filename)
	if err != nil {
		return "", err
	}
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(fileContent), nil
}
