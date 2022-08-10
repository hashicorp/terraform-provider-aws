package networkmanager

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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

func ResourceVpcAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVpcAttachmentCreate,
		ReadWithoutTimeout:   resourceVpcAttachmentRead,
		UpdateWithoutTimeout: resourceVpcAttachmentUpdate,
		DeleteWithoutTimeout: resourceVpcAttachmentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
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
						"ipv6_support": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
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

func resourceVpcAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &networkmanager.CreateVpcAttachmentInput{
		CoreNetworkId: aws.String(d.Get("core_network_id").(string)),
		VpcArn:        aws.String(d.Get("vpc_arn").(string)),
		SubnetArns:    flex.ExpandStringSet(d.Get("subnet_arns").(*schema.Set)),
	}

	if v, ok := d.GetOk("options"); ok {
		optsMap := v.([]interface{})[0].(map[string]interface{})
		input.Options = &networkmanager.VpcOptions{
			Ipv6Support: aws.Bool(optsMap["ipv6_support"].(bool)),
		}
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Network Manager VPC Attachment: %s", input)
	output, err := conn.CreateVpcAttachmentWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating Network Manager VPC Attachment: %s", err)
	}

	d.SetId(aws.StringValue(output.VpcAttachment.Attachment.AttachmentId))

	if _, err := waitVpcAttachmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for Network Manager VPC Attachment (%s) create: %s", d.Id(), err)
	}

	return resourceVpcAttachmentRead(ctx, d, meta)
}

func resourceVpcAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	vpcAttachment, err := conn.GetVpcAttachment(&networkmanager.GetVpcAttachmentInput{
		AttachmentId: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Network Manager VPC Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Network Manager VPC Attachment (%s): %s", d.Id(), err)
	}

	a := vpcAttachment.VpcAttachment.Attachment
	subnetArns := vpcAttachment.VpcAttachment.SubnetArns
	opts := vpcAttachment.VpcAttachment.Options

	d.Set("core_network_id", a.CoreNetworkId)
	d.Set("state", a.State)
	d.Set("core_network_arn", a.CoreNetworkArn)
	d.Set("attachment_policy_rule_number", a.AttachmentPolicyRuleNumber)
	d.Set("attachment_type", a.AttachmentType)
	d.Set("edge_location", a.EdgeLocation)
	d.Set("owner_account_id", a.OwnerAccountId)
	d.Set("resource_arn", a.ResourceArn)
	d.Set("segment_name", a.SegmentName)

	// VPC arn is not outputted, therefore use resource arn
	d.Set("vpc_arn", a.ResourceArn)

	// options
	d.Set("options", []interface{}{map[string]interface{}{
		"ipv6_support": aws.BoolValue(opts.Ipv6Support),
	}})

	// subnetArns
	d.Set("subnet_arns", subnetArns)

	tags := KeyValueTags(a.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceVpcAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		acnt := meta.(*conns.AWSClient).AccountID
		arn := fmt.Sprintf("arn:aws:networkmanager::%s:attachment/%s", acnt, d.Id())

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return diag.Errorf("error updating VPC Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChangesExcept("tags", "tags_all") {
		input := &networkmanager.UpdateVpcAttachmentInput{
			AttachmentId: aws.String(d.Id()),
		}

		if d.HasChange("options") {
			input.Options = expandVpcAttachmentOptions(d.Get("options").([]interface{}))
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
			subnetArnsUpdateAdd := ns.Difference(os)
			subnetArnsUpdateRemove := os.Difference(ns)

			if len(subnetArnsUpdateAdd.List()) > 0 {
				input.AddSubnetArns = flex.ExpandStringSet(subnetArnsUpdateAdd)
			}

			if len(subnetArnsUpdateRemove.List()) > 0 {
				input.RemoveSubnetArns = flex.ExpandStringSet(subnetArnsUpdateRemove)
			}
		}
		_, err := conn.UpdateVpcAttachmentWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error updating vpc attachment (%s): %s", d.Id(), err)
		}

		if _, err := waitVpcAttachmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("error waiting for Network Manager VPC Attachment (%s) update: %s", d.Id(), err)
		}
	}

	return resourceVpcAttachmentRead(ctx, d, meta)
}

func expandVpcAttachmentOptions(l []interface{}) *networkmanager.VpcOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	opts := &networkmanager.VpcOptions{
		Ipv6Support: aws.Bool(bool(m["ipv6_support"].(bool))),
	}

	return opts
}

func resourceVpcAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	log.Printf("[DEBUG] Deleting Network Manager VPC Attachment: %s", d.Id())

	state := d.Get("state").(string)

	if state == networkmanager.AttachmentStatePendingAttachmentAcceptance || state == networkmanager.AttachmentStatePendingTagAcceptance {
		return diag.Errorf("error deleting Network Manager VPC Attachment (%s): Cannot delete attachment that is pending acceptance.", d.Id())
	}

	_, err := conn.DeleteAttachmentWithContext(ctx, &networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Network Manager VPC Attachment (%s): %s", d.Id(), err)
	}

	if _, err := waitVpcAttachmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("error waiting for Network Manager VPC Attachment (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func VpcAttachmentIDNotFoundError(err error) bool {
	return validationExceptionMessageContains(err, networkmanager.ValidationExceptionReasonFieldValidationFailed, "VPC Attachment not found")
}

func FindVpcAttachmentByID(ctx context.Context, conn *networkmanager.NetworkManager, id string) (*networkmanager.VpcAttachment, error) {

	output, err := conn.GetVpcAttachment(&networkmanager.GetVpcAttachmentInput{
		AttachmentId: aws.String(id),
	})

	if err != nil {
		return nil, err
	}

	return output.VpcAttachment, nil
}

func StatusVpcAttachmentState(ctx context.Context, conn *networkmanager.NetworkManager, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVpcAttachmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Attachment.State), nil
	}
}

func waitVpcAttachmentCreated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.AttachmentStateCreating, networkmanager.AttachmentStatePendingNetworkUpdate},
		Target:  []string{networkmanager.AttachmentStateAvailable, networkmanager.AttachmentStatePendingAttachmentAcceptance},
		Timeout: timeout,
		Refresh: StatusVpcAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVpcAttachmentDeleted(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{networkmanager.AttachmentStateDeleting},
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        StatusVpcAttachmentState(ctx, conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVpcAttachmentUpdated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.AttachmentStateUpdating},
		Target:  []string{networkmanager.AttachmentStateAvailable, networkmanager.AttachmentStatePendingTagAcceptance},
		Timeout: timeout,
		Refresh: StatusVpcAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

const (
	VpcAttachmentValidationExceptionTimeout = 2 * time.Minute
)
