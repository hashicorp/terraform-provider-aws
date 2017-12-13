package cidr

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"testing"
)

func TestSubnet(t *testing.T) {
	type Case struct {
		Base   string
		Bits   int
		Num    int
		Output string
		Error  bool
	}

	cases := []Case{
		Case{
			Base:   "192.168.2.0/20",
			Bits:   4,
			Num:    6,
			Output: "192.168.6.0/24",
		},
		Case{
			Base:   "192.168.2.0/20",
			Bits:   4,
			Num:    0,
			Output: "192.168.0.0/24",
		},
		Case{
			Base:   "192.168.0.0/31",
			Bits:   1,
			Num:    1,
			Output: "192.168.0.1/32",
		},
		Case{
			Base:   "192.168.0.0/21",
			Bits:   4,
			Num:    7,
			Output: "192.168.3.128/25",
		},
		Case{
			Base:   "fe80::/48",
			Bits:   16,
			Num:    6,
			Output: "fe80:0:0:6::/64",
		},
		Case{
			Base:   "fe80::/49",
			Bits:   16,
			Num:    7,
			Output: "fe80:0:0:3:8000::/65",
		},
		Case{
			Base:  "192.168.2.0/31",
			Bits:  2,
			Num:   0,
			Error: true, // not enough bits to expand into
		},
		Case{
			Base:  "fe80::/126",
			Bits:  4,
			Num:   0,
			Error: true, // not enough bits to expand into
		},
		Case{
			Base:  "192.168.2.0/24",
			Bits:  4,
			Num:   16,
			Error: true, // can't fit 16 into 4 bits
		},
	}

	for _, testCase := range cases {
		_, base, _ := net.ParseCIDR(testCase.Base)
		gotNet, err := Subnet(base, testCase.Bits, testCase.Num)
		desc := fmt.Sprintf("Subnet(%#v,%#v,%#v)", testCase.Base, testCase.Bits, testCase.Num)
		if err != nil {
			if !testCase.Error {
				t.Errorf("%s failed: %s", desc, err.Error())
			}
		} else {
			got := gotNet.String()
			if testCase.Error {
				t.Errorf("%s = %s; want error", desc, got)
			} else {
				if got != testCase.Output {
					t.Errorf("%s = %s; want %s", desc, got, testCase.Output)
				}
			}
		}
	}
}

func TestHost(t *testing.T) {
	type Case struct {
		Range  string
		Num    int
		Output string
		Error  bool
	}

	cases := []Case{
		Case{
			Range:  "192.168.2.0/20",
			Num:    6,
			Output: "192.168.0.6",
		},
		Case{
			Range:  "192.168.0.0/20",
			Num:    257,
			Output: "192.168.1.1",
		},
		Case{
			Range:  "2001:db8::/32",
			Num:    1,
			Output: "2001:db8::1",
		},
		Case{
			Range: "192.168.1.0/24",
			Num:   256,
			Error: true, // only 0-255 will fit in 8 bits
		},
		Case{
			Range:  "192.168.0.0/30",
			Num:    -3,
			Output: "192.168.0.1", // 4 address (0-3) in 2 bits; 3rd from end = 1
		},
		Case{
			Range:  "192.168.0.0/30",
			Num:    -4,
			Output: "192.168.0.0", // 4 address (0-3) in 2 bits; 4th from end = 0
		},
		Case{
			Range: "192.168.0.0/30",
			Num:   -5,
			Error: true, // 4 address (0-3) in 2 bits; cannot accomodate 5
		},
	}

	for _, testCase := range cases {
		_, network, _ := net.ParseCIDR(testCase.Range)
		gotIP, err := Host(network, testCase.Num)
		desc := fmt.Sprintf("Host(%#v,%#v)", testCase.Range, testCase.Num)
		if err != nil {
			if !testCase.Error {
				t.Errorf("%s failed: %s", desc, err.Error())
			}
		} else {
			got := gotIP.String()
			if testCase.Error {
				t.Errorf("%s = %s; want error", desc, got)
			} else {
				if got != testCase.Output {
					t.Errorf("%s = %s; want %s", desc, got, testCase.Output)
				}
			}
		}
	}
}

