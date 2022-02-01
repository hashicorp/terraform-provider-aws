package directconnect

import (
	"testing"
)

func TestValidConnectionBandWidth(t *testing.T) {
	validBandwidths := []string{
		"1Gbps",
		"2Gbps",
		"5Gbps",
		"10Gbps",
		"100Gbps",
		"50Mbps",
		"100Mbps",
		"200Mbps",
		"300Mbps",
		"400Mbps",
		"500Mbps",
	}
	for _, v := range validBandwidths {
		_, errors := validConnectionBandWidth()(v, "bandwidth")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid bandwidth: %q", v, errors)
		}
	}

	invalidBandwidths := []string{
		"1Tbps",
		"10GBpS",
		"42Mbps",
		"0",
		"???",
		"a lot",
	}
	for _, v := range invalidBandwidths {
		_, errors := validConnectionBandWidth()(v, "bandwidth")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid bandwidth", v)
		}
	}
}

func TestValidAmazonSideASN(t *testing.T) {
	validAsns := []string{
		"7224",
		"9059",
		"10124",
		"17493",
		"64512",
		"64513",
		"65533",
		"65534",
		"4200000000",
		"4200000001",
		"4294967293",
		"4294967294",
	}
	for _, v := range validAsns {
		_, errors := validAmazonSideASN(v, "amazon_side_asn")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid ASN: %q", v, errors)
		}
	}

	invalidAsns := []string{
		"1",
		"ABCDEFG",
		"",
		"7225",
		"9058",
		"10125",
		"17492",
		"64511",
		"65535",
		"4199999999",
		"4294967295",
		"9999999999",
	}
	for _, v := range invalidAsns {
		_, errors := validAmazonSideASN(v, "amazon_side_asn")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid ASN", v)
		}
	}
}
