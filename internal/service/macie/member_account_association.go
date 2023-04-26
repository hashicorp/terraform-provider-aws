package macie

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_macie_member_account_association")
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

	memberAccountID := d.Get("member_account_id").(string)
	input := &macie.AssociateMemberAccountInput{
		MemberAccountId: aws.String(memberAccountID),
	}

	_, err := conn.AssociateMemberAccountWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie Classic Member Account Association (%s): %s", memberAccountID, err)
	}

	d.SetId(memberAccountID)

	return append(diags, resourceMemberAccountAssociationRead(ctx, d, meta)...)
}

func resourceMemberAccountAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	memberAccount, err := FindMemberAccountByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Macie Classic Member Account Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return diag.Errorf("reading Macie Classic Member Account Association (%s): %s", d.Id(), err)
	}

	d.Set("member_account_id", memberAccount.AccountId)

	return diags
}

func resourceMemberAccountAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	log.Printf("[DEBUG] Deleting Macie Classic Member Account Association: %s", d.Id())
	_, err := conn.DisassociateMemberAccountWithContext(ctx, &macie.DisassociateMemberAccountInput{
		MemberAccountId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, macie.ErrCodeInvalidInputException, "is a master Macie account and cannot be disassociated") {
		return diags
	}

	if tfawserr.ErrMessageContains(err, macie.ErrCodeInvalidInputException, "is not yet associated with Macie") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Macie Classic Member Account Association (%s): %s", d.Id(), err)
	}

	return diags
}

func FindMemberAccountByID(ctx context.Context, conn *macie.Macie, id string) (*macie.MemberAccount, error) {
	input := &macie.ListMemberAccountsInput{}
	var output *macie.MemberAccount

	err := conn.ListMemberAccountsPagesWithContext(ctx, input, func(page *macie.ListMemberAccountsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.MemberAccounts {
			if v != nil && aws.StringValue(v.AccountId) == id {
				output = v
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}
