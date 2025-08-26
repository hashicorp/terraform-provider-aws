// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	awstypes "github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_location_tracker_association", name="Tracker Association")
func ResourceTrackerAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrackerAssociationCreate,
		ReadWithoutTimeout:   resourceTrackerAssociationRead,
		DeleteWithoutTimeout: resourceTrackerAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"consumer_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"tracker_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
		},
	}
}

const (
	ResNameTrackerAssociation = "Tracker Association"
)

func resourceTrackerAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	consumerArn := d.Get("consumer_arn").(string)
	trackerName := d.Get("tracker_name").(string)

	in := &location.AssociateTrackerConsumerInput{
		ConsumerArn: aws.String(consumerArn),
		TrackerName: aws.String(trackerName),
	}

	out, err := conn.AssociateTrackerConsumer(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionCreating, ResNameTrackerAssociation, d.Get(names.AttrName).(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionCreating, ResNameTrackerAssociation, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	d.SetId(fmt.Sprintf("%s|%s", trackerName, consumerArn))

	return append(diags, resourceTrackerAssociationRead(ctx, d, meta)...)
}

func resourceTrackerAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	trackerAssociationId, err := TrackerAssociationParseID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionReading, ResNameTrackerAssociation, d.Id(), err)
	}

	err = FindTrackerAssociationByTrackerNameAndConsumerARN(ctx, conn, trackerAssociationId.TrackerName, trackerAssociationId.ConsumerARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Location TrackerAssociation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionReading, ResNameTrackerAssociation, d.Id(), err)
	}

	d.Set("consumer_arn", trackerAssociationId.ConsumerARN)
	d.Set("tracker_name", trackerAssociationId.TrackerName)

	return diags
}

func resourceTrackerAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	log.Printf("[INFO] Deleting Location TrackerAssociation %s", d.Id())

	trackerAssociationId, err := TrackerAssociationParseID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionReading, ResNameTrackerAssociation, d.Id(), err)
	}

	_, err = conn.DisassociateTrackerConsumer(ctx, &location.DisassociateTrackerConsumerInput{
		ConsumerArn: aws.String(trackerAssociationId.ConsumerARN),
		TrackerName: aws.String(trackerAssociationId.TrackerName),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionDeleting, ResNameTrackerAssociation, d.Id(), err)
	}

	return diags
}

// FindTrackerAssociationByTrackerNameAndConsumerARN returns an error if an association for specified tracker and consumer cannot be found
func FindTrackerAssociationByTrackerNameAndConsumerARN(ctx context.Context, conn *location.Client, trackerName, consumerARN string) error {
	in := &location.ListTrackerConsumersInput{
		TrackerName: aws.String(trackerName),
	}

	found := false

	pages := location.NewListTrackerConsumersPaginator(conn, in)

	for pages.HasMorePages() && !found {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return err
		}

		if slices.Contains(page.ConsumerArns, consumerARN) {
			found = true
		}
	}

	if !found {
		return &retry.NotFoundError{}
	}

	return nil
}

type TrackerAssociationID struct {
	TrackerName string
	ConsumerARN string
}

func TrackerAssociationParseID(id string) (TrackerAssociationID, error) {
	idParts := strings.Split(id, "|")
	if len(idParts) != 2 {
		return TrackerAssociationID{}, fmt.Errorf("please make sure the ID is in the form TRACKERNAME|CONSUMERARN")
	}

	return TrackerAssociationID{idParts[0], idParts[1]}, nil
}
