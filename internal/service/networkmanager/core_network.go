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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
func ResourceCoreNetwork() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCoreNetworkCreate,
		ReadWithoutTimeout:   resourceCoreNetworkRead,
		UpdateWithoutTimeout: resourceCoreNetworkUpdate,
		DeleteWithoutTimeout: resourceCoreNetworkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

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
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				ConflictsWith: []string{"base_policy_region", "base_policy_regions"},
			},
			"base_policy_region": {
				Deprecated: "Use the base_policy_regions argument instead. " +
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

func resourceCoreNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

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
			regions := []interface{}{meta.(*conns.AWSClient).Region}
			if v, ok := d.GetOk("base_policy_region"); ok {
				regions = []interface{}{v.(string)}
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

	output, err := conn.CreateCoreNetworkWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Core Network: %s", err)
	}

	d.SetId(aws.StringValue(output.CoreNetwork.CoreNetworkId))

	if _, err := waitCoreNetworkCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Core Network (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCoreNetworkRead(ctx, d, meta)...)
}

func resourceCoreNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	coreNetwork, err := FindCoreNetworkByID(ctx, conn, d.Id())

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
		d.Set(names.AttrCreatedAt, aws.TimeValue(coreNetwork.CreatedAt).Format(time.RFC3339))
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

func resourceCoreNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	if d.HasChange(names.AttrDescription) {
		_, err := conn.UpdateCoreNetworkWithContext(ctx, &networkmanager.UpdateCoreNetworkInput{
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
			regions := []interface{}{meta.(*conns.AWSClient).Region}
			if v, ok := d.GetOk("base_policy_region"); ok {
				regions = []interface{}{v.(string)}
			} else if v, ok := d.GetOk("base_policy_regions"); ok && v.(*schema.Set).Len() > 0 {
				regions = v.(*schema.Set).List()
			}

			policyDocumentTarget, err := buildCoreNetworkBasePolicyDocument(regions)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "Formatting Core Network Base Policy: %s", err)
			}

			err = PutAndExecuteCoreNetworkPolicy(ctx, conn, d.Id(), policyDocumentTarget)

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

func resourceCoreNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	log.Printf("[DEBUG] Deleting Network Manager Core Network: %s", d.Id())
	_, err := conn.DeleteCoreNetworkWithContext(ctx, &networkmanager.DeleteCoreNetworkInput{
		CoreNetworkId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
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

func FindCoreNetworkByID(ctx context.Context, conn *networkmanager.NetworkManager, id string) (*networkmanager.CoreNetwork, error) {
	input := &networkmanager.GetCoreNetworkInput{
		CoreNetworkId: aws.String(id),
	}

	output, err := conn.GetCoreNetworkWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
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

func FindCoreNetworkPolicyByTwoPartKey(ctx context.Context, conn *networkmanager.NetworkManager, coreNetworkID string, policyVersionID int64) (*networkmanager.CoreNetworkPolicy, error) {
	input := &networkmanager.GetCoreNetworkPolicyInput{
		CoreNetworkId: aws.String(coreNetworkID),
	}
	if policyVersionID >= minimumValidPolicyVersionID {
		input.PolicyVersionId = aws.Int64(policyVersionID)
	}

	output, err := conn.GetCoreNetworkPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
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

func statusCoreNetworkState(ctx context.Context, conn *networkmanager.NetworkManager, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCoreNetworkByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitCoreNetworkCreated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.CoreNetwork, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.CoreNetworkStateCreating, coreNetworkStatePending},
		Target:  []string{networkmanager.CoreNetworkStateAvailable},
		Timeout: timeout,
		Refresh: statusCoreNetworkState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.CoreNetwork); ok {
		return output, err
	}

	return nil, err
}

func waitCoreNetworkUpdated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.CoreNetwork, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.CoreNetworkStateUpdating},
		Target:  []string{networkmanager.CoreNetworkStateAvailable},
		Timeout: timeout,
		Refresh: statusCoreNetworkState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.CoreNetwork); ok {
		return output, err
	}

	return nil, err
}

func waitCoreNetworkDeleted(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.CoreNetwork, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{networkmanager.CoreNetworkStateDeleting},
		Target:     []string{},
		Timeout:    timeout,
		Delay:      5 * time.Minute,
		MinTimeout: 10 * time.Second,
		Refresh:    statusCoreNetworkState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.CoreNetwork); ok {
		return output, err
	}

	return nil, err
}

