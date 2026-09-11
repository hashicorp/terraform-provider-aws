// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestSuppressManagedCertCustomizations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// API response after a managed certificate is issued: a single customizations
	// element carrying only a certificate.
	flattenedWithCert := func() fwtypes.ListNestedObjectValueOf[customizationsModel] {
		return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &customizationsModel{
			Certificate: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &certificateModel{
				ARN: fwtypes.ARNValue("arn:aws:acm:us-east-1:123456789012:certificate/managed"), //lintignore:AWSAT003,AWSAT005
			}),
			GeoRestriction: fwtypes.NewListNestedObjectValueOfNull[geoRestrictionCustomizationModel](ctx),
			WebAcl:         fwtypes.NewListNestedObjectValueOfNull[webAclCustomizationModel](ctx),
		})
	}

	// API response with a configured geo_restriction plus an injected certificate;
	// the geo_restriction.locations set is computed by the API.
	flattenedWithCertAndGeo := func() fwtypes.ListNestedObjectValueOf[customizationsModel] {
		return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &customizationsModel{
			Certificate: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &certificateModel{
				ARN: fwtypes.ARNValue("arn:aws:acm:us-east-1:123456789012:certificate/managed"), //lintignore:AWSAT003,AWSAT005
			}),
			GeoRestriction: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &geoRestrictionCustomizationModel{
				Locations:       fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"US", "CA"}),
				RestrictionType: fwtypes.StringEnumValue(awstypes.GeoRestrictionTypeWhitelist),
			}),
			WebAcl: fwtypes.NewListNestedObjectValueOfNull[webAclCustomizationModel](ctx),
		})
	}

	// Planned: a customizations block with a geo_restriction but no certificate.
	configuredGeoNoCert := func() fwtypes.ListNestedObjectValueOf[customizationsModel] {
		return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &customizationsModel{
			Certificate: fwtypes.NewListNestedObjectValueOfNull[certificateModel](ctx),
			GeoRestriction: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &geoRestrictionCustomizationModel{
				Locations:       fwtypes.NewSetValueOfNull[basetypes.StringValue](ctx),
				RestrictionType: fwtypes.StringEnumValue(awstypes.GeoRestrictionTypeWhitelist),
			}),
			WebAcl: fwtypes.NewListNestedObjectValueOfNull[webAclCustomizationModel](ctx),
		})
	}

	// Planned: an explicitly configured certificate.
	configuredWithCert := func() fwtypes.ListNestedObjectValueOf[customizationsModel] {
		return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &customizationsModel{
			Certificate: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &certificateModel{
				ARN: fwtypes.ARNValue("arn:aws:acm:us-east-1:123456789012:certificate/user"), //lintignore:AWSAT003,AWSAT005
			}),
			GeoRestriction: fwtypes.NewListNestedObjectValueOfNull[geoRestrictionCustomizationModel](ctx),
			WebAcl:         fwtypes.NewListNestedObjectValueOfNull[webAclCustomizationModel](ctx),
		})
	}

	// Planned: a non-null, zero-element list ("customizations = []").
	emptyCustomizations := func() fwtypes.ListNestedObjectValueOf[customizationsModel] {
		return fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []customizationsModel{})
	}

	testCases := map[string]struct {
		planned   fwtypes.ListNestedObjectValueOf[customizationsModel]
		flattened fwtypes.ListNestedObjectValueOf[customizationsModel]
		assert    func(t *testing.T, got fwtypes.ListNestedObjectValueOf[customizationsModel])
	}{
		// No customizations in config, API injected a cert: must collapse back to null.
		"planned null, API injected certificate": {
			planned:   fwtypes.NewListNestedObjectValueOfNull[customizationsModel](ctx),
			flattened: flattenedWithCert(),
			assert: func(t *testing.T, got fwtypes.ListNestedObjectValueOf[customizationsModel]) {
				if !got.IsNull() {
					t.Fatalf("expected null customizations, got %#v", got)
				}
			},
		},
		// Unknown plan: returning unknown post-apply would itself be an inconsistency.
		"planned unknown": {
			planned:   fwtypes.NewListNestedObjectValueOfUnknown[customizationsModel](ctx),
			flattened: flattenedWithCert(),
			assert: func(t *testing.T, got fwtypes.ListNestedObjectValueOf[customizationsModel]) {
				if !got.Equal(flattenedWithCert()) {
					t.Fatalf("expected flattened value to be returned unchanged, got %#v", got)
				}
			},
		},
		// geo_restriction configured, cert injected: only the cert is nulled; the
		// API-computed geo locations survive.
		"planned geo without certificate, API injected certificate": {
			planned:   configuredGeoNoCert(),
			flattened: flattenedWithCertAndGeo(),
			assert: func(t *testing.T, got fwtypes.ListNestedObjectValueOf[customizationsModel]) {
				elems, diags := got.ToSlice(ctx)
				if diags.HasError() {
					t.Fatalf("unexpected diags: %v", diags)
				}
				if len(elems) != 1 {
					t.Fatalf("expected 1 customizations element, got %d", len(elems))
				}
				if !elems[0].Certificate.IsNull() {
					t.Errorf("expected certificate to be nulled, got %#v", elems[0].Certificate)
				}
				geo, d := elems[0].GeoRestriction.ToPtr(ctx)
				if d.HasError() {
					t.Fatalf("unexpected diags reading geo_restriction: %v", d)
				}
				if geo == nil {
					t.Fatal("expected API-computed geo_restriction to be preserved, got null")
				}
				wantLocations := fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"US", "CA"})
				if !geo.Locations.Equal(wantLocations) {
					t.Errorf("expected API-computed locations to be preserved, got %#v", geo.Locations)
				}
			},
		},
		// Explicitly configured certificate: returned unchanged.
		"planned explicit certificate": {
			planned:   configuredWithCert(),
			flattened: flattenedWithCert(),
			assert: func(t *testing.T, got fwtypes.ListNestedObjectValueOf[customizationsModel]) {
				if !got.Equal(flattenedWithCert()) {
					t.Fatalf("expected flattened value to be returned unchanged, got %#v", got)
				}
			},
		},
		// Non-null empty list: no element to inspect, returned as-is.
		"planned non-null empty list": {
			planned:   emptyCustomizations(),
			flattened: flattenedWithCert(),
			assert: func(t *testing.T, got fwtypes.ListNestedObjectValueOf[customizationsModel]) {
				if !got.Equal(emptyCustomizations()) {
					t.Fatalf("expected planned empty list to be returned unchanged, got %#v", got)
				}
			},
		},
		// Configured customizations but flattened came back null: fall back to planned.
		"planned geo without certificate, flattened null": {
			planned:   configuredGeoNoCert(),
			flattened: fwtypes.NewListNestedObjectValueOfNull[customizationsModel](ctx),
			assert: func(t *testing.T, got fwtypes.ListNestedObjectValueOf[customizationsModel]) {
				if !got.Equal(configuredGeoNoCert()) {
					t.Fatalf("expected planned value to be returned when flattened is null, got %#v", got)
				}
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := suppressManagedCertCustomizations(ctx, tc.planned, tc.flattened)
			tc.assert(t, got)
		})
	}
}
