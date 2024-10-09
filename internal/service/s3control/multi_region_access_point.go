// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3control_multi_region_access_point")
func resourceMultiRegionAccessPoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMultiRegionAccessPointCreate,
		ReadWithoutTimeout:   resourceMultiRegionAccessPointRead,
		DeleteWithoutTimeout: resourceMultiRegionAccessPointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrAlias: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"details": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateS3MultiRegionAccessPointName,
						},
						"public_access_block": {
							Type:             schema.TypeList,
							Optional:         true,
							ForceNew:         true,
							MinItems:         0,
							MaxItems:         1,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"block_public_acls": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									"block_public_policy": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									"ignore_public_acls": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									"restrict_public_buckets": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
								},
							},
						},
						names.AttrRegion: {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							MinItems: 1,
							MaxItems: 20,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrBucket: {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(3, 255),
									},
									"bucket_account_id": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidAccountID,
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMultiRegionAccessPointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
	}
	input := &s3control.CreateMultiRegionAccessPointInput{
		AccountId: aws.String(accountID),
	}

	if v, ok := d.GetOk("details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Details = expandCreateMultiRegionAccessPointInput_(v.([]interface{})[0].(map[string]interface{}))
	}

	id := MultiRegionAccessPointCreateResourceID(accountID, aws.ToString(input.Details.Name))

	output, err := conn.CreateMultiRegionAccessPoint(ctx, input, func(o *s3control.Options) {
		// All Multi-Region Access Point actions are routed to the US West (Oregon) Region.
		o.Region = names.USWest2RegionID
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Multi-Region Access Point (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitMultiRegionAccessPointRequestSucceeded(ctx, conn, accountID, aws.ToString(output.RequestTokenARN), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Multi-Region Access Point (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceMultiRegionAccessPointRead(ctx, d, meta)...)
}

func resourceMultiRegionAccessPointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := MultiRegionAccessPointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	accessPoint, err := findMultiRegionAccessPointByTwoPartKey(ctx, conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Multi-Region Access Point (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Multi-Region Access Point (%s): %s", d.Id(), err)
	}

	alias := aws.ToString(accessPoint.Alias)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "s3",
		AccountID: accountID,
		Resource:  fmt.Sprintf("accesspoint/%s", alias),
	}.String()
	d.Set(names.AttrAccountID, accountID)
	d.Set(names.AttrAlias, alias)
	d.Set(names.AttrARN, arn)
	if err := d.Set("details", []interface{}{flattenMultiRegionAccessPointReport(accessPoint)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting details: %s", err)
	}
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide//MultiRegionAccessPointRequests.html#MultiRegionAccessPointHostnames.
	d.Set(names.AttrDomainName, meta.(*conns.AWSClient).PartitionHostname(ctx, alias+".accesspoint.s3-global"))
	d.Set(names.AttrStatus, accessPoint.Status)

	return diags
}

func resourceMultiRegionAccessPointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := MultiRegionAccessPointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3control.DeleteMultiRegionAccessPointInput{
		AccountId: aws.String(accountID),
		Details: &types.DeleteMultiRegionAccessPointInput{
			Name: aws.String(name),
		},
	}

	log.Printf("[DEBUG] Deleting S3 Multi-Region Access Point: %s", d.Id())
	output, err := conn.DeleteMultiRegionAccessPoint(ctx, input, func(o *s3control.Options) {
		// All Multi-Region Access Point actions are routed to the US West (Oregon) Region.
		o.Region = names.USWest2RegionID
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchMultiRegionAccessPoint) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Multi-Region Access Point (%s): %s", d.Id(), err)
	}

	if _, err := waitMultiRegionAccessPointRequestSucceeded(ctx, conn, accountID, aws.ToString(output.RequestTokenARN), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Multi-Region Access Point (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findMultiRegionAccessPointByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, name string) (*types.MultiRegionAccessPointReport, error) {
	input := &s3control.GetMultiRegionAccessPointInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetMultiRegionAccessPoint(ctx, input, func(o *s3control.Options) {
		// All Multi-Region Access Point actions are routed to the US West (Oregon) Region.
		o.Region = names.USWest2RegionID
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchMultiRegionAccessPoint) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AccessPoint == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AccessPoint, nil
}

func findMultiRegionAccessPointOperationByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, requestTokenARN string) (*types.AsyncOperation, error) {
	input := &s3control.DescribeMultiRegionAccessPointOperationInput{
		AccountId:       aws.String(accountID),
		RequestTokenARN: aws.String(requestTokenARN),
	}

	output, err := conn.DescribeMultiRegionAccessPointOperation(ctx, input, func(o *s3control.Options) {
		// All Multi-Region Access Point actions are routed to the US West (Oregon) Region.
		o.Region = names.USWest2RegionID
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAsyncRequest) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AsyncOperation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AsyncOperation, nil
}

func statusMultiRegionAccessPointRequest(ctx context.Context, conn *s3control.Client, accountID, requestTokenARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findMultiRegionAccessPointOperationByTwoPartKey(ctx, conn, accountID, requestTokenARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.RequestStatus), nil
	}
}

func waitMultiRegionAccessPointRequestSucceeded(ctx context.Context, conn *s3control.Client, accountID, requestTokenARN string, timeout time.Duration) (*types.AsyncOperation, error) { //nolint:unparam
	const (
		// AsyncOperation.RequestStatus values.
		asyncOperationRequestStatusFailed    = "FAILED"
		asyncOperationRequestStatusSucceeded = "SUCCEEDED"
	)
	stateConf := &retry.StateChangeConf{
		Target:     []string{asyncOperationRequestStatusSucceeded},
		Timeout:    timeout,
		Refresh:    statusMultiRegionAccessPointRequest(ctx, conn, accountID, requestTokenARN),
		MinTimeout: 5 * time.Second,
		Delay:      15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AsyncOperation); ok {
		if status, responseDetails := aws.ToString(output.RequestStatus), output.ResponseDetails; status == asyncOperationRequestStatusFailed && responseDetails != nil && responseDetails.ErrorDetails != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(responseDetails.ErrorDetails.Code), aws.ToString(responseDetails.ErrorDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

const multiRegionAccessPointResourceIDSeparator = ":"

func MultiRegionAccessPointCreateResourceID(accountID, accessPointName string) string {
	parts := []string{accountID, accessPointName}
	id := strings.Join(parts, multiRegionAccessPointResourceIDSeparator)

	return id
}

func MultiRegionAccessPointParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, multiRegionAccessPointResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected account-id%[2]saccess-point-name", id, multiRegionAccessPointResourceIDSeparator)
}

func expandCreateMultiRegionAccessPointInput_(tfMap map[string]interface{}) *types.CreateMultiRegionAccessPointInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CreateMultiRegionAccessPointInput{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["public_access_block"].([]interface{}); ok && len(v) > 0 {
		apiObject.PublicAccessBlock = expandPublicAccessBlockConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrRegion].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Regions = expandRegions(v.List())
	}

	return apiObject
}

func expandPublicAccessBlockConfiguration(tfMap map[string]interface{}) *types.PublicAccessBlockConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PublicAccessBlockConfiguration{}

	if v, ok := tfMap["block_public_acls"].(bool); ok {
		apiObject.BlockPublicAcls = aws.Bool(v)
	}

	if v, ok := tfMap["block_public_policy"].(bool); ok {
		apiObject.BlockPublicPolicy = aws.Bool(v)
	}

	if v, ok := tfMap["ignore_public_acls"].(bool); ok {
		apiObject.IgnorePublicAcls = aws.Bool(v)
	}

	if v, ok := tfMap["restrict_public_buckets"].(bool); ok {
		apiObject.RestrictPublicBuckets = aws.Bool(v)
	}

	return apiObject
}

func expandRegion(tfMap map[string]interface{}) *types.Region {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Region{}

	if v, ok := tfMap[names.AttrBucket].(string); ok && v != "" {
		apiObject.Bucket = aws.String(v)
	}

	if v, ok := tfMap["bucket_account_id"].(string); ok && v != "" {
		apiObject.BucketAccountId = aws.String(v)
	}

	return apiObject
}

func expandRegions(tfList []interface{}) []types.Region {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.Region

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandRegion(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenMultiRegionAccessPointReport(apiObject *types.MultiRegionAccessPointReport) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.PublicAccessBlock; v != nil {
		tfMap["public_access_block"] = []interface{}{flattenPublicAccessBlockConfiguration(v)}
	}

	if v := apiObject.Regions; v != nil {
		tfMap[names.AttrRegion] = flattenRegionReports(v)
	}

	return tfMap
}

func flattenPublicAccessBlockConfiguration(apiObject *types.PublicAccessBlockConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BlockPublicAcls; v != nil {
		tfMap["block_public_acls"] = aws.ToBool(v)
	}

	if v := apiObject.BlockPublicPolicy; v != nil {
		tfMap["block_public_policy"] = aws.ToBool(v)
	}

	if v := apiObject.IgnorePublicAcls; v != nil {
		tfMap["ignore_public_acls"] = aws.ToBool(v)
	}

	if v := apiObject.RestrictPublicBuckets; v != nil {
		tfMap["restrict_public_buckets"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenRegionReport(apiObject types.RegionReport) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Bucket; v != nil {
		tfMap[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.BucketAccountId; v != nil {
		tfMap["bucket_account_id"] = aws.ToString(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	return tfMap
}

func flattenRegionReports(apiObjects []types.RegionReport) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenRegionReport(apiObject))
	}

	return tfList
}
