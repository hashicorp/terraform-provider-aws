package s3control

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	_sp.registerSDKResourceFactory("aws_s3control_access_point_policy", resourceAccessPointPolicy)
}

func resourceAccessPointPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessPointPolicyCreate,
		ReadWithoutTimeout:   resourceAccessPointPolicyRead,
		UpdateWithoutTimeout: resourceAccessPointPolicyUpdate,
		DeleteWithoutTimeout: resourceAccessPointPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAccessPointPolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"access_point_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"has_public_access_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"policy": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceAccessPointPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	resourceID, err := AccessPointCreateResourceID(d.Get("access_point_arn").(string))

	if err != nil {
		return diag.FromErr(err)
	}

	accountID, name, err := AccessPointParseResourceID(resourceID)

	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
	}

	input := &s3control.PutAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Policy:    aws.String(policy),
	}

	_, err = conn.PutAccessPointPolicyWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating S3 Access Point (%s) Policy: %s", resourceID, err)
	}

	d.SetId(resourceID)

	return resourceAccessPointPolicyRead(ctx, d, meta)
}

func resourceAccessPointPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	accountID, name, err := AccessPointParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	policy, status, err := FindAccessPointPolicyAndStatusByTwoPartKey(ctx, conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Access Point Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Access Point Policy (%s): %s", d.Id(), err)
	}

	d.Set("has_public_access_policy", status.IsPublic)

	if policy != "" {
		policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), policy)
		if err != nil {
			return diag.FromErr(err)
		}

		d.Set("policy", policyToSet)
	} else {
		d.Set("policy", "")
	}

	return nil
}

func resourceAccessPointPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	accountID, name, err := AccessPointParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
	}

	input := &s3control.PutAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Policy:    aws.String(policy),
	}

	_, err = conn.PutAccessPointPolicyWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating S3 Access Point Policy (%s): %s", d.Id(), err)
	}

	return resourceAccessPointPolicyRead(ctx, d, meta)
}

func resourceAccessPointPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	accountID, name, err := AccessPointParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting S3 Access Point Policy: %s", d.Id())
	_, err = conn.DeleteAccessPointPolicyWithContext(ctx, &s3control.DeleteAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Access Point Policy (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAccessPointPolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceID, err := AccessPointCreateResourceID(d.Id())

	if err != nil {
		return nil, err
	}

	d.Set("access_point_arn", d.Id())
	d.SetId(resourceID)

	return []*schema.ResourceData{d}, nil
}

func FindAccessPointPolicyAndStatusByTwoPartKey(ctx context.Context, conn *s3control.S3Control, accountID string, name string) (string, *s3control.PolicyStatus, error) {
	input1 := &s3control.GetAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output1, err := conn.GetAccessPointPolicyWithContext(ctx, input1)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return "", nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input1,
		}
	}

	if err != nil {
		return "", nil, err
	}

	if output1 == nil {
		return "", nil, tfresource.NewEmptyResultError(input1)
	}

	policy := aws.StringValue(output1.Policy)

	if policy == "" {
		return "", nil, tfresource.NewEmptyResultError(input1)
	}

	input2 := &s3control.GetAccessPointPolicyStatusInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output2, err := conn.GetAccessPointPolicyStatusWithContext(ctx, input2)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return "", nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input2,
		}
	}

	if err != nil {
		return "", nil, err
	}

	if output2 == nil || output2.PolicyStatus == nil {
		return "", nil, tfresource.NewEmptyResultError(input2)
	}

	return policy, output2.PolicyStatus, nil
}
