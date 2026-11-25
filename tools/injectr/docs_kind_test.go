// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import "testing"

func TestDocKindDocPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		kind     docKind
		typeName string
		want     string
	}{
		"resource":    {kind: resourceKind, typeName: "aws_s3_bucket", want: "website/docs/r/s3_bucket.html.markdown"},
		"data source": {kind: dataSourceKind, typeName: "aws_s3_bucket", want: "website/docs/d/s3_bucket.html.markdown"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tc.kind.docPath(tc.typeName); got != tc.want {
				t.Errorf("docPath = %q, want %q", got, tc.want)
			}
		})
	}
}
