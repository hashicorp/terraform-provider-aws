// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func Diagnostic(in *tfprotov5.Diagnostic) *tfplugin5.Diagnostic {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.Diagnostic{
		Attribute:        AttributePath(in.Attribute),
		Detail:           ForceValidUTF8(in.Detail),
		FunctionArgument: in.FunctionArgument,
		Severity:         Diagnostic_Severity(in.Severity),
		Summary:          ForceValidUTF8(in.Summary),
	}

	return resp
}

func Diagnostic_Severity(in tfprotov5.DiagnosticSeverity) tfplugin5.Diagnostic_Severity {
	return tfplugin5.Diagnostic_Severity(in)
}

func Diagnostics(in []*tfprotov5.Diagnostic) []*tfplugin5.Diagnostic {
	resp := make([]*tfplugin5.Diagnostic, 0, len(in))

	for _, diag := range in {
		resp = append(resp, Diagnostic(diag))
	}

	return resp
}

// ForceValidUTF8 returns a string guaranteed to be valid UTF-8 even if the
// input isn't, by replacing any invalid bytes with a valid UTF-8 encoding of
// the Unicode Replacement Character (\uFFFD).
//
// The protobuf serialization library will reject invalid UTF-8 with an
// unhelpful error message:
//
//	string field contains invalid UTF-8
//
// Passing a string result through this function makes invalid UTF-8 instead
// emerge as placeholder characters on the other side of the wire protocol,
// giving a better chance of still returning a partially-legible message
// instead of a generic character encoding error.
//
// This is intended for user-facing messages such as diagnostic summary and
// detail messages, where Terraform will just treat the value as opaque and
// it's ultimately up to the user and their terminal or web browser to
// interpret the result. Don't use this for strings that have machine-readable
// meaning.
func ForceValidUTF8(s string) string {
	// Most strings that pass through here will already be valid UTF-8 and
	// utf8.ValidString has a fast path which will beat our rune-by-rune
	// analysis below, so it's worth the cost of walking the string twice
	// in the rarer invalid case.
	if utf8.ValidString(s) {
		return s
	}

	// If we get down here then we know there's at least one invalid UTF-8
	// sequence in the string, so in this slow path we'll reconstruct the
	// string one rune at a time, guaranteeing that we'll only write valid
	// UTF-8 sequences into the resulting buffer.
	//
	// Any invalid string will grow at least a little larger as a result of
	// this operation because we'll be replacing each invalid byte with
	// the three-byte sequence \xEF\xBF\xBD, which is the UTF-8 encoding of
	// the replacement character \uFFFD. 9 is a magic number giving room for
	// three such expansions without any further allocation.
	ret := make([]byte, 0, len(s)+9)
	for {
		// If the first byte in s is not the start of a valid UTF-8 sequence
		// then the following will return utf8.RuneError, 1, where
		// utf8.RuneError is the unicode replacement character.
		r, advance := utf8.DecodeRuneInString(s)
		if advance == 0 {
			break
		}
		s = s[advance:]
		ret = utf8.AppendRune(ret, r)
	}
	return string(ret)
}
