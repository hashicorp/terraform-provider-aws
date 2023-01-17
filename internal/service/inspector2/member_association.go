package inspector2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	// "github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	// "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	// "github.com/hashicorp/terraform-provider-aws/internal/enum"
	// "github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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

const (
	ResNameMemberAssociation = "Member Association"
)

func resourceMemberAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Inspector2Client()

	id := d.Get("account_id").(string)

	input := &inspector2.AssociateMemberInput{
		AccountId: aws.String(id),
	}

	output, err := conn.AssociateMember(ctx, input)

	if err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionCreating, ResNameMemberAssociation, id, err)
	}

	if err == nil && (output == nil || output.AccountId == nil) {
		return create.DiagError(names.Inspector2, create.ErrActionCreating, ResNameMemberAssociation, id, errors.New("empty output"))
	}

	d.SetId(id)

	return resourceMemberAssociationRead(ctx, d, meta)
}

func resourceMemberAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Inspector2Client()

	ai, _, err := FindAssociatedMemberStatus(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector2 Associated Member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionReading, ResNameMemberAssociation, d.Id(), err)
	}

	if err := d.Set("account_id", ai); err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionReading, ResNameMemberAssociation, d.Id(), err)
	}

	return nil
}

func resourceMemberAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Inspector2Client()
	in := &inspector2.DisassociateMemberInput{
		AccountId: aws.String(d.Get("account_id").(string)),
	}

	_, err := conn.DisassociateMember(ctx, in)

	if err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionDeleting, ResNameMemberAssociation, d.Id(), err)
	}

	return nil
}

// // Values returns all known values for AssociateMemberStatus
// func (AssociateMemberStatus) Values() []AssociatedMemberStatus {
// 	return []AssociatedMemberStatus{
// 		AssociatedMemberStatus(types.RelationshipStatusCreated),
// 		AssociatedMemberStatus(types.RelationshipStatusInvited),
// 		AssociatedMemberStatus(types.RelationshipStatusDisabled),
// 		AssociatedMemberStatus(types.RelationshipStatusEnabled),
// 		AssociatedMemberStatus(types.RelationshipStatusRemoved),
// 		AssociatedMemberStatus(types.RelationshipStatusResigned),
// 		AssociatedMemberStatus(types.RelationshipStatusDeleted),
// 		AssociatedMemberStatus(types.RelationshipStatusEmailVerificationInProgress),
// 		AssociatedMemberStatus(types.RelationshipStatusEmailVerificationFailed),
// 		AssociatedMemberStatus(types.RelationshipStatusRegionDisabled),
// 		AssociatedMemberStatus(types.RelationshipStatusAccountSuspended),
// 		AssociatedMemberStatus(types.RelationshipStatusCannotCreateDetectorInOrgMaster),
// 	}
// }

// func waitMemberAssociationEnabled(ctx context.Context, conn *inspector2.Client, id string, timeout time.Duration) error {
// 	stateConf := &resource.StateChangeConf{
// 		Pending: enum.Slice(),
// 		Target: enum.Slice(types.StatusEnabled),
// 		Refresh: statusMemberAssociation(ctx, conn, id),
// 		Timeout: timeout,
// 	}

// 	_, err := stateConf.WaitForStateContext(ctx)

// 	return err
// }

// func waitMemberAssociationDisabled(ctx context.Context, conn *inspector2.Client, id string, timeout time.Duration) error {
// 	stateConf := &resource.StateChangeConf{
// 		Pending: enum.Slice(),
// 		Target: enum.Slice(),
// 		Refresh statusMemberAssociation(types.StatusDisabled),
// 		Timeout: timeout,
// 	}

// 	_, err := stateConf.WaitForStateContext(ctx)

// 	return err
// }

// func statusMemberAssociation(ctx context.Context, conn *inspector2.Client, id string) resource.StateRefreshFunc {
// 	return func() (interface{}, string, error) {
// 		st, _, err := FindAssociatedMemberStatus(ctx, conn, id)

// 		if tfresource.NotFound(err) {
// 			return nil, "", nil
// 		}

// 		if err != nil {
// 			return nil, "", err
// 		}

// 		return "", st, nil
// 	}
// }

func FindAssociatedMemberStatus(ctx context.Context, conn *inspector2.Client, id string) (string, string, error) {
	in := &inspector2.GetMemberInput{
		AccountId: aws.String(id),
	}

	out, err := conn.GetMember(ctx, in)

	if err != nil {
		return "", "", fmt.Errorf("calling GetMember: %s", err)
	}

	if out == nil {
		return "", "", fmt.Errorf("empty response calling GetMember")
	}

	returnedMember := out.Member

	accountIdValue := aws.ToString(returnedMember.AccountId)
	relationshipValue := string(returnedMember.RelationshipStatus)

	return accountIdValue, relationshipValue, err
}
