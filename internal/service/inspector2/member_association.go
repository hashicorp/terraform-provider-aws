package inspector2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
				ValidateFunc: verify.ValidAccountID,
				Required:     true,
				ForceNew:     true,
			},
		},
	}
}

func resourceMemberAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client()

	accountId := d.Get("account_id").(string)

	input := &inspector2.AssociateMemberInput{
		AccountId: aws.String(accountId),
	}

	output, err := conn.AssociateMember(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspect2 Member Association: %s", err)
	}

	if err == nil && (output == nil || output.AccountId == nil) {
		return sdkdiag.AppendErrorf(diags, "creating Inspect2 Member Association (%s): empty output", d.Id())
	}

	d.SetId(accountId)

	return resourceMemberAssociationRead(ctx, d, meta)
}

func resourceMemberAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client()

	getMemberInput := &inspector2.GetMemberInput{
		AccountId: aws.String(d.Id()),
	}

	out, err := conn.GetMember(ctx, getMemberInput)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector2 Associated Member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspect2 Member Association: %s", err)
	}

	if err := d.Set("account_id", out.Member.AccountId); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspect2 Member Association (%s): empty output", d.Id())
	}

	return nil
}

func resourceMemberAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client()
	in := &inspector2.DisassociateMemberInput{
		AccountId: aws.String(d.Get("account_id").(string)),
	}

	_, err := conn.DisassociateMember(ctx, in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Inspect2 Member Association (%s): empty output", d.Id())
	}

	return nil
}
