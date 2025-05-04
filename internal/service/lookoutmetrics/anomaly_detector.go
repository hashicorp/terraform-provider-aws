// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lookoutmetrics

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lookoutmetrics"
	"github.com/aws/aws-sdk-go-v2/service/lookoutmetrics/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lookoutmetrics_anomaly_detector", name="Anomaly Detector")
// @Tags(identifierAttribute="arn")
func ResourceAnomalyDetector() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAnomalyDetectorCreate,
		ReadWithoutTimeout:   resourceAnomalyDetectorRead,
		// UpdateWithoutTimeout: resourceAnomalyDetectorUpdate,
		// DeleteWithoutTimeout: resourceAnomalyDetectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"frequency": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.Frequency](),
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameAnomalyDetector = "Anomaly Detector"
)

func resourceAnomalyDetectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LookoutMetricsClient(ctx)

	in := &lookoutmetrics.CreateAnomalyDetectorInput{
		AnomalyDetectorName:   aws.String(d.Get("name").(string)),
		AnomalyDetectorConfig: expandConfig(d.Get("config").([]interface{})[0].(map[string]interface{})),

		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		in.AnomalyDetectorDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		in.KmsKeyArn = aws.String(v.(string))
	}

	out, err := conn.CreateAnomalyDetector(ctx, in)
	if err != nil {
		return create.DiagError(names.LookoutMetrics, create.ErrActionCreating, ResNameAnomalyDetector, d.Get("name").(string), err)
	}

	if out == nil || out.AnomalyDetectorArn == nil {
		return create.DiagError(names.LookoutMetrics, create.ErrActionCreating, ResNameAnomalyDetector, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.AnomalyDetectorArn))

	if _, err := waitAnomalyDetectorCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.LookoutMetrics, create.ErrActionWaitingForCreation, ResNameAnomalyDetector, d.Id(), err)
	}

	return resourceAnomalyDetectorRead(ctx, d, meta)
}

func resourceAnomalyDetectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LookoutMetricsClient(ctx)

	out, err := findAnomalyDetectorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] LookoutMetrics AnomalyDetector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.LookoutMetrics, create.ErrActionReading, ResNameAnomalyDetector, d.Id(), err)
	}

	d.Set(names.AttrARN, out.AnomalyDetectorArn)
	d.Set(names.AttrDescription, out.AnomalyDetectorDescription)
	d.Set(names.AttrKMSKeyARN, out.KmsKeyArn)
	d.Set(names.AttrName, out.AnomalyDetectorName)

	if out.AnomalyDetectorConfig != nil {
		if err := d.Set("config", []interface{}{flattenConfig(out.AnomalyDetectorConfig)}); err != nil {
			return create.DiagError(names.LookoutMetrics, create.ErrActionSetting, ResNameAnomalyDetector, d.Id(), err)
		}
	} else {
		d.Set("config", nil)
	}

	// TODO: not sure how to handle tags here
	// setTagsOut(ctx, out.Tags)
	return nil
}

// func resourceAnomalyDetectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	// TIP: ==== RESOURCE UPDATE ====
// 	// Not all resources have Update functions. There are a few reasons:
// 	// a. The AWS API does not support changing a resource
// 	// b. All arguments have ForceNew: true, set
// 	// c. The AWS API uses a create call to modify an existing resource
// 	//
// 	// In the cases of a. and b., the main resource function will not have a
// 	// UpdateWithoutTimeout defined. In the case of c., Update and Create are
// 	// the same.
// 	//
// 	// The rest of the time, there should be an Update function and it should
// 	// do the following things. Make sure there is a good reason if you don't
// 	// do one of these.
// 	//
// 	// 1. Get a client connection to the relevant service
// 	// 2. Populate a modify input structure and check for changes
// 	// 3. Call the AWS modify/update function
// 	// 4. Use a waiter to wait for update to complete
// 	// 5. Call the Read function in the Update return

// 	// TIP: -- 1. Get a client connection to the relevant service
// 	conn := meta.(*conns.AWSClient).LookoutMetricsClient()

// 	// TIP: -- 2. Populate a modify input structure and check for changes
// 	//
// 	// When creating the input structure, only include mandatory fields. Other
// 	// fields are set as needed. You can use a flag, such as update below, to
// 	// determine if a certain portion of arguments have been changed and
// 	// whether to call the AWS update function.
// 	update := false

// 	in := &lookoutmetrics.UpdateAnomalyDetectorInput{
// 		Id: aws.String(d.Id()),
// 	}

// 	if d.HasChanges("an_argument") {
// 		in.AnArgument = aws.String(d.Get("an_argument").(string))
// 		update = true
// 	}

// 	if !update {
// 		// TIP: If update doesn't do anything at all, which is rare, you can
// 		// return nil. Otherwise, return a read call, as below.
// 		return nil
// 	}

// 	// TIP: -- 3. Call the AWS modify/update function
// 	log.Printf("[DEBUG] Updating LookoutMetrics AnomalyDetector (%s): %#v", d.Id(), in)
// 	out, err := conn.UpdateAnomalyDetector(ctx, in)
// 	if err != nil {
// 		return create.DiagError(names.LookoutMetrics, create.ErrActionUpdating, ResNameAnomalyDetector, d.Id(), err)
// 	}

