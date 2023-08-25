// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/mitchellh/go-homedir"
)

const cevMutexKey = `aws_rds_custom_engine_version`

// @SDKResource("aws_rds_custom_db_engine_version", name="Custom DB Engine Version")
// @Tags(identifierAttribute="arn")
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_installation_files_s3_bucket_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			"database_installation_files_s3_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"db_parameter_group_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"engine": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(fmt.Sprintf(`^%s.*$`, InstanceEngineCustomPrefix)), fmt.Sprintf("must begin with %s", InstanceEngineCustomPrefix)),
					validation.StringLenBetween(1, 35),
				),
			},
			"engine_version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 60),
			},
			"filename": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"manifest"},
			},
			"image_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"kms_key_id": {
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
				Computed: true,
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
			"manifest_hash": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(), // TIP: Many, but not all, resources have `tags` and `tags_all` attributes.
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
		Engine:        aws.String(d.Get("engine").(string)),
		EngineVersion: aws.String(d.Get("engine_version").(string)),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("database_installation_files_s3_bucket_name"); ok {
		input.DatabaseInstallationFilesS3BucketName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("database_installation_files_s3_prefix"); ok {
		input.DatabaseInstallationFilesS3Prefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_id"); ok {
		input.ImageId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
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
			return diag.Errorf("unable to load %q: %s", filename, err)
		}
		input.Manifest = aws.String(file)
	} else if v, ok := d.GetOk("manifest"); ok {
		input.Manifest = aws.String(v.(string))
	}

	output, err := conn.CreateCustomDBEngineVersionWithContext(ctx, &input)
	if err != nil {
		return append(diags, create.DiagError(names.RDS, create.ErrActionCreating, ResNameCustomDBEngineVersion, fmt.Sprintf("%s:%s", aws.StringValue(output.Engine), aws.StringValue(output.EngineVersion)), err)...)
	}

	if output == nil {
		return append(diags, create.DiagError(names.RDS, create.ErrActionCreating, ResNameCustomDBEngineVersion, fmt.Sprintf("%s:%s", aws.StringValue(output.Engine), aws.StringValue(output.EngineVersion)), errors.New("empty output"))...)
	}

	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(output.Engine), aws.StringValue(output.EngineVersion)))

	if _, err := waitCustomDBEngineVersionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return append(diags, create.DiagError(names.RDS, create.ErrActionWaitingForCreation, ResNameCustomDBEngineVersion, d.Id(), err)...)
	}

	return append(diags, resourceCustomDBEngineVersionRead(ctx, d, meta)...)
}

func resourceCustomDBEngineVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	out, err := findCustomDBEngineVersionByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS CustomDBEngineVersion (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.RDS, create.ErrActionReading, ResNameCustomDBEngineVersion, d.Id(), err)...)
	}

	d.Set("arn", out.DBEngineVersionArn)
	d.Set("create_time", out.CreateTime)
	d.Set("database_installation_files_s3_bucket_name", out.DatabaseInstallationFilesS3BucketName)
	d.Set("database_installation_files_s3_prefix", out.DatabaseInstallationFilesS3Prefix)
	d.Set("db_parameter_group_family", out.DBParameterGroupFamily)
	d.Set("description", out.DBEngineVersionDescription)
	d.Set("engine", out.Engine)
	d.Set("engine_version", out.EngineVersion)
	d.Set("image_id", out.Image.ImageId)
	d.Set("kms_key_id", out.KMSKeyId)
	d.Set("major_engine_version", out.MajorEngineVersion)
	d.Set("manifest", out.CustomDBEngineVersionManifest)
	d.Set("status", out.Status)

	setTagsOut(ctx, out.TagList)

	return diags
}

func resourceCustomDBEngineVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	update := false

	input := &rds.UpdateCustomDBEngineVersionInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChanges("an_argument") {
		input.AnArgument = aws.String(d.Get("an_argument").(string))
		update = true
	}

	if !update {
		// TIP: If update doesn't do anything at all, which is rare, you can
		// return diags. Otherwise, return a read call, as below.
		return diags
	}

	// TIP: -- 3. Call the AWS modify/update function
	log.Printf("[DEBUG] Updating RDS CustomDBEngineVersion (%s): %#v", d.Id(), in)
	out, err := conn.UpdateCustomDBEngineVersion(ctx, put)
	if err != nil {
		return append(diags, create.DiagError(names.RDS, create.ErrActionUpdating, ResNameCustomDBEngineVersion, d.Id(), err)...)
	}

	// TIP: -- 4. Use a waiter to wait for update to complete
	if _, err := waitCustomDBEngineVersionUpdated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return append(diags, create.DiagError(names.RDS, create.ErrActionWaitingForUpdate, ResNameCustomDBEngineVersion, d.Id(), err)...)
	}

	// TIP: -- 5. Call the Read function in the Update return
	return append(diags, resourceCustomDBEngineVersionRead(ctx, d, meta)...)
}

func resourceCustomDBEngineVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// TIP: ==== RESOURCE DELETE ====
	// Most resources have Delete functions. There are rare situations
	// where you might not need a delete:
	// a. The AWS API does not provide a way to delete the resource
	// b. The point of your resource is to perform an action (e.g., reboot a
	//    server) and deleting serves no purpose.
	//
	// The Delete function should do the following things. Make sure there
	// is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Populate a delete input structure
	// 3. Call the AWS delete function
	// 4. Use a waiter to wait for delete to complete
	// 5. Return diags

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	// TIP: -- 2. Populate a delete input structure
	log.Printf("[INFO] Deleting RDS CustomDBEngineVersion %s", d.Id())

	// TIP: -- 3. Call the AWS delete function
	_, err := conn.DeleteCustomDBEngineVersion(ctx, &rds.DeleteCustomDBEngineVersionInput{
		Id: aws.String(d.Id()),
	})

	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}
	if err != nil {
		return append(diags, create.DiagError(names.RDS, create.ErrActionDeleting, ResNameCustomDBEngineVersion, d.Id(), err)...)
	}

	// TIP: -- 4. Use a waiter to wait for delete to complete
	if _, err := waitCustomDBEngineVersionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return append(diags, create.DiagError(names.RDS, create.ErrActionWaitingForDeletion, ResNameCustomDBEngineVersion, d.Id(), err)...)
	}

	// TIP: -- 5. Return diags
	return diags
}

const (
	statusAvailable  = "available"
	statusCreating   = "creating"
	statusDeleting   = "deleting"
	statusDeprecated = "deprecated"
	statusFailed     = "failed"
)

func waitCustomDBEngineVersionCreated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBEngineVersion, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusCreating},
		Target:                    []string{statusAvailable},
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

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.

func waitCustomDBEngineVersionUpdated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBEngineVersion, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
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

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.

func waitCustomDBEngineVersionDeleted(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBEngineVersion, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
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
		out, err := findCustomDBEngineVersionByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.StringValue(out.Status), nil
	}
}

func findCustomDBEngineVersionByID(ctx context.Context, conn *rds.RDS, id string) (*rds.DBEngineVersion, error) {
	engine, engineVersion, err := customEngineVersionParseID(id)
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

// TIP: ==== FLEX ====
// Flatteners and expanders ("flex" functions) help handle complex data
// types. Flatteners take an API data type and return something you can use in
// a d.Set() call. In other words, flatteners translate from AWS -> Terraform.
//
// On the other hand, expanders take a Terraform data structure and return
// something that you can send to the AWS API. In other words, expanders
// translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
func flattenComplexArgument(apiObject *rds.ComplexArgument) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SubFieldOne; v != nil {
		m["sub_field_one"] = aws.ToString(v)
	}

	if v := apiObject.SubFieldTwo; v != nil {
		m["sub_field_two"] = aws.ToString(v)
	}

	return m
}

// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
func flattenComplexArguments(apiObjects []*rds.ComplexArgument) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		l = append(l, flattenComplexArgument(apiObject))
	}

	return l
}

// TIP: Remember, as mentioned above, expanders take a Terraform data structure
// and return something that you can send to the AWS API. In other words,
// expanders translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
func expandComplexArgument(tfMap map[string]interface{}) *rds.ComplexArgument {
	if tfMap == nil {
		return nil
	}

	a := &rds.ComplexArgument{}

	if v, ok := tfMap["sub_field_one"].(string); ok && v != "" {
		a.SubFieldOne = aws.String(v)
	}

	if v, ok := tfMap["sub_field_two"].(string); ok && v != "" {
		a.SubFieldTwo = aws.String(v)
	}

	return a
}

// TIP: Even when you have a list with max length of 1, this plural function
// works brilliantly. However, if the AWS API takes a structure rather than a
// slice of structures, you will not need it.
func expandComplexArguments(tfList []interface{}) []*rds.ComplexArgument {
	// TIP: The AWS API can be picky about whether you send a nil or zero-
	// length for an argument that should be cleared. For example, in some
	// cases, if you send a nil value, the AWS API interprets that as "make no
	// changes" when what you want to say is "remove everything." Sometimes
	// using a zero-length list will cause an error.
	//
	// As a result, here are two options. Usually, option 1, nil, will work as
	// expected, clearing the field. But, test going from something to nothing
	// to make sure it works. If not, try the second option.

	// TIP: Option 1: Returning nil for zero-length list
	if len(tfList) == 0 {
		return nil
	}

	var s []*rds.ComplexArgument

	// TIP: Option 2: Return zero-length list for zero-length list. If option 1 does
	// not work, after testing going from something to nothing (if that is
	// possible), uncomment out the next line and remove option 1.
	//
	// s := make([]*rds.ComplexArgument, 0)

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandComplexArgument(m)

		if a == nil {
			continue
		}

		s = append(s, a)
	}

	return s
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
