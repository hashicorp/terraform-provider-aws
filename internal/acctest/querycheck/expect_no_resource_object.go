// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package querycheck

import (
	"context"
	"fmt"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
)

var _ querycheck.QueryResultCheck = expectNoResourceObject{}

type expectNoResourceObject struct {
	listResourceAddress string
	filter              queryfilter.QueryFilter
}

func (e expectNoResourceObject) QueryFilters(ctx context.Context) []queryfilter.QueryFilter {
	if e.filter == nil {
		return []queryfilter.QueryFilter{}
	}

	return []queryfilter.QueryFilter{
		e.filter,
	}
}

// Adapted (with updates) from github.com/hashicorp/terraform-plugin-testing/statecheck.ExpectIdentity
func (e expectNoResourceObject) CheckQuery(_ context.Context, req querycheck.CheckQueryRequest, resp *querycheck.CheckQueryResponse) {
	var listRes []tfjson.ListResourceFoundData
	for _, res := range req.Query {
		if e.listResourceAddress == strings.TrimPrefix(res.Address, "list.") {
			listRes = append(listRes, res)
		}
	}

	if len(listRes) == 0 {
		resp.Error = fmt.Errorf("%s - no query results found after filtering", e.listResourceAddress)
		return
	}

	if len(listRes) > 1 {
		resp.Error = fmt.Errorf("%s - more than 1 query result found after filtering", e.listResourceAddress)
		return
	}

	res := listRes[0]

	if res.ResourceObject != nil {
		resp.Error = fmt.Errorf("%s - Expected no resource data, but was present!", e.listResourceAddress)
		return
	}
}

// ExpectNoResourceObjet returns a query check that asserts that the given list resource contains no resource data.
//
// This query check can only be used with managed resources that support resource identity and query. Query is only supported in Terraform v1.14+
func ExpectNoResourceObject(resourceAddress string, filter queryfilter.QueryFilter) querycheck.QueryResultCheck {
	return expectNoResourceObject{
		listResourceAddress: resourceAddress,
		filter:              filter,
	}
}
