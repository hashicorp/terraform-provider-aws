// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package docs

import (
	"path"
	"strings"
)

// docKind is a kind of Terraform type whose documentation lists IAM
// permissions. To support another kind, e.g. list resources
// ("list-resources"), ephemeral resources ("ephemeral-resources"), or actions
// ("actions"):
//
//  1. Add it to model.Map once sdkmap reports it.
//  2. Define its docKind below.
//  3. Add a function returning its docPermissions in permissions.go.
//  4. Add its docs in allDocs.
type docKind struct {
	dir   string // directory under website/docs
	intro string // first sentence of the IAM Permissions section
}

var (
	resourceKind = docKind{
		dir:   "r",
		intro: "The following permissions are required to manage this resource.",
	}
	dataSourceKind = docKind{
		dir:   "d",
		intro: "The following permissions are required to read this data source.",
	}
)

// newDoc returns the documentation of typeName, a type of kind k, with p as
// its IAM permissions.
func (k docKind) newDoc(typeName, service string, p docPermissions) doc {
	return doc{
		path:    k.docPath(typeName),
		service: service,
		section: renderSection(k.intro, p),
	}
}

// docPath returns the documentation path of typeName, e.g. for a resource,
// "aws_s3_bucket" is "website/docs/r/s3_bucket.html.markdown".
func (k docKind) docPath(typeName string) string {
	return path.Join("website/docs", k.dir, strings.TrimPrefix(typeName, "aws_")+".html.markdown")
}
