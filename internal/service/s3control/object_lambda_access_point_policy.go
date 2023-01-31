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
	_sp.registerSDKResourceFactory("aws_s3control_object_lambda_access_point_policy", resourceObjectLambdaAccessPointPolicy)
}

func resourceObjectLambdaAccessPointPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceObjectLambdaAccessPointPolicyCreate,
		ReadWithoutTimeout:   resourceObjectLambdaAccessPointPolicyRead,
		UpdateWithoutTimeout: resourceObjectLambdaAccessPointPolicyUpdate,
		DeleteWithoutTimeout: resourceObjectLambdaAccessPointPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"has_public_access_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceObjectLambdaAccessPointPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}
	name := d.Get("name").(string)
	resourceID := ObjectLambdaAccessPointCreateResourceID(accountID, name)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
	}

	input := &s3control.PutAccessPointPolicyForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Policy:    aws.String(policy),
	}

	_, err = conn.PutAccessPointPolicyForObjectLambdaWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating S3 Object Lambda Access Point (%s) Policy: %s", resourceID, err)
	}

	d.SetId(resourceID)

	return resourceObjectLambdaAccessPointPolicyRead(ctx, d, meta)
}

func resourceObjectLambdaAccessPointPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	policy, status, err := FindObjectLambdaAccessPointPolicyAndStatusByTwoPartKey(ctx, conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Object Lambda Access Point Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Object Lambda Access Point Policy (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("has_public_access_policy", status.IsPublic)
	d.Set("name", name)

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

func resourceObjectLambdaAccessPointPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
	}

	input := &s3control.PutAccessPointPolicyForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Policy:    aws.String(policy),
	}

	_, err = conn.PutAccessPointPolicyForObjectLambdaWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating S3 Object Lambda Access Point Policy (%s): %s", d.Id(), err)
	}

	return resourceObjectLambdaAccessPointPolicyRead(ctx, d, meta)
}

func resourceObjectLambdaAccessPointPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting S3 Object Lambda Access Point Policy: %s", d.Id())
	_, err = conn.DeleteAccessPointPolicyForObjectLambdaWithContext(ctx, &s3control.DeleteAccessPointPolicyForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Object Lambda Access Point Policy (%s): %s", d.Id(), err)
	}

	return nil
}

func FindObjectLambdaAccessPointPolicyAndStatusByTwoPartKey(ctx context.Context, conn *s3control.S3Control, accountID string, name string) (string, *s3control.PolicyStatus, error) {
	input1 := &s3control.GetAccessPointPolicyForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output1, err := conn.GetAccessPointPolicyForObjectLambdaWithContext(ctx, input1)

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

	input2 := &s3control.GetAccessPointPolicyStatusForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output2, err := conn.GetAccessPointPolicyStatusForObjectLambdaWithContext(ctx, input2)

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
