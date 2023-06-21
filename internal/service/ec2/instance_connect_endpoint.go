package ec2

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_ec2_instance_connect_endpoint")
// @Tags(identifierAttribute="arn")
func ResourceInstanceConnectEndpoint() *schema.Resource {
	return &schema.Resource{
		// TIP: ==== ASSIGN CRUD FUNCTIONS ====
		// These 4 functions handle CRUD responsibilities below.
		CreateWithoutTimeout: resourceInstanceConnectEndpointCreate,
		ReadWithoutTimeout:   resourceInstanceConnectEndpointRead,
		UpdateWithoutTimeout: resourceInstanceConnectEndpointRead,
		DeleteWithoutTimeout: resourceInstanceConnectEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"preserve_client_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"fips_dns_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"network_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"state_message": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameInstanceConnectEndpoint = "Instance Connect Endpoint"
)

func resourceInstanceConnectEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	in := &ec2.CreateInstanceConnectEndpointInput{
		ClientToken:       aws.String(id.UniqueId()),
		SubnetId:          aws.String(d.Get("subnet_id").(string)),
		PreserveClientIp:  aws.Bool(d.Get("preserve_client_ip").(bool)),
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeInstanceConnectEndpoint)}

	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		in.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	out, err := conn.CreateInstanceConnectEndpointWithContext(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionCreating, ResNameInstanceConnectEndpoint, d.Get("name").(string), err)...)
	}

	if out == nil || out.InstanceConnectEndpoint == nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionCreating, ResNameInstanceConnectEndpoint, d.Get("name").(string), errors.New("empty output"))...)
	}

	d.SetId(aws.StringValue(out.InstanceConnectEndpoint.InstanceConnectEndpointId))

	if _, err = waitInstanceConnectEndpointAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Instance Connect Endpoint create: %s", err)
	}

	if tags := GetTagsIn(ctx); in.TagSpecifications == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceInstanceConnectEndpointRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting EC2 Instance Connect Endpoint tags: %s", err)
		}
	}

	return append(diags, resourceInstanceConnectEndpointRead(ctx, d, meta)...)
}

func resourceInstanceConnectEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	out, err := findInstanceConnectEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 InstanceConnectEndpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionReading, ResNameInstanceConnectEndpoint, d.Id(), err)...)
	}

	d.Set("availability_zone", out.AvailabilityZone)
	d.Set("dns_name", out.DnsName)
	d.Set("fips_dns_name", out.FipsDnsName)
	d.Set("arn", out.InstanceConnectEndpointArn)
	d.Set("endpoint_id", out.InstanceConnectEndpointId)
	d.Set("security_group_ids", out.SecurityGroupIds)
	d.Set("state", out.State)
	d.Set("state_message", out.StateMessage)
	d.Set("subnet_id", out.SubnetId)

	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionSetting, ResNameInstanceConnectEndpoint, d.Id(), err)...)
	}

	return diags
}

func resourceInstanceConnectEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[INFO] Deleting EC2 InstanceConnectEndpoint %s", d.Id())

	_, err := conn.DeleteInstanceConnectEndpointWithContext(ctx, &ec2.DeleteInstanceConnectEndpointInput{
		InstanceConnectEndpointId: aws.String(d.Id()),
	})

	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionDeleting, ResNameInstanceConnectEndpoint, d.Id(), err)...)
	}

	if _, err := waitInstanceConnectEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionWaitingForDeletion, ResNameInstanceConnectEndpoint, d.Id(), err)...)
	}

	return diags
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitInstanceConnectEndpointAvailable(ctx context.Context, conn *ec2.EC2, instanceConnectEndpointId string, timeout time.Duration) (*ec2.Ec2InstanceConnectEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2ICECreateInProgress},
		Target:     []string{ec2ICECreateComplete},
		Timeout:    timeout,
		Refresh:    statusInstanceConnectEndpoint(ctx, conn, instanceConnectEndpointId),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Ec2InstanceConnectEndpoint); ok {
		if state, lastError := aws.StringValue(output.State), output.StateMessage; state == ec2ICECreateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s", aws.StringValue(output.StateMessage))) // Check here False negative
		}

		return output, err
	}

	return nil, err
}

func waitInstanceConnectEndpointDeleted(ctx context.Context, conn *ec2.EC2, instanceConnectEndpointId string, timeout time.Duration) (*ec2.Ec2InstanceConnectEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2ICEDeleteInProgress, statusNormal},
		Target:  []string{},
		Refresh: statusInstanceConnectEndpoint(ctx, conn, instanceConnectEndpointId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ec2.Ec2InstanceConnectEndpoint); ok {
		return out, err
	}

	return nil, err
}

func statusInstanceConnectEndpoint(ctx context.Context, conn *ec2.EC2, instanceConnectEndpointId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findInstanceConnectEndpointByID(ctx, conn, instanceConnectEndpointId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.StringValue(out.State), nil
	}
}

func findInstanceConnectEndpointByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Ec2InstanceConnectEndpoint, error) {
	in := &ec2.DescribeInstanceConnectEndpointsInput{
		InstanceConnectEndpointIds: aws.StringSlice([]string{id}),
	}

	out, err := findInstanceConnectEndpoint(ctx, conn, in)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(out.State); state == ec2ICEDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: in,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(out.InstanceConnectEndpointId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: in,
		}
	}

	return out, nil
}

func findInstanceConnectEndpoint(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInstanceConnectEndpointsInput) (*ec2.Ec2InstanceConnectEndpoint, error) {
	out, err := findInstanceConnectEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(out) == 0 || out[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out[0], nil
}

func findInstanceConnectEndpoints(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInstanceConnectEndpointsInput) ([]*ec2.Ec2InstanceConnectEndpoint, error) {
	var out []*ec2.Ec2InstanceConnectEndpoint

	err := conn.DescribeInstanceConnectEndpointsPagesWithContext(ctx, input, func(page *ec2.DescribeInstanceConnectEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InstanceConnectEndpoints {
			if v != nil {
				out = append(out, v)
			}
		}

		return !lastPage
	})

	//if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIdNotFound) {
	//	return nil, &retry.NotFoundError{
	//		LastError:   err,
	//		LastRequest: input,
	//	}
	//}

	if err != nil {
		return nil, err
	}

	return out, nil
}
