package ec2

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_ec2_instance_connect_endpoint", name="Instance Connect Endpoint")

func ResourceInstanceConnectEndpoint() *schema.Resource {
	return &schema.Resource{
		// TIP: ==== ASSIGN CRUD FUNCTIONS ====
		// These 4 functions handle CRUD responsibilities below.
		CreateWithoutTimeout: resourceInstanceConnectEndpointCreate,
		ReadWithoutTimeout:   resourceInstanceConnectEndpointRead,
		UpdateWithoutTimeout: resourceInstanceConnectEndpointUpdate,
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
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"prevent_client_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	in := &ec2.CreateInstanceConnectEndpointInput{
		ClientToken:      aws.String(id.UniqueId()),
		SubnetId:         aws.String(d.Get("subnet_id").(string)),
		PreserveClientIp: aws.Bool(d.Get("prevent_client_ip").(bool)),
		//TagSpecifications: getTagSpecificationsIn(ctx, ec2.Create),
	}

	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		in.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	out, err := conn.CreateInstanceConnectEndpoint(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionCreating, ResNameInstanceConnectEndpoint, d.Get("name").(string), err)...)
	}

	if out == nil || out.InstanceConnectEndpoint == nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionCreating, ResNameInstanceConnectEndpoint, d.Get("name").(string), errors.New("empty output"))...)
	}

	d.SetId(aws.ToString(out.InstanceConnectEndpoint.InstanceConnectEndpointId))

	if _, err := waitInstanceConnectEndpointCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionWaitingForCreation, ResNameInstanceConnectEndpoint, d.Id(), err)...)
	}

	return append(diags, resourceInstanceConnectEndpointRead(ctx, d, meta)...)
}

func resourceInstanceConnectEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	out, err := findInstanceConnectEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 InstanceConnectEndpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionReading, ResNameInstanceConnectEndpoint, d.Id(), err)...)
	}

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)

	if err := d.Set("complex_argument", flattenComplexArguments(out.ComplexArguments)); err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionSetting, ResNameInstanceConnectEndpoint, d.Id(), err)...)
	}

	p, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.Policy))
	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionSetting, ResNameInstanceConnectEndpoint, d.Id(), err)...)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionSetting, ResNameInstanceConnectEndpoint, d.Id(), err)...)
	}

	d.Set("policy", p)

	// TIP: -- 6. Return diags
	return diags
}

func resourceInstanceConnectEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	update := false

	in := &ec2.UpdateInstanceConnectEndpointInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChanges("an_argument") {
		in.AnArgument = aws.String(d.Get("an_argument").(string))
		update = true
	}

	if !update {
		// TIP: If update doesn't do anything at all, which is rare, you can
		// return diags. Otherwise, return a read call, as below.
		return diags
	}

	log.Printf("[DEBUG] Updating EC2 InstanceConnectEndpoint (%s): %#v", d.Id(), in)
	out, err := conn.UpdateInstanceConnectEndpoint(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionUpdating, ResNameInstanceConnectEndpoint, d.Id(), err)...)
	}

	if _, err := waitInstanceConnectEndpointUpdated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionWaitingForUpdate, ResNameInstanceConnectEndpoint, d.Id(), err)...)
	}

	return append(diags, resourceInstanceConnectEndpointRead(ctx, d, meta)...)
}

func resourceInstanceConnectEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 InstanceConnectEndpoint %s", d.Id())

	_, err := conn.DeleteInstanceConnectEndpoint(ctx, &ec2.DeleteInstanceConnectEndpointInput{
		Id: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}
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

func waitInstanceConnectEndpointCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*ec2.InstanceConnectEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusInstanceConnectEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ec2.InstanceConnectEndpoint); ok {
		return out, err
	}

	return nil, err
}

func waitInstanceConnectEndpointUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*ec2.InstanceConnectEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusInstanceConnectEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ec2.InstanceConnectEndpoint); ok {
		return out, err
	}

	return nil, err
}

func waitInstanceConnectEndpointDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*ec2.InstanceConnectEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusInstanceConnectEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ec2.InstanceConnectEndpoint); ok {
		return out, err
	}

	return nil, err
}

func statusInstanceConnectEndpoint(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findInstanceConnectEndpointByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

//func findInstanceConnectEndpointByID(ctx context.Context, conn *ec2.Client, id string) (*ec2.DescribeInstanceConnectEndpointsOutput, error) {
//	in := ec2.DescribeInstanceConnectEndpointsInput{
//		InstanceConnectEndpointIds: aws.StringSlice([]string{id}),
//	}
//	out, err := conn.DescribeInstanceConnectEndpoints(ctx, in)
//	if errs.IsA[*types.ResourceNotFoundException](err) {
//		return nil, &retry.NotFoundError{
//			LastError:   err,
//			LastRequest: in,
//		}
//	}
//	if err != nil {
//		return nil, err
//	}
//
//	if out == nil || out.InstanceConnectEndpoint == nil {
//		return nil, tfresource.NewEmptyResultError(in)
//	}
//
//	return out.InstanceConnectEndpoint, nil
//}

func flattenComplexArgument(apiObject *ec2.ComplexArgument) map[string]interface{} {
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

func flattenComplexArguments(apiObjects []*ec2.ComplexArgument) []interface{} {
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

func expandComplexArgument(tfMap map[string]interface{}) *ec2.ComplexArgument {
	if tfMap == nil {
		return nil
	}

	if tfMap == nil {
		return nil
	}

	a := &ec2.ComplexArgument{}

	if v, ok := tfMap["sub_field_one"].(string); ok && v != "" {
		a.SubFieldOne = aws.String(v)
	}

	if v, ok := tfMap["sub_field_two"].(string); ok && v != "" {
		a.SubFieldTwo = aws.String(v)
	}

	return a
}

func expandComplexArguments(tfList []interface{}) []*ec2.ComplexArgument {

	if len(tfList) == 0 {
		return nil
	}

	var s []*ec2.ComplexArgument

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
