// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAppSync_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"APIKey": {
			acctest.CtBasic: testAccAPIKey_basic,
			"description":   testAccAPIKey_description,
			"expires":       testAccAPIKey_expires,
		},
		"Function": {
			acctest.CtBasic:           testAccFunction_basic,
			"code":                    testAccFunction_code,
			acctest.CtDisappears:      testAccFunction_disappears,
			"description":             testAccFunction_description,
			"responseMappingTemplate": testAccFunction_responseMappingTemplate,
			"sync":                    testAccFunction_syncConfig,
		},
		"Resolver": {
			acctest.CtBasic:      testAccResolver_basic,
			"code":               testAccResolver_code,
			acctest.CtDisappears: testAccResolver_disappears,
			"dataSource":         testAccResolver_dataSource,
			"DataSource_lambda":  testAccResolver_DataSource_lambda,
			"requestTemplate":    testAccResolver_requestTemplate,
			"responseTemplate":   testAccResolver_responseTemplate,
			"multipleResolvers":  testAccResolver_multipleResolvers,
			"pipeline":           testAccResolver_pipeline,
			"caching":            testAccResolver_caching,
			"sync":               testAccResolver_syncConfig,
		},
		"ApiCache": {
			acctest.CtBasic:      testAccAPICache_basic,
			acctest.CtDisappears: testAccAPICache_disappears,
		},
		"Type": {
			acctest.CtBasic:      testAccType_basic,
			acctest.CtDisappears: testAccType_disappears,
		},
		"DomainName": {
			acctest.CtBasic:      testAccDomainName_basic,
			acctest.CtDisappears: testAccDomainName_disappears,
			"description":        testAccDomainName_description,
		},
		"DomainNameAssociation": {
			acctest.CtBasic:      testAccDomainNameAPIAssociation_basic,
			acctest.CtDisappears: testAccDomainNameAPIAssociation_disappears,
		},
		"SourceApiAssociation": {
			acctest.CtBasic:      testAccAppSyncSourceAPIAssociation_basic,
			acctest.CtDisappears: testAccAppSyncSourceAPIAssociation_disappears,
			"update":             testAccAppSyncSourceAPIAssociation_update,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
