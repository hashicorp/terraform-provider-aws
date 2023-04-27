package inspector2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_inspector2_member_association")
func ResourceMemberAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMemberAssociationCreate,
		ReadWithoutTimeout:   resourceMemberAssociationRead,
		DeleteWithoutTimeout: resourceMemberAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},
	}
}

func resourceMemberAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client()

	accountID := d.Get("account_id").(string)
	input := &inspector2.AssociateMemberInput{
		AccountId: aws.String(accountID),
	}

	_, err := conn.AssociateMember(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspector2 Member Association (%s): %s", accountID, err)
	}

	d.SetId(accountID)

	return append(diags, resourceMemberAssociationRead(ctx, d, meta)...)
}

func resourceMemberAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client()

	member, err := FindMemberByAccountID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector2 Member Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector2 Member Association (%s): %s", d.Id(), err)
	}

	d.Set("account_id", member.AccountId)

	return diags
}

func resourceMemberAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client()

	log.Printf("[DEBUG] Deleting Inspector2 Member Association: %s", d.Id())
	_, err := conn.DisassociateMember(ctx, &inspector2.DisassociateMemberInput{
		AccountId: aws.String(d.Get("account_id").(string)),
	})

	// An error occurred (ValidationException) when calling the DisassociateMember operation: The request is rejected because the current account cannot disassociate the given member account ID since the latter is not yet associated to it.
	if errs.IsAErrorMessageContains[*types.ValidationException](err, "is not yet associated to it") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Inspector2 Member Association (%s): %s", d.Id(), err)
	}

	return diags
}

func FindMemberByAccountID(ctx context.Context, conn *inspector2.Client, id string) (*types.Member, error) {
	input := &inspector2.GetMemberInput{
		AccountId: aws.String(id),
	}

	output, err := conn.GetMember(ctx, input)

	if errs.IsA[*types.AccessDeniedException](err) || errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Member == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.Member.RelationshipStatus; status == types.RelationshipStatusRemoved {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output.Member, nil
}
