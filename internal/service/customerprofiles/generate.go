// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ServiceTagsMap -KVTValues -ListTags -UpdateTags
//go:generate go run ../../generate/listpages/main.go -ListOps=ListDomains,SearchProfiles
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package customerprofiles
