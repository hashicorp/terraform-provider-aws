package ecrpublic

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRepositoryPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryPolicyPut,
		ReadWithoutTimeout:   resourceRepositoryPolicyRead,
		UpdateWithoutTimeout: resourceRepositoryPolicyPut,
		DeleteWithoutTimeout: resourceRepositoryPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:                  schema.TypeString,
				Required:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	policyPutTimeout = 2 * time.Minute
)

func resourceRepositoryPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRPublicConn()

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	repositoryName := d.Get("repository_name").(string)
	input := &ecrpublic.SetRepositoryPolicyInput{
		PolicyText:     aws.String(policy),
		RepositoryName: aws.String(repositoryName),
	}

	log.Printf("[DEBUG] Setting ECR Public Repository Policy: %s", input)
	outputRaw, err := tfresource.RetryWhen(ctx, policyPutTimeout,
		func() (interface{}, error) {
			return conn.SetRepositoryPolicyWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, ecrpublic.ErrCodeInvalidParameterException, "Invalid repository policy provided") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ECR Public Repository (%s) Policy: %s", repositoryName, err)
	}

	if d.IsNewResource() {
		d.SetId(aws.StringValue(outputRaw.(*ecrpublic.SetRepositoryPolicyOutput).RepositoryName))
	}

	return append(diags, resourceRepositoryPolicyRead(ctx, d, meta)...)
}

func resourceRepositoryPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRPublicConn()

	output, err := FindRepositoryPolicyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Public Repository Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Public Repository Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.StringValue(output.PolicyText))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is an invalid JSON: %s", policyToSet, err)
	}

	d.Set("policy", policyToSet)
	d.Set("registry_id", output.RegistryId)
	d.Set("repository_name", output.RepositoryName)

	return diags
}

func resourceRepositoryPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRPublicConn()

	_, err := conn.DeleteRepositoryPolicyWithContext(ctx, &ecrpublic.DeleteRepositoryPolicyInput{
		RegistryId:     aws.String(d.Get("registry_id").(string)),
		RepositoryName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ecrpublic.ErrCodeRepositoryNotFoundException, ecrpublic.ErrCodeRepositoryPolicyNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Public Repository Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func FindRepositoryPolicyByName(ctx context.Context, conn *ecrpublic.ECRPublic, name string) (*ecrpublic.GetRepositoryPolicyOutput, error) {
	input := &ecrpublic.GetRepositoryPolicyInput{
		RepositoryName: aws.String(name),
	}

	output, err := conn.GetRepositoryPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ecrpublic.ErrCodeRepositoryNotFoundException, ecrpublic.ErrCodeRepositoryPolicyNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