// 	// TIP: -- 4. Use a waiter to wait for update to complete
// 	if _, err := waitAnomalyDetectorUpdated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
// 		return create.DiagError(names.LookoutMetrics, create.ErrActionWaitingForUpdate, ResNameAnomalyDetector, d.Id(), err)
// 	}

// 	// TIP: -- 5. Call the Read function in the Update return
// 	return resourceAnomalyDetectorRead(ctx, d, meta)
// }

// func resourceAnomalyDetectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	// TIP: ==== RESOURCE DELETE ====
// 	// Most resources have Delete functions. There are rare situations
// 	// where you might not need a delete:
// 	// a. The AWS API does not provide a way to delete the resource
// 	// b. The point of your resource is to perform an action (e.g., reboot a
// 	//    server) and deleting serves no purpose.
// 	//
// 	// The Delete function should do the following things. Make sure there
// 	// is a good reason if you don't do one of these.
// 	//
// 	// 1. Get a client connection to the relevant service
// 	// 2. Populate a delete input structure
// 	// 3. Call the AWS delete function
// 	// 4. Use a waiter to wait for delete to complete
// 	// 5. Return nil

// 	// TIP: -- 1. Get a client connection to the relevant service
// 	conn := meta.(*conns.AWSClient).LookoutMetricsClient()

// 	// TIP: -- 2. Populate a delete input structure
// 	log.Printf("[INFO] Deleting LookoutMetrics AnomalyDetector %s", d.Id())

// 	// TIP: -- 3. Call the AWS delete function
// 	_, err := conn.DeleteAnomalyDetector(ctx, &lookoutmetrics.DeleteAnomalyDetectorInput{
// 		Id: aws.String(d.Id()),
// 	})

// 	// TIP: On rare occassions, the API returns a not found error after deleting a
// 	// resource. If that happens, we don't want it to show up as an error.
// 	if err != nil {
// 		var nfe *types.ResourceNotFoundException
// 		if errors.As(err, &nfe) {
// 			return nil
// 		}

// 		return create.DiagError(names.LookoutMetrics, create.ErrActionDeleting, ResNameAnomalyDetector, d.Id(), err)
// 	}

// 	// TIP: -- 4. Use a waiter to wait for delete to complete
// 	if _, err := waitAnomalyDetectorDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
// 		return create.DiagError(names.LookoutMetrics, create.ErrActionWaitingForDeletion, ResNameAnomalyDetector, d.Id(), err)
// 	}

// 	// TIP: -- 5. Return nil
// 	return nil
// }

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., amp.WorkspaceStatusCodeActive).
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.

func waitAnomalyDetectorCreated(ctx context.Context, conn *lookoutmetrics.Client, id string, timeout time.Duration) (*types.AnomalyDetectorSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusAnomalyDetector(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.AnomalyDetectorSummary); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.

// func waitAnomalyDetectorUpdated(ctx context.Context, conn *lookoutmetrics.Client, id string, timeout time.Duration) (*lookoutmetrics.AnomalyDetector, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{statusChangePending},
// 		Target:                    []string{statusUpdated},
// 		Refresh:                   statusAnomalyDetector(ctx, conn, id),
// 		Timeout:                   timeout,
// 		NotFoundChecks:            20,
// 		ContinuousTargetOccurence: 2,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*lookoutmetrics.AnomalyDetector); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.

// func waitAnomalyDetectorDeleted(ctx context.Context, conn *lookoutmetrics.Client, id string, timeout time.Duration) (*lookoutmetrics.AnomalyDetector, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending: []string{statusDeleting, statusNormal},
// 		Target:  []string{},
// 		Refresh: statusAnomalyDetector(ctx, conn, id),
// 		Timeout: timeout,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*lookoutmetrics.AnomalyDetector); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.

func statusAnomalyDetector(ctx context.Context, conn *lookoutmetrics.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findAnomalyDetectorByARN(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findAnomalyDetectorByARN(ctx context.Context, conn *lookoutmetrics.Client, arn string) (*lookoutmetrics.DescribeAnomalyDetectorOutput, error) {

	input := &lookoutmetrics.DescribeAnomalyDetectorInput{
		AnomalyDetectorArn: aws.String(arn),
	}

	output, err := conn.DescribeAnomalyDetector(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

	if status := output.Status; status == types.AnomalyDetectorStatusDeleting {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func flattenConfig(config *types.AnomalyDetectorConfigSummary) map[string]interface{} {
	if config == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["frequency"] = config.AnomalyDetectorFrequency

	return m
}

func expandConfig(config map[string]interface{}) *types.AnomalyDetectorConfig {
	if config == nil {
		return nil
	}

	a := &types.AnomalyDetectorConfig{}

	if v, ok := config["frequency"].(string); ok && v != "" {
		a.AnomalyDetectorFrequency = types.Frequency(v)
	}

	return a
}
