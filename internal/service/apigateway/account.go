package apigateway

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountUpdate,
		ReadWithoutTimeout:   resourceAccountRead,
		UpdateWithoutTimeout: resourceAccountUpdate,
		DeleteWithoutTimeout: resourceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cloudwatch_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"throttle_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"burst_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"rate_limit": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	account, err := conn.GetAccountWithContext(ctx, &apigateway.GetAccountInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Account: %s", err)
	}

	log.Printf("[DEBUG] Received API Gateway Account: %s", account)

	if _, ok := d.GetOk("cloudwatch_role_arn"); ok {
		// CloudwatchRoleArn cannot be empty nor made empty via API
		// This resource can however be useful w/out defining cloudwatch_role_arn
		// (e.g. for referencing throttle_settings)
		d.Set("cloudwatch_role_arn", account.CloudwatchRoleArn)
	}
	if err := d.Set("throttle_settings", FlattenThrottleSettings(account.ThrottleSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Account: %s", err)
	}

	return diags
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	input := apigateway.UpdateAccountInput{}
	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("cloudwatch_role_arn") {
		arn := d.Get("cloudwatch_role_arn").(string)
		if len(arn) > 0 {
			// Unfortunately AWS API doesn't allow empty ARNs,
			// even though that's default settings for new AWS accounts
			// BadRequestException: The role ARN is not well formed
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String("replace"),
				Path:  aws.String("/cloudwatchRoleArn"),
				Value: aws.String(arn),
			})
		}
	}
	input.PatchOperations = operations

	log.Printf("[INFO] Updating API Gateway Account: %s", input)

	// Retry due to eventual consistency of IAM
	expectedErrMsg := "The role ARN does not have required permissions"
	otherErrMsg := "API Gateway could not successfully write to CloudWatch Logs using the ARN specified"
	var out *apigateway.Account
	var err error
	err = resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		out, err = conn.UpdateAccountWithContext(ctx, &input)

		if err != nil {
			if tfawserr.ErrMessageContains(err, "BadRequestException", expectedErrMsg) ||
				tfawserr.ErrMessageContains(err, "BadRequestException", otherErrMsg) {
				log.Printf("[DEBUG] Retrying API Gateway Account update: %s", err)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.UpdateAccountWithContext(ctx, &input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Updating API Gateway Account failed: %s", err)
	}
	log.Printf("[DEBUG] API Gateway Account updated: %s", out)

	d.SetId("api-gateway-account")
	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var
	// There is no API for "deleting" account or resetting it to "default" settings
	diags diag.Diagnostics

	return diags
}
