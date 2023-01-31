package macie

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMemberAccountAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMemberAccountAssociationCreate,
		ReadWithoutTimeout:   resourceMemberAccountAssociationRead,
		DeleteWithoutTimeout: resourceMemberAccountAssociationDelete,

		Schema: map[string]*schema.Schema{
			"member_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},
	}
}

func resourceMemberAccountAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	memberAccountId := d.Get("member_account_id").(string)
	req := &macie.AssociateMemberAccountInput{
		MemberAccountId: aws.String(memberAccountId),
	}

	log.Printf("[DEBUG] Creating Macie member account association: %#v", req)
	_, err := conn.AssociateMemberAccountWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie member account association: %s", err)
	}

	d.SetId(memberAccountId)
	return append(diags, resourceMemberAccountAssociationRead(ctx, d, meta)...)
}

func resourceMemberAccountAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	req := &macie.ListMemberAccountsInput{}

	var res *macie.MemberAccount
	err := conn.ListMemberAccountsPagesWithContext(ctx, req, func(page *macie.ListMemberAccountsOutput, lastPage bool) bool {
		for _, v := range page.MemberAccounts {
			if aws.StringValue(v.AccountId) == d.Get("member_account_id").(string) {
				res = v
				return false
			}
		}

		return true
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Macie member account associations: %s", err)
	}

	if res == nil {
		log.Printf("[WARN] Macie member account association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	return diags
}

func resourceMemberAccountAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	log.Printf("[DEBUG] Deleting Macie member account association: %s", d.Id())

	_, err := conn.DisassociateMemberAccountWithContext(ctx, &macie.DisassociateMemberAccountInput{
		MemberAccountId: aws.String(d.Get("member_account_id").(string)),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, macie.ErrCodeInvalidInputException, "is a master Macie account and cannot be disassociated") {
			log.Printf("[INFO] Macie master account (%s) cannot be disassociated, removing from state", d.Id())
			return diags
		} else if tfawserr.ErrMessageContains(err, macie.ErrCodeInvalidInputException, "is not yet associated with Macie") {
			return diags
		} else {
			return sdkdiag.AppendErrorf(diags, "deleting Macie member account association: %s", err)
		}
	}

	return diags
}
