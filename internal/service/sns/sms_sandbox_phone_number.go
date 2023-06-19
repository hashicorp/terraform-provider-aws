package sns

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/comprehend/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_sns_sms_sandbox_phone_number")
func ResourceSMSSandboxPhoneNumber() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSMSSandboxPhoneNumberCreate,
		ReadWithoutTimeout:   resourceSMSSandboxPhoneNumberRead,
		DeleteWithoutTimeout: resourceSMSSandboxPhoneNumberDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"phone_number": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"language_code": {
				Type:             schema.TypeString,
				Required:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.SyntaxLanguageCode](),
			},
		},
	}
}

func resourceSMSSandboxPhoneNumberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSConn(ctx)

	phoneNumber := d.Get("phone_number").(string)
	input := &sns.CreateSMSSandboxPhoneNumberInput{
		PhoneNumber:  aws.String(phoneNumber),
		LanguageCode: aws.String(d.Get("language_code").(string)),
	}

	_, err := conn.CreateSMSSandboxPhoneNumberWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SNS SMS Sandbox Phone Number (%s): %s", d.Id(), err)
	}

	if d.IsNewResource() {
		d.SetId(phoneNumber)
	}

	return resourceSMSSandboxPhoneNumberRead(ctx, d, meta)
}

func resourceSMSSandboxPhoneNumberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSConn(ctx)

	phoneNumber := ""
	status := ""

	err := conn.ListSMSSandboxPhoneNumbersPagesWithContext(ctx, &sns.ListSMSSandboxPhoneNumbersInput{}, func(page *sns.ListSMSSandboxPhoneNumbersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, i := range page.PhoneNumbers {
			phoneNumber = aws.StringValue(i.PhoneNumber)
			status = aws.StringValue(i.Status)

			if phoneNumber == d.Id() {
				break
			}
		}

		return !lastPage
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, sns.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] SNS SMS Sandbox Phone Number (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SNS SMS Sandbox Phone Number: %s", err)
	}

	d.Set("phone_number", phoneNumber)
	d.Set("status", status)

	return diags
}

func resourceSMSSandboxPhoneNumberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSConn(ctx)

	_, err := conn.PutDataProtectionPolicyWithContext(ctx, &sns.PutDataProtectionPolicyInput{
		DataProtectionPolicy: aws.String(""),
		ResourceArn:          aws.String(d.Get("arn").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SNS Data Protection Policy (%s): %s", d.Id(), err)
	}

	return diags
}
