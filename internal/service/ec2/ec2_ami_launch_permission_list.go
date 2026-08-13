// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_ami_launch_permission", name="AMI Launch Permission")
func newAMILaunchPermissionResourceAsListResource() inttypes.ListResourceForSDK {
	l := amiLaunchPermissionListResource{}
	l.SetResourceSchema(resourceAMILaunchPermission())
	return &l
}

var _ list.ListResource = &amiLaunchPermissionListResource{}

type amiLaunchPermissionListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *amiLaunchPermissionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	tflog.Info(ctx, "Listing resources")
	stream.Results = func(yield func(list.ListResult) bool) {
		for perm, err := range listAMILaunchPermissions(ctx, conn) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			imageID := perm.imageID
			accountID := perm.accountID
			group := perm.group
			organizationARN := perm.organizationARN
			organizationalUnitARN := perm.organizationalUnitARN

			id := amiLaunchPermissionCreateResourceID(imageID, accountID, group, organizationARN, organizationalUnitARN)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)

			resourceAMILaunchPermissionFlatten(rd, imageID, accountID, group, organizationARN, organizationalUnitARN)

			if request.IncludeResource { //nolint:revive,staticcheck // Be explicit about IncludeResource handling.
				// No-op, all readable attributes are already populated above.
			}

			result.DisplayName = fmt.Sprintf("%s (%s)", imageID, id)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type amiLaunchPermission struct {
	imageID               string
	accountID             string
	group                 string
	organizationARN       string
	organizationalUnitARN string
}

func listAMILaunchPermissions(ctx context.Context, conn *ec2.Client) iter.Seq2[amiLaunchPermission, error] {
	return func(yield func(amiLaunchPermission, error) bool) {
		input := ec2.DescribeImagesInput{
			Owners: []string{"self"},
		}
		images, err := findImages(ctx, conn, &input)
		if err != nil {
			yield(amiLaunchPermission{}, fmt.Errorf("listing EC2 AMI Launch Permissions: listing AMIs: %w", err))
			return
		}
		for _, image := range images {
			fmt.Println(*image.OwnerId)
			imageID := aws.ToString(image.ImageId)
			perms, err := findImageLaunchPermissionsByID(ctx, conn, imageID)
			if err != nil {
				// Image may have been deregistered between listing and fetching permissions.
				continue
			}

			for _, perm := range perms {
				p := amiLaunchPermission{
					imageID:               imageID,
					accountID:             aws.ToString(perm.UserId),
					group:                 string(perm.Group),
					organizationARN:       aws.ToString(perm.OrganizationArn),
					organizationalUnitARN: aws.ToString(perm.OrganizationalUnitArn),
				}
				if !yield(p, nil) {
					return
				}
			}
		}
	}
}
