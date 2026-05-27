// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tags/main.go -ServiceTagsSlice -ListTags -ListTagsInIDElem=ReportName -UpdateTags -TagInIDElem=ReportName -KVTValues
// ONLY generate directives and package declaration! Do not add anything else to this file.

package cur
