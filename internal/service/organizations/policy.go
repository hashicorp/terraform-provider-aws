package organizations

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_policy", name="Policy")
// @Tags(identifierAttribute="id")
func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourcePolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				ValidateFunc:     validation.StringIsJSON,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"skip_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      organizations.PolicyTypeServiceControlPolicy,
				ValidateFunc: validation.StringInSlice(organizations.PolicyType_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	name := d.Get("name").(string)
	input := &organizations.CreatePolicyInput{
		Content:     aws.String(d.Get("content").(string)),
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(name),
		Type:        aws.String(d.Get("type").(string)),
		Tags:        GetTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 4*time.Minute, func() (interface{}, error) {
		return conn.CreatePolicyWithContext(ctx, input)
	}, organizations.ErrCodeFinalizingOrganizationException)

	if err != nil {
		return diag.Errorf("creating Organizations Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*organizations.CreatePolicyOutput).Policy.PolicySummary.Id))

	return resourcePolicyRead(ctx, d, meta)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	input := &organizations.DescribePolicyInput{
		PolicyId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Organizations policy: %s", input)
	resp, err := conn.DescribePolicyWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, organizations.ErrCodePolicyNotFoundException) {
		log.Printf("[WARN] Organizations policy does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Organizations Policy (%s): %s", d.Id(), err)
	}

	if resp.Policy == nil || resp.Policy.PolicySummary == nil {
		if !d.IsNewResource() {
			log.Printf("[WARN] Organizations policy does not exist, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}

		return diag.FromErr(&retry.NotFoundError{})
	}

	d.Set("arn", resp.Policy.PolicySummary.Arn)
	d.Set("content", resp.Policy.Content)
	d.Set("description", resp.Policy.PolicySummary.Description)
	d.Set("name", resp.Policy.PolicySummary.Name)
	d.Set("type", resp.Policy.PolicySummary.Type)

	if aws.BoolValue(resp.Policy.PolicySummary.AwsManaged) {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "AWS-managed Organizations policies cannot be imported",
				Detail:   fmt.Sprintf("This resource should be removed from your Terraform state using `terraform state rm` (https://www.terraform.io/docs/commands/state/rm.html) and references should use the ID (%s) directly.", d.Id()),
			},
		}
	}

	return nil
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	input := &organizations.UpdatePolicyInput{
		PolicyId: aws.String(d.Id()),
	}

	if d.HasChange("content") {
		input.Content = aws.String(d.Get("content").(string))
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("name") {
		input.Name = aws.String(d.Get("name").(string))
	}

	_, err := conn.UpdatePolicyWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating Organizations policy (%s): %s", d.Id(), err)
	}

	return resourcePolicyRead(ctx, d, meta)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	if v, ok := d.GetOk("skip_destroy"); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Organizations Policy: %s", d.Id())
		return nil
	}

	input := &organizations.DeletePolicyInput{
		PolicyId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Organizations Policy: %s", d.Id())
	_, err := conn.DeletePolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodePolicyNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Organizations policy (%s): %s", d.Id(), err)
	}

	return nil
}

func resourcePolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	input := &organizations.DescribePolicyInput{
		PolicyId: aws.String(d.Id()),
	}
	resp, err := conn.DescribePolicyWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if aws.BoolValue(resp.Policy.PolicySummary.AwsManaged) {
		return nil, fmt.Errorf("AWS-managed Organizations policy (%s) cannot be imported. Use the policy ID directly in your configuration.", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
