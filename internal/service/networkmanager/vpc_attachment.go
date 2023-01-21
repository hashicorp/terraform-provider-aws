package networkmanager

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCAttachmentCreate,
		ReadWithoutTimeout:   resourceVPCAttachmentRead,
		UpdateWithoutTimeout: resourceVPCAttachmentUpdate,
		DeleteWithoutTimeout: resourceVPCAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachment_policy_rule_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"attachment_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"edge_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"appliance_mode_support": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"ipv6_support": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"segment_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_arns": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceVPCAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	coreNetworkID := d.Get("core_network_id").(string)
	vpcARN := d.Get("vpc_arn").(string)
	input := &networkmanager.CreateVpcAttachmentInput{
		CoreNetworkId: aws.String(coreNetworkID),
		SubnetArns:    flex.ExpandStringSet(d.Get("subnet_arns").(*schema.Set)),
		VpcArn:        aws.String(vpcARN),
	}

	if v, ok := d.GetOk("options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Options = expandVpcOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Network Manager VPC Attachment: %s", input)
	output, err := conn.CreateVpcAttachmentWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Network Manager VPC (%s) Attachment (%s): %s", vpcARN, coreNetworkID, err)
	}

	d.SetId(aws.StringValue(output.VpcAttachment.Attachment.AttachmentId))

	if _, err := waitVPCAttachmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Network Manager VPC Attachment (%s) create: %s", d.Id(), err)
	}

	return resourceVPCAttachmentRead(ctx, d, meta)
}

func resourceVPCAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	vpcAttachment, err := FindVPCAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager VPC Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Network Manager VPC Attachment (%s): %s", d.Id(), err)
	}

	a := vpcAttachment.Attachment
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "networkmanager",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("attachment/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("attachment_policy_rule_number", a.AttachmentPolicyRuleNumber)
	d.Set("attachment_type", a.AttachmentType)
	d.Set("core_network_arn", a.CoreNetworkArn)
	d.Set("core_network_id", a.CoreNetworkId)
	d.Set("edge_location", a.EdgeLocation)
	if vpcAttachment.Options != nil {
		if err := d.Set("options", []interface{}{flattenVpcOptions(vpcAttachment.Options)}); err != nil {
			return diag.Errorf("setting options: %s", err)
		}
	} else {
		d.Set("options", nil)
	}
	d.Set("owner_account_id", a.OwnerAccountId)
	d.Set("resource_arn", a.ResourceArn)
	d.Set("segment_name", a.SegmentName)
	d.Set("state", a.State)
	d.Set("subnet_arns", aws.StringValueSlice(vpcAttachment.SubnetArns))
	d.Set("vpc_arn", a.ResourceArn)

	tags := KeyValueTags(a.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("Setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceVPCAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &networkmanager.UpdateVpcAttachmentInput{
			AttachmentId: aws.String(d.Id()),
		}

		if d.HasChange("options") {
			if v, ok := d.GetOk("options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.Options = expandVpcOptions(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("subnet_arns") {
			o, n := d.GetChange("subnet_arns")
			if o == nil {
				o = new(schema.Set)
			}
			if n == nil {
				n = new(schema.Set)
			}
			os := o.(*schema.Set)
			ns := n.(*schema.Set)

			if add := ns.Difference(os); len(add.List()) > 0 {
				input.AddSubnetArns = flex.ExpandStringSet(add)
			}

			if del := os.Difference(ns); len(del.List()) > 0 {
				input.RemoveSubnetArns = flex.ExpandStringSet(del)
			}
		}

		_, err := conn.UpdateVpcAttachmentWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Network Manager VPC Attachment (%s): %s", d.Id(), err)
		}

		if _, err := waitVPCAttachmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for Network Manager VPC Attachment (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating Network Manager VPC Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceVPCAttachmentRead(ctx, d, meta)
}

func resourceVPCAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	// If ResourceAttachmentAccepter is used, then VPC Attachment state
	// is not updated from StatePendingAttachmentAcceptance and the delete fails if deleted immediately after create
	output, sErr := FindVPCAttachmentByID(ctx, conn, d.Id())

	if tfawserr.ErrCodeEquals(sErr, networkmanager.ErrCodeResourceNotFoundException) {
		return nil
	}

	if sErr != nil {
		return diag.Errorf("reading Network Manager VPC Attachment (%s): %s", d.Id(), sErr)
	}

	d.Set("state", output.Attachment.State)

	if state := d.Get("state").(string); state == networkmanager.AttachmentStatePendingAttachmentAcceptance || state == networkmanager.AttachmentStatePendingTagAcceptance {
		return diag.Errorf("cannot delete Network Manager VPC Attachment (%s) in %s state", d.Id(), state)
	}

	log.Printf("[DEBUG] Deleting Network Manager VPC Attachment: %s", d.Id())
	_, err := conn.DeleteAttachmentWithContext(ctx, &networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Network Manager VPC Attachment (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCAttachmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Network Manager VPC Attachment (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindVPCAttachmentByID(ctx context.Context, conn *networkmanager.NetworkManager, id string) (*networkmanager.VpcAttachment, error) {
	input := &networkmanager.GetVpcAttachmentInput{
		AttachmentId: aws.String(id),
	}

	output, err := conn.GetVpcAttachmentWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VpcAttachment == nil || output.VpcAttachment.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VpcAttachment, nil
}

func StatusVPCAttachmentState(ctx context.Context, conn *networkmanager.NetworkManager, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCAttachmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Attachment.State), nil
	}
}

func waitVPCAttachmentCreated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.AttachmentStateCreating, networkmanager.AttachmentStatePendingNetworkUpdate},
		Target:  []string{networkmanager.AttachmentStateAvailable, networkmanager.AttachmentStatePendingAttachmentAcceptance},
		Timeout: timeout,
		Refresh: StatusVPCAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVPCAttachmentDeleted(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{networkmanager.AttachmentStateDeleting},
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        StatusVPCAttachmentState(ctx, conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVPCAttachmentUpdated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.AttachmentStateUpdating},
		Target:  []string{networkmanager.AttachmentStateAvailable, networkmanager.AttachmentStatePendingTagAcceptance},
		Timeout: timeout,
		Refresh: StatusVPCAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func expandVpcOptions(tfMap map[string]interface{}) *networkmanager.VpcOptions { // nosemgrep:ci.caps5-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &networkmanager.VpcOptions{}

	if v, ok := tfMap["appliance_mode_support"].(bool); ok {
		apiObject.ApplianceModeSupport = aws.Bool(v)
	}

	if v, ok := tfMap["ipv6_support"].(bool); ok {
		apiObject.Ipv6Support = aws.Bool(v)
	}

	return apiObject
}

func flattenVpcOptions(apiObject *networkmanager.VpcOptions) map[string]interface{} { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ApplianceModeSupport; v != nil {
		tfMap["appliance_mode_support"] = aws.BoolValue(v)
	}

	if v := apiObject.Ipv6Support; v != nil {
		tfMap["ipv6_support"] = aws.BoolValue(v)
	}

	return tfMap
}