func TestAddressRange(t *testing.T) {
	type Case struct {
		Range string
		First string
		Last  string
	}

	cases := []Case{
		Case{
			Range: "192.168.0.0/16",
			First: "192.168.0.0",
			Last:  "192.168.255.255",
		},
		Case{
			Range: "192.168.0.0/17",
			First: "192.168.0.0",
			Last:  "192.168.127.255",
		},
		Case{
			Range: "fe80::/64",
			First: "fe80::",
			Last:  "fe80::ffff:ffff:ffff:ffff",
		},
	}

	for _, testCase := range cases {
		_, network, _ := net.ParseCIDR(testCase.Range)
		firstIP, lastIP := AddressRange(network)
		desc := fmt.Sprintf("AddressRange(%#v)", testCase.Range)
		gotFirstIP := firstIP.String()
		gotLastIP := lastIP.String()
		if gotFirstIP != testCase.First {
			t.Errorf("%s first is %s; want %s", desc, gotFirstIP, testCase.First)
		}
		if gotLastIP != testCase.Last {
			t.Errorf("%s last is %s; want %s", desc, gotLastIP, testCase.Last)
		}
	}

}

func TestAddressCount(t *testing.T) {
	type Case struct {
		Range string
		Count uint64
	}

	cases := []Case{
		Case{
			Range: "192.168.0.0/16",
			Count: 65536,
		},
		Case{
			Range: "192.168.0.0/17",
			Count: 32768,
		},
		Case{
			Range: "192.168.0.0/32",
			Count: 1,
		},
		Case{
			Range: "192.168.0.0/31",
			Count: 2,
		},
		Case{
			Range: "0.0.0.0/0",
			Count: 4294967296,
		},
		Case{
			Range: "0.0.0.0/1",
			Count: 2147483648,
		},
		Case{
			Range: "::/65",
			Count: 9223372036854775808,
		},
		Case{
			Range: "::/128",
			Count: 1,
		},
		Case{
			Range: "::/127",
			Count: 2,
		},
	}

	for _, testCase := range cases {
		_, network, _ := net.ParseCIDR(testCase.Range)
		gotCount := AddressCount(network)
		desc := fmt.Sprintf("AddressCount(%#v)", testCase.Range)
		if gotCount != testCase.Count {
			t.Errorf("%s = %d; want %d", desc, gotCount, testCase.Count)
		}
	}

}

func TestIncDec(t *testing.T) {

	testCase := [][]string{
		[]string{"0.0.0.0", "0.0.0.1"},
		[]string{"10.0.0.0", "10.0.0.1"},
		[]string{"9.255.255.255", "10.0.0.0"},
		[]string{"255.255.255.255", "0.0.0.0"},
		[]string{"::", "::1"},
		[]string{"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", "::"},
		[]string{"2001:db8:c001:ba00::", "2001:db8:c001:ba00::1"},
	}

	for _, tc := range testCase {
		ip1 := net.ParseIP(tc[0])
		ip2 := net.ParseIP(tc[1])
		iIP := Inc(ip1)
		if !iIP.Equal(ip2) {
			t.Logf("%s should inc to equal %s\n", tc[0], tc[1])
			t.Errorf("%v should equal %v\n", iIP, ip2)
		}
		if ip1.Equal(ip2) {
			t.Errorf("[%v] should not have been modified to [%v]", ip2, iIP)
		}
	}
	for _, tc := range testCase {
		ip1 := net.ParseIP(tc[0])
		ip2 := net.ParseIP(tc[1])
		dIP := Dec(ip2)
		if !ip1.Equal(dIP) {
			t.Logf("%s should dec equal %s\n", tc[0], tc[1])
			t.Errorf("%v should equal %v\n", ip1, dIP)
		}
		if ip2.Equal(dIP) {
			t.Errorf("[%v] should not have been modified to [%v]", ip2, dIP)
		}
	}
}

