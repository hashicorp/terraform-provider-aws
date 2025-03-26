// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
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

const (
	// CoreNetwork is in PENDING state before AVAILABLE. No value for PENDING at the moment.
	coreNetworkStatePending = "PENDING"
	// Minimum valid policy version id is 1
	minimumValidPolicyVersionID = 1
	// Using the following in the FindCoreNetworkPolicyByID function will default to get the latest policy version
	latestPolicyVersionID = -1
	// Wait time value for core network policy - the default update for the core network policy of 30 minutes is excessive
	waitCoreNetworkPolicyCreatedTimeInMinutes = 5
)

// @SDKResource("aws_networkmanager_core_network", name="Core Network")
// @Tags(identifierAttribute="arn")
func resourceCoreNetwork() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCoreNetworkCreate,
		ReadWithoutTimeout:   resourceCoreNetworkRead,
		UpdateWithoutTimeout: resourceCoreNetworkUpdate,
		DeleteWithoutTimeout: resourceCoreNetworkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_policy_document": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 10000000),
					validation.StringIsJSON,
				),
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				ConflictsWith: []string{"base_policy_region", "base_policy_regions"},
			},
			"base_policy_region": {
				Deprecated: "base_policy_region is deprecated. Use base_policy_regions instead. " +
					"This argument will be removed in the next major version of the provider.",
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  verify.ValidRegionName,
				ConflictsWith: []string{"base_policy_document", "base_policy_regions"},
			},
			"base_policy_regions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidRegionName,
				},
				ConflictsWith: []string{"base_policy_document", "base_policy_region"},
			},
			"create_base_policy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"edges": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"asn": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"edge_location": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"inside_cidr_blocks": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"global_network_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
			"segments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"edge_locations": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"shared_segments": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCoreNetworkCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	input := &networkmanager.CreateCoreNetworkInput{
		ClientToken:     aws.String(id.UniqueId()),
		GlobalNetworkId: aws.String(globalNetworkID),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	// check if the user wants to create a base policy document
	// this creates the core network with a starting policy document set to LIVE
	// this is required for the first terraform apply if there attachments to the core network
	if _, ok := d.GetOk("create_base_policy"); ok {
		// if user supplies a full base_policy_document for maximum flexibility, use it. Otherwise, use regions list
		// var policyDocumentTarget string
		if v, ok := d.GetOk("base_policy_document"); ok {
			input.PolicyDocument = aws.String(v.(string))
		} else {
			// if user supplies a region or multiple regions use it in the base policy, otherwise use current region
			regions := []any{meta.(*conns.AWSClient).Region(ctx)}
			if v, ok := d.GetOk("base_policy_region"); ok {
				regions = []any{v.(string)}
			} else if v, ok := d.GetOk("base_policy_regions"); ok && v.(*schema.Set).Len() > 0 {
				regions = v.(*schema.Set).List()
			}

			policyDocumentTarget, err := buildCoreNetworkBasePolicyDocument(regions)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "Formatting Core Network Base Policy: %s", err)
			}
			input.PolicyDocument = aws.String(policyDocumentTarget)
		}
	}

	output, err := conn.CreateCoreNetwork(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Core Network: %s", err)
	}

	d.SetId(aws.ToString(output.CoreNetwork.CoreNetworkId))

	if _, err := waitCoreNetworkCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Core Network (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCoreNetworkRead(ctx, d, meta)...)
}

func resourceCoreNetworkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	coreNetwork, err := findCoreNetworkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Core Network %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Core Network (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, coreNetwork.CoreNetworkArn)
	if coreNetwork.CreatedAt != nil {
		d.Set(names.AttrCreatedAt, aws.ToTime(coreNetwork.CreatedAt).Format(time.RFC3339))
	} else {
		d.Set(names.AttrCreatedAt, nil)
	}
	d.Set(names.AttrDescription, coreNetwork.Description)
	if err := d.Set("edges", flattenCoreNetworkEdges(coreNetwork.Edges)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting edges: %s", err)
	}
	d.Set("global_network_id", coreNetwork.GlobalNetworkId)
	if err := d.Set("segments", flattenCoreNetworkSegments(coreNetwork.Segments)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting segments: %s", err)
	}
	d.Set(names.AttrState, coreNetwork.State)

	setTagsOut(ctx, coreNetwork.Tags)

	return diags
}

func resourceCoreNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	if d.HasChange(names.AttrDescription) {
		_, err := conn.UpdateCoreNetwork(ctx, &networkmanager.UpdateCoreNetworkInput{
			CoreNetworkId: aws.String(d.Id()),
			Description:   aws.String(d.Get(names.AttrDescription).(string)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Network Manager Core Network (%s): %s", d.Id(), err)
		}

		if _, err := waitCoreNetworkUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Core Network (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("create_base_policy") {
		if _, ok := d.GetOk("create_base_policy"); ok {
			// if user supplies a region or multiple regions use it in the base policy, otherwise use current region
			regions := []any{meta.(*conns.AWSClient).Region(ctx)}
			if v, ok := d.GetOk("base_policy_region"); ok {
				regions = []any{v.(string)}
			} else if v, ok := d.GetOk("base_policy_regions"); ok && v.(*schema.Set).Len() > 0 {
				regions = v.(*schema.Set).List()
			}

			policyDocumentTarget, err := buildCoreNetworkBasePolicyDocument(regions)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "Formatting Core Network Base Policy: %s", err)
			}

			err = putAndExecuteCoreNetworkPolicy(ctx, conn, d.Id(), policyDocumentTarget)

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			if _, err := waitCoreNetworkUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Core Network (%s) update: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceCoreNetworkRead(ctx, d, meta)...)
}

func resourceCoreNetworkDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	log.Printf("[DEBUG] Deleting Network Manager Core Network: %s", d.Id())
	_, err := conn.DeleteCoreNetwork(ctx, &networkmanager.DeleteCoreNetworkInput{
		CoreNetworkId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager Core Network (%s): %s", d.Id(), err)
	}

	if _, err := waitCoreNetworkDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Core Network (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findCoreNetworkByID(ctx context.Context, conn *networkmanager.Client, id string) (*awstypes.CoreNetwork, error) {
	input := &networkmanager.GetCoreNetworkInput{
		CoreNetworkId: aws.String(id),
	}

	output, err := conn.GetCoreNetwork(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CoreNetwork == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.CoreNetwork, nil
}

func findCoreNetworkPolicyByTwoPartKey(ctx context.Context, conn *networkmanager.Client, coreNetworkID string, policyVersionID *int32) (*awstypes.CoreNetworkPolicy, error) {
	input := &networkmanager.GetCoreNetworkPolicyInput{
		CoreNetworkId: aws.String(coreNetworkID),
	}
	if aws.ToInt32(policyVersionID) >= minimumValidPolicyVersionID {
		input.PolicyVersionId = policyVersionID
	}

	output, err := conn.GetCoreNetworkPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CoreNetworkPolicy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.CoreNetworkPolicy, nil
}

func statusCoreNetworkState(ctx context.Context, conn *networkmanager.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findCoreNetworkByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitCoreNetworkCreated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.CoreNetwork, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CoreNetworkStateCreating, coreNetworkStatePending),
		Target:  enum.Slice(awstypes.CoreNetworkStateAvailable),
		Timeout: timeout,
		Refresh: statusCoreNetworkState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CoreNetwork); ok {
		return output, err
	}

	return nil, err
}

func waitCoreNetworkUpdated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.CoreNetwork, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CoreNetworkStateUpdating),
		Target:  enum.Slice(awstypes.CoreNetworkStateAvailable),
		Timeout: timeout,
		Refresh: statusCoreNetworkState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CoreNetwork); ok {
		return output, err
	}

	return nil, err
}

func waitCoreNetworkDeleted(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.CoreNetwork, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.CoreNetworkStateDeleting),
		Target:     []string{},
		Timeout:    timeout,
		Delay:      5 * time.Minute,
		MinTimeout: 10 * time.Second,
		Refresh:    statusCoreNetworkState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CoreNetwork); ok {
		return output, err
	}

	return nil, err
}

func flattenCoreNetworkEdge(apiObject awstypes.CoreNetworkEdge) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Asn; v != nil {
		tfMap["asn"] = aws.ToInt64(v)
	}

	if v := apiObject.EdgeLocation; v != nil {
		tfMap["edge_location"] = aws.ToString(v)
	}

	if v := apiObject.InsideCidrBlocks; v != nil {
		tfMap["inside_cidr_blocks"] = v
	}

	return tfMap
}

func flattenCoreNetworkEdges(apiObjects []awstypes.CoreNetworkEdge) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCoreNetworkEdge(apiObject))
	}

	return tfList
}

