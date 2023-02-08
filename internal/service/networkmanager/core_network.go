package networkmanager

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// CoreNetwork is in PENDING state before AVAILABLE. No value for PENDING at the moment.
	coreNetworkStatePending = "PENDING"
)

// This resource is explicitly NOT exported from the provider until design is finalized.
// Its Delete handler is used by sweepers.
func ResourceCoreNetwork() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCoreNetworkCreate,
		ReadWithoutTimeout:   resourceCoreNetworkRead,
		UpdateWithoutTimeout: resourceCoreNetworkUpdate,
		DeleteWithoutTimeout: resourceCoreNetworkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			resourceCoreNetworkCustomizeDiff,
			verify.SetTagsDiff,
		),

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
			"base_policy_region": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidRegionName,
			},
			"create_base_policy": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"policy_document"},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
			"policy_document": {
				Deprecated: "Use the aws_networkmanager_core_network_policy_attachment resource instead. " +
					"This attribute will be removed in the next major version of the provider.",
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 10000000),
					validation.StringIsJSON,
				),
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				ConflictsWith: []string{"create_base_policy"},
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
						"name": {
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
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCoreNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	globalNetworkID := d.Get("global_network_id").(string)

	input := &networkmanager.CreateCoreNetworkInput{
		ClientToken:     aws.String(resource.UniqueId()),
		GlobalNetworkId: aws.String(globalNetworkID),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy_document"); ok {
		input.PolicyDocument = aws.String(v.(string))
	}

	// check if the user wants to create a base policy document
	// this creates the core network with a starting policy document set to LIVE
	// this is required for the first terraform apply if there attachments to the core network
	// and the core network is created without the policy_document argument set
	if _, ok := d.GetOk("create_base_policy"); ok {
		// if user supplies a region use it in the base policy, otherwise use current region
		region := meta.(*conns.AWSClient).Region
		if v, ok := d.GetOk("base_policy_region"); ok {
			region = v.(string)
		}

		policyDocumentTarget := buildCoreNetworkBasePolicyDocument(region)
		input.PolicyDocument = aws.String(policyDocumentTarget)
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateCoreNetworkWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Core Network: %s", err)
	}

	d.SetId(aws.StringValue(output.CoreNetwork.CoreNetworkId))

	if _, err := waitCoreNetworkCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Network Manager Core Network (%s) create: %s", d.Id(), err)
	}

	return resourceCoreNetworkRead(ctx, d, meta)
}

func resourceCoreNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	coreNetwork, err := FindCoreNetworkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Core Network %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Network Manager Core Network (%s): %s", d.Id(), err)
	}

	d.Set("arn", coreNetwork.CoreNetworkArn)
	if coreNetwork.CreatedAt != nil {
		d.Set("created_at", aws.TimeValue(coreNetwork.CreatedAt).Format(time.RFC3339))
	} else {
		d.Set("created_at", nil)
	}
	d.Set("description", coreNetwork.Description)
	if err := d.Set("edges", flattenCoreNetworkEdges(coreNetwork.Edges)); err != nil {
		return diag.Errorf("setting edges: %s", err)
	}
	d.Set("global_network_id", coreNetwork.GlobalNetworkId)
	if err := d.Set("segments", flattenCoreNetworkSegments(coreNetwork.Segments)); err != nil {
		return diag.Errorf("setting segments: %s", err)
	}
	d.Set("state", coreNetwork.State)

	// getting the policy document uses a different API call
	// policy document is also optional
	coreNetworkPolicy, err := FindCoreNetworkPolicyByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		d.Set("policy_document", nil)
	} else if err != nil {
		return diag.Errorf("reading Network Manager Core Network (%s) policy: %s", d.Id(), err)
	} else {
		encodedPolicyDocument, err := protocol.EncodeJSONValue(coreNetworkPolicy.PolicyDocument, protocol.NoEscape)

		if err != nil {
			return diag.Errorf("encoding Network Manager Core Network (%s) policy document: %s", d.Id(), err)
		}

		d.Set("policy_document", encodedPolicyDocument)
	}

	tags := KeyValueTags(coreNetwork.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceCoreNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	if d.HasChange("description") {
		_, err := conn.UpdateCoreNetworkWithContext(ctx, &networkmanager.UpdateCoreNetworkInput{
			CoreNetworkId: aws.String(d.Id()),
			Description:   aws.String(d.Get("description").(string)),
		})

		if err != nil {
			return diag.Errorf("updating Network Manager Core Network (%s): %s", d.Id(), err)
		}

		if _, err := waitCoreNetworkUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for Network Manager Core Network (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("policy_document") {
		err := PutAndExecuteCoreNetworkPolicy(ctx, conn, d.Id(), d.Get("policy_document").(string))

		if err != nil {
			return diag.FromErr(err)
		}

		if _, err := waitCoreNetworkUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for Network Manager Core Network (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("create_base_policy") {
		if _, ok := d.GetOk("create_base_policy"); ok {
			// if user supplies a region use it in the base policy, otherwise use current region
			region := meta.(*conns.AWSClient).Region
			if v, ok := d.GetOk("base_policy_region"); ok {
				region = v.(string)
			}

			policyDocumentTarget := buildCoreNetworkBasePolicyDocument(region)
			err := PutAndExecuteCoreNetworkPolicy(ctx, conn, d.Id(), policyDocumentTarget)

			if err != nil {
				return diag.FromErr(err)
			}

			if _, err := waitCoreNetworkUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return diag.Errorf("waiting for Network Manager Core Network (%s) update: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating Network Manager Core Network (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceCoreNetworkRead(ctx, d, meta)
}

func resourceCoreNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	log.Printf("[DEBUG] Deleting Network Manager Core Network: %s", d.Id())
	_, err := conn.DeleteCoreNetworkWithContext(ctx, &networkmanager.DeleteCoreNetworkInput{
		CoreNetworkId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Network Manager Core Network (%s): %s", d.Id(), err)
	}

	if _, err := waitCoreNetworkDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Network Manager Core Network (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func resourceCoreNetworkCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	if d.HasChange("policy_document") {
		if o, n := d.GetChange("policy_document"); !verify.JSONStringsEqual(o.(string), n.(string)) {
			d.SetNewComputed("edges")
			d.SetNewComputed("segments")
		}
	}

	return nil
}

func FindCoreNetworkByID(ctx context.Context, conn *networkmanager.NetworkManager, id string) (*networkmanager.CoreNetwork, error) {
	input := &networkmanager.GetCoreNetworkInput{
		CoreNetworkId: aws.String(id),
	}

	output, err := conn.GetCoreNetworkWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindCoreNetworkPolicyByID(ctx context.Context, conn *networkmanager.NetworkManager, id string) (*networkmanager.CoreNetworkPolicy, error) {
	input := &networkmanager.GetCoreNetworkPolicyInput{
		CoreNetworkId: aws.String(id),
	}

	output, err := conn.GetCoreNetworkPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func statusCoreNetworkState(ctx context.Context, conn *networkmanager.NetworkManager, id string) resource.StateRefreshFunc {
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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.CoreNetworkStateDeleting},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusCoreNetworkState(ctx, conn, id),
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
		tfMap["name"] = aws.StringValue(v)
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
		ClientToken:    aws.String(resource.UniqueId()),
		CoreNetworkId:  aws.String(coreNetworkId),
		PolicyDocument: v,
	})

	if err != nil {
		return fmt.Errorf("putting Network Manager Core Network (%s) policy: %s", coreNetworkId, err)
	}

	policyVersionID := aws.Int64Value(output.CoreNetworkPolicy.PolicyVersionId)

	// new policy documents goes from Pending generation to Ready to execute
	_, err = tfresource.RetryWhen(ctx, 4*time.Minute,
		func() (interface{}, error) {
			return conn.ExecuteCoreNetworkChangeSetWithContext(ctx, &networkmanager.ExecuteCoreNetworkChangeSetInput{
				CoreNetworkId:   aws.String(coreNetworkId),
				PolicyVersionId: aws.Int64(policyVersionID),
			})
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, networkmanager.ErrCodeValidationException, "Incorrect input") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return fmt.Errorf("executing Network Manager Core Network (%s) change set (%d): %s", coreNetworkId, policyVersionID, err)
	}

	return nil
}

// buildCoreNetworkBasePolicyDocument returns a base policy document
func buildCoreNetworkBasePolicyDocument(region string) string {
	return fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"64512-65534\"],\"edge-locations\":[{\"location\":\"%s\"}]},\"segments\":[{\"name\":\"segment\",\"description\":\"base-policy\"}],\"version\":\"2021.12\"}", region)
}