func TestPreviousSubnet(t *testing.T) {

	testCases := [][]string{
		[]string{"10.0.0.0/24", "9.255.255.0/24", "false"},
		[]string{"100.0.0.0/26", "99.255.255.192/26", "false"},
		[]string{"0.0.0.0/26", "255.255.255.192/26", "true"},
		[]string{"2001:db8:e000::/36", "2001:db8:d000::/36", "false"},
		[]string{"::/64", "ffff:ffff:ffff:ffff::/64", "true"},
	}
	for _, tc := range testCases {
		_, c1, _ := net.ParseCIDR(tc[0])
		_, c2, _ := net.ParseCIDR(tc[1])
		mask, _ := c1.Mask.Size()
		p1, rollback := PreviousSubnet(c1, mask)
		if !p1.IP.Equal(c2.IP) {
			t.Errorf("IP expected %v, got %v\n", c2.IP, p1.IP)
		}
		if !bytes.Equal(p1.Mask, c2.Mask) {
			t.Errorf("Mask expected %v, got %v\n", c2.Mask, p1.Mask)
		}
		if p1.String() != c2.String() {
			t.Errorf("%s should have been equal %s\n", p1.String(), c2.String())
		}
		if check, _ := strconv.ParseBool(tc[2]); rollback != check {
			t.Errorf("%s to %s  should have rolled\n", tc[0], tc[1])
		}
	}
	for _, tc := range testCases {
		_, c1, _ := net.ParseCIDR(tc[0])
		_, c2, _ := net.ParseCIDR(tc[1])
		mask, _ := c1.Mask.Size()
		n1, rollover := NextSubnet(c2, mask)
		if !n1.IP.Equal(c1.IP) {
			t.Errorf("IP expected %v, got %v\n", c1.IP, n1.IP)
		}
		if !bytes.Equal(n1.Mask, c1.Mask) {
			t.Errorf("Mask expected %v, got %v\n", c1.Mask, n1.Mask)
		}
		if n1.String() != c1.String() {
			t.Errorf("%s should have been equal %s\n", n1.String(), c1.String())
		}
		if check, _ := strconv.ParseBool(tc[2]); rollover != check {
			t.Errorf("%s to %s  should have rolled\n", tc[0], tc[1])
		}
	}
}

func TestVerifyNetowrk(t *testing.T) {

	type testVerifyNetwork struct {
		CIDRBlock string
		CIDRList  []string
	}

	testCases := []*testVerifyNetwork{
		&testVerifyNetwork{
			CIDRBlock: "192.168.8.0/21",
			CIDRList: []string{
				"192.168.8.0/24",
				"192.168.9.0/24",
				"192.168.10.0/24",
				"192.168.11.0/25",
				"192.168.11.128/25",
				"192.168.12.0/25",
				"192.168.12.128/26",
				"192.168.12.192/26",
				"192.168.13.0/26",
				"192.168.13.64/27",
				"192.168.13.96/27",
				"192.168.13.128/27",
			},
		},
	}
	failCases := []*testVerifyNetwork{
		&testVerifyNetwork{
			CIDRBlock: "192.168.8.0/21",
			CIDRList: []string{
				"192.168.8.0/24",
				"192.168.9.0/24",
				"192.168.10.0/24",
				"192.168.11.0/25",
				"192.168.11.128/25",
				"192.168.12.0/25",
				"192.168.12.64/26",
				"192.168.12.128/26",
			},
		},
		&testVerifyNetwork{
			CIDRBlock: "192.168.8.0/21",
			CIDRList: []string{
				"192.168.7.0/24",
				"192.168.9.0/24",
				"192.168.10.0/24",
				"192.168.11.0/25",
				"192.168.11.128/25",
				"192.168.12.0/25",
				"192.168.12.64/26",
				"192.168.12.128/26",
			},
		},
	}

	for _, tc := range testCases {
		subnets := make([]*net.IPNet, len(tc.CIDRList))
		for i, s := range tc.CIDRList {
			_, n, err := net.ParseCIDR(s)
			if err != nil {
				t.Errorf("Bad test data %s\n", s)
			}
			subnets[i] = n
		}
		_, CIDRBlock, perr := net.ParseCIDR(tc.CIDRBlock)
		if perr != nil {
			t.Errorf("Bad test data %s\n", tc.CIDRBlock)
		}
		test := VerifyNoOverlap(subnets, CIDRBlock)
		if test != nil {
			t.Errorf("Failed test with %v\n", test)
		}
	}
	for _, tc := range failCases {
		subnets := make([]*net.IPNet, len(tc.CIDRList))
		for i, s := range tc.CIDRList {
			_, n, err := net.ParseCIDR(s)
			if err != nil {
				t.Errorf("Bad test data %s\n", s)
			}
			subnets[i] = n
		}
		_, CIDRBlock, perr := net.ParseCIDR(tc.CIDRBlock)
		if perr != nil {
			t.Errorf("Bad test data %s\n", tc.CIDRBlock)
		}
		test := VerifyNoOverlap(subnets, CIDRBlock)
		if test == nil {
			t.Errorf("Test should have failed with CIDR %s\n", tc.CIDRBlock)
		}
	}
}
