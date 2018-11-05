package aws

import (
	"testing"
)

func TestDataSyncParseLocationURI(t *testing.T) {
	testCases := []struct {
		LocationURI  string
		LocationType string
		GlobalID     string
		Subdirectory string
	}{
		{
			LocationURI:  "efs://us-east-2.fs-abcd1234/",
			LocationType: "efs",
			GlobalID:     "us-east-2.fs-abcd1234",
			Subdirectory: "/",
		},
		{
			LocationURI:  "efs://us-east-2.fs-abcd1234/path",
			LocationType: "efs",
			GlobalID:     "us-east-2.fs-abcd1234",
			Subdirectory: "/path",
		},
		{
			LocationURI:  "nfs://example.com/",
			LocationType: "nfs",
			GlobalID:     "example.com",
			Subdirectory: "/",
		},
		{
			LocationURI:  "nfs://example.com/path",
			LocationType: "nfs",
			GlobalID:     "example.com",
			Subdirectory: "/path",
		},
		{
			LocationURI:  "s3://myBucket/",
			LocationType: "s3",
			GlobalID:     "myBucket",
			Subdirectory: "/",
		},
		{
			LocationURI:  "s3://myBucket/path",
			LocationType: "s3",
			GlobalID:     "myBucket",
			Subdirectory: "/path",
		},
	}

	for i, tc := range testCases {
		locationType, globalID, subdirectory, err := dataSyncParseLocationURI(tc.LocationURI)
		if err != nil {
			t.Fatalf("%d: received error parsing (%s): %s", i, tc.LocationURI, err)
		}
		if locationType != tc.LocationType {
			t.Fatalf("%d: expected type (%s), received: %s", i, tc.LocationType, locationType)
		}
		if globalID != tc.GlobalID {
			t.Fatalf("%d: expected global ID (%s), received: %s", i, tc.GlobalID, globalID)
		}
		if subdirectory != tc.Subdirectory {
			t.Fatalf("%d: expected subdirectory (%s), received: %s", i, tc.Subdirectory, subdirectory)
		}
	}
}
