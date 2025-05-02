// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/memorydb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/memorydb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_memorydb_acl", name="ACL")
// @Tags(identifierAttribute="arn")
func resourceACL() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceACLCreate,
		ReadWithoutTimeout:   resourceACLRead,
		UpdateWithoutTimeout: resourceACLUpdate,
		DeleteWithoutTimeout: resourceACLDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"minimum_engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validateResourceName(aclNameMaxLength),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validateResourceNamePrefix(aclNameMaxLength - id.UniqueIDSuffixLength),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"user_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, userNameMaxLength),
				},
			},
		},
	}
}

func resourceACLCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &memorydb.CreateACLInput{
		ACLName: aws.String(name),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("user_names"); ok && v.(*schema.Set).Len() > 0 {
		input.UserNames = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	_, err := conn.CreateACL(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MemoryDB ACL (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitACLActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB ACL (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceACLRead(ctx, d, meta)...)
}

func resourceACLRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	acl, err := findACLByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB ACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MemoryDB ACL (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, acl.ARN)
	d.Set("minimum_engine_version", acl.MinimumEngineVersion)
	d.Set(names.AttrName, acl.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(acl.Name)))
	d.Set("user_names", acl.UserNames)

	return diags
}

func resourceACLUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &memorydb.UpdateACLInput{
			ACLName: aws.String(d.Id()),
		}

		o, n := d.GetChange("user_names")
		os, ns := o.(*schema.Set), n.(*schema.Set)

		if toAdd := ns.Difference(os); toAdd.Len() > 0 {
			input.UserNamesToAdd = flex.ExpandStringValueSet(toAdd)
		}

		// When a user is deleted, MemoryDB will implicitly remove it from any
		// ACL-s it was associated with.
		//
		// Attempting to remove a user that isn't in the ACL will fail with
		// InvalidParameterValueException. To work around this, filter out any
		// users that have been reported as no longer being in the group.

		initialState, err := findACLByName(ctx, conn, d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading MemoryDB ACL (%s): %s", d.Id(), err)
		}

		initialUserNames := map[string]struct{}{}
		for _, userName := range initialState.UserNames {
			initialUserNames[userName] = struct{}{}
		}

		for _, v := range os.Difference(ns).List() {
			userNameToRemove := v.(string)
			_, userNameStillPresent := initialUserNames[userNameToRemove]

			if userNameStillPresent {
				input.UserNamesToRemove = append(input.UserNamesToRemove, userNameToRemove)
			}
		}

		if len(input.UserNamesToAdd) > 0 || len(input.UserNamesToRemove) > 0 {
			_, err := conn.UpdateACL(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating MemoryDB ACL (%s): %s", d.Id(), err)
			}

			if _, err := waitACLActive(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB ACL (%s) update: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceACLRead(ctx, d, meta)...)
}

func resourceACLDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	log.Printf("[DEBUG] Deleting MemoryDB ACL: (%s)", d.Id())
	_, err := conn.DeleteACL(ctx, &memorydb.DeleteACLInput{
		ACLName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ACLNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MemoryDB ACL (%s): %s", d.Id(), err)
	}

	if _, err := waitACLDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB ACL (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findACLByName(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.ACL, error) {
	input := &memorydb.DescribeACLsInput{
		ACLName: aws.String(name),
	}

	return findACL(ctx, conn, input)
}

func findACL(ctx context.Context, conn *memorydb.Client, input *memorydb.DescribeACLsInput) (*awstypes.ACL, error) {
	output, err := findACLs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findACLs(ctx context.Context, conn *memorydb.Client, input *memorydb.DescribeACLsInput) ([]awstypes.ACL, error) {
	var output []awstypes.ACL

	pages := memorydb.NewDescribeACLsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ACLNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ACLs...)
	}

	return output, nil
}

func statusACL(ctx context.Context, conn *memorydb.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findACLByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitACLActive(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.ACL, error) { //nolint:unparam
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{aclStatusCreating, aclStatusModifying},
		Target:  []string{aclStatusActive},
		Refresh: statusACL(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ACL); ok {
		return output, err
	}

	return nil, err
}

func waitACLDeleted(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.ACL, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:      []string{aclStatusDeleting},
		Target:       []string{},
		Refresh:      statusACL(ctx, conn, name),
		Timeout:      timeout,
		Delay:        30 * time.Second,
		PollInterval: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ACL); ok {
		return output, err
	}

	return nil, err
}
