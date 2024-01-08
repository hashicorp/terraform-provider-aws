// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package v2

import (
	_ "embed"
)

// Package v2 contains the template bodies for tag methods
// using the AWS SDK Go V2.

//go:embed header_body.tmpl
var HeaderBody string

//go:embed get_tag_body.tmpl
var GetTagBody string

//go:embed list_tags_body.tmpl
var ListTagsBody string

//go:embed service_tags_map_body.tmpl
var ServiceTagsMapBody string

//go:embed service_tags_value_map_body.tmpl
var ServiceTagsValueMapBody string

//go:embed service_tags_slice_body.tmpl
var ServiceTagsSliceBody string

//go:embed update_tags_body.tmpl
var UpdateTagsBody string

//go:embed wait_tags_propagated_body.tmpl
var WaitTagsPropagatedBody string