func flattenCoreNetworkEdge(apiObject *networkmanager.CoreNetworkEdge) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Asn; v != nil {
		tfMap["asn"] = aws.Int64Value(v)
	}

	if v := apiObject.EdgeLocation; v != nil {
		tfMap["edge_location"] = aws.StringValue(v)
	}

	if v := apiObject.InsideCidrBlocks; v != nil {
		tfMap["inside_cidr_blocks"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenCoreNetworkEdges(apiObjects []*networkmanager.CoreNetworkEdge) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCoreNetworkEdge(apiObject))
	}

	return tfList
}

func flattenCoreNetworkSegment(apiObject *networkmanager.CoreNetworkSegment) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EdgeLocations; v != nil {
		tfMap["edge_locations"] = aws.StringValueSlice(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.StringValue(v)
	}

	if v := apiObject.SharedSegments; v != nil {
		tfMap["shared_segments"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenCoreNetworkSegments(apiObjects []*networkmanager.CoreNetworkSegment) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCoreNetworkSegment(apiObject))
	}

	return tfList
}

func PutAndExecuteCoreNetworkPolicy(ctx context.Context, conn *networkmanager.NetworkManager, coreNetworkId, policyDocument string) error {
	v, err := protocol.DecodeJSONValue(policyDocument, protocol.NoEscape)

	if err != nil {
		return fmt.Errorf("decoding Network Manager Core Network (%s) policy document: %s", coreNetworkId, err)
	}

	output, err := conn.PutCoreNetworkPolicyWithContext(ctx, &networkmanager.PutCoreNetworkPolicyInput{
		ClientToken:    aws.String(id.UniqueId()),
		CoreNetworkId:  aws.String(coreNetworkId),
		PolicyDocument: v,
	})

	if err != nil {
		return fmt.Errorf("putting Network Manager Core Network (%s) policy: %s", coreNetworkId, err)
	}

	policyVersionID := aws.Int64Value(output.CoreNetworkPolicy.PolicyVersionId)

	if _, err := waitCoreNetworkPolicyCreated(ctx, conn, coreNetworkId, policyVersionID, waitCoreNetworkPolicyCreatedTimeInMinutes*time.Minute); err != nil {
		return fmt.Errorf("waiting for Network Manager Core Network Policy from Core Network (%s) create: %s", coreNetworkId, err)
	}

	_, err = conn.ExecuteCoreNetworkChangeSetWithContext(ctx, &networkmanager.ExecuteCoreNetworkChangeSetInput{
		CoreNetworkId:   aws.String(coreNetworkId),
		PolicyVersionId: aws.Int64(policyVersionID),
	})
	if err != nil {
		return fmt.Errorf("executing Network Manager Core Network (%s) change set (%d): %s", coreNetworkId, policyVersionID, err)
	}

	return nil
}

func statusCoreNetworkPolicyState(ctx context.Context, conn *networkmanager.NetworkManager, coreNetworkId string, policyVersionId int64) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCoreNetworkPolicyByTwoPartKey(ctx, conn, coreNetworkId, policyVersionId)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ChangeSetState), nil
	}
}

func waitCoreNetworkPolicyCreated(ctx context.Context, conn *networkmanager.NetworkManager, coreNetworkId string, policyVersionId int64, timeout time.Duration) (*networkmanager.CoreNetworkPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.ChangeSetStatePendingGeneration},
		Target:  []string{networkmanager.ChangeSetStateReadyToExecute},
		Timeout: timeout,
		Refresh: statusCoreNetworkPolicyState(ctx, conn, coreNetworkId, policyVersionId),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.CoreNetworkPolicy); ok {
		return output, err
	}

	if output, ok := outputRaw.(*networkmanager.CoreNetworkPolicy); ok {
		if state, v := aws.StringValue(output.ChangeSetState), output.PolicyErrors; state == networkmanager.ChangeSetStateFailedGeneration && len(v) > 0 {
			var errs []error

			for _, err := range v {
				errs = append(errs, fmt.Errorf("%s: %s", aws.StringValue(err.ErrorCode), aws.StringValue(err.Message)))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return output, err
	}

	return nil, err
}

// buildCoreNetworkBasePolicyDocument returns a base policy document
func buildCoreNetworkBasePolicyDocument(regions []interface{}) (string, error) {
	edgeLocations := make([]*coreNetworkPolicyCoreNetworkEdgeLocation, len(regions))
	for i, location := range regions {
		edgeLocations[i] = &coreNetworkPolicyCoreNetworkEdgeLocation{Location: location.(string)}
	}

	basePolicy := &coreNetworkPolicyDocument{
		Version: "2021.12",
		CoreNetworkConfiguration: &coreNetworkPolicyCoreNetworkConfiguration{
			AsnRanges:     coreNetworkPolicyExpandStringList([]interface{}{"64512-65534"}),
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