func flattenCoreNetworkSegment(apiObject awstypes.CoreNetworkSegment) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.EdgeLocations; v != nil {
		tfMap["edge_locations"] = v
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.SharedSegments; v != nil {
		tfMap["shared_segments"] = v
	}

	return tfMap
}

func flattenCoreNetworkSegments(apiObjects []awstypes.CoreNetworkSegment) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCoreNetworkSegment(apiObject))
	}

	return tfList
}

func putAndExecuteCoreNetworkPolicy(ctx context.Context, conn *networkmanager.Client, coreNetworkId, policyDocument string) error {
	document, err := structure.NormalizeJsonString(policyDocument)

	if err != nil {
		return fmt.Errorf("decoding Network Manager Core Network (%s) policy document: %s", coreNetworkId, err)
	}

	output, err := conn.PutCoreNetworkPolicy(ctx, &networkmanager.PutCoreNetworkPolicyInput{
		ClientToken:    aws.String(id.UniqueId()),
		CoreNetworkId:  aws.String(coreNetworkId),
		PolicyDocument: aws.String(document),
	})

	if err != nil {
		return fmt.Errorf("putting Network Manager Core Network (%s) policy: %s", coreNetworkId, err)
	}

	policyVersionID := output.CoreNetworkPolicy.PolicyVersionId

	if _, err := waitCoreNetworkPolicyCreated(ctx, conn, coreNetworkId, policyVersionID, waitCoreNetworkPolicyCreatedTimeInMinutes*time.Minute); err != nil {
		return fmt.Errorf("waiting for Network Manager Core Network Policy from Core Network (%s) create: %s", coreNetworkId, err)
	}

	_, err = conn.ExecuteCoreNetworkChangeSet(ctx, &networkmanager.ExecuteCoreNetworkChangeSetInput{
		CoreNetworkId:   aws.String(coreNetworkId),
		PolicyVersionId: policyVersionID,
	})
	if err != nil {
		return fmt.Errorf("executing Network Manager Core Network (%s) change set (%d): %s", coreNetworkId, policyVersionID, err)
	}

	return nil
}

func statusCoreNetworkPolicyState(ctx context.Context, conn *networkmanager.Client, coreNetworkId string, policyVersionId *int32) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findCoreNetworkPolicyByTwoPartKey(ctx, conn, coreNetworkId, policyVersionId)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ChangeSetState), nil
	}
}

func waitCoreNetworkPolicyCreated(ctx context.Context, conn *networkmanager.Client, coreNetworkId string, policyVersionId *int32, timeout time.Duration) (*awstypes.CoreNetworkPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ChangeSetStatePendingGeneration),
		Target:  enum.Slice(awstypes.ChangeSetStateReadyToExecute),
		Timeout: timeout,
		Refresh: statusCoreNetworkPolicyState(ctx, conn, coreNetworkId, policyVersionId),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CoreNetworkPolicy); ok {
		return output, err
	}

	if output, ok := outputRaw.(*awstypes.CoreNetworkPolicy); ok {
		if state, v := output.ChangeSetState, output.PolicyErrors; state == awstypes.ChangeSetStateFailedGeneration && len(v) > 0 {
			var errs []error

			for _, err := range v {
				errs = append(errs, fmt.Errorf("%s: %s", aws.ToString(err.ErrorCode), aws.ToString(err.Message)))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return output, err
	}

	return nil, err
}

// buildCoreNetworkBasePolicyDocument returns a base policy document
func buildCoreNetworkBasePolicyDocument(regions []any) (string, error) {
	edgeLocations := make([]*coreNetworkPolicyCoreNetworkEdgeLocation, len(regions))
	for i, location := range regions {
		edgeLocations[i] = &coreNetworkPolicyCoreNetworkEdgeLocation{Location: location.(string)}
	}

	basePolicy := &coreNetworkPolicyDocument{
		Version: "2021.12",
		CoreNetworkConfiguration: &coreNetworkPolicyCoreNetworkConfiguration{
			AsnRanges:     coreNetworkPolicyExpandStringList([]any{"64512-65534"}),
			EdgeLocations: edgeLocations,
		},
		Segments: []*coreNetworkPolicySegment{
			{
				Name:        "segment",
				Description: "base-policy",
			},
		},
	}

	b, err := json.MarshalIndent(basePolicy, "", "  ")
	if err != nil {
		// should never happen if the above code is correct
		return "", fmt.Errorf("building base policy document: %s", err)
	}

	return string(b), nil
}
