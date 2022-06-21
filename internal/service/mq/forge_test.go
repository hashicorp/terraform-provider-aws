package mq_test

import (
	"testing"

	tfmq "github.com/hashicorp/terraform-provider-aws/internal/service/mq"
)

func TestCanonicalXML(t *testing.T) {
	cases := []struct {
		Name        string
		Config      string
		Expected    string
		ExpectError bool
	}{
		{
			Name:     "Config sample from MSDN",
			Config:   testAccForgeConfig_testExampleXMLFromMsdn,
			Expected: testAccForgeConfig_testExampleXMLFromMsdn,
		},
		{
			Name:     "Config sample from MSDN, modified",
			Config:   testAccForgeConfig_testExampleXMLFromMsdn,
			Expected: testExampleXML_from_msdn_modified,
		},
		{
			Name:        "Config sample from MSDN, flaw",
			Config:      testAccForgeConfig_testExampleXMLFromMsdn,
			Expected:    testExampleXML_from_msdn_flawed,
			ExpectError: true,
		},
		{
			Name: "A note",
			Config: `
<?xml version="1.0"?>
<note>
<to>You</to>
<from>Me</from>
<heading>Reminder</heading>
<body>You're awesome</body>
<rant/>
<rant/>
</note>
`,
			Expected: `
<?xml version="1.0"?>
<note>
	<to>You</to>
	<from>Me</from>
	<heading>
    Reminder
    </heading>
	<body>You're awesome</body>
	<rant/>
	<rant>
</rant>
</note>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			config, err := tfmq.CanonicalXML(tc.Config)
			if err != nil {
				t.Fatalf("Error getting canonical xml for given config: %s", err)
			}
			expected, err := tfmq.CanonicalXML(tc.Expected)
			if err != nil {
				t.Fatalf("Error getting canonical xml for expected config: %s", err)
			}

			if config != expected {
				if !tc.ExpectError {
					t.Fatalf("Error matching canonical xmls:\n\tconfig: %s\n\n\texpected: %s\n", config, expected)
				}
			}
		})
	}
}

const testAccForgeConfig_testExampleXMLFromMsdn = `
<?xml version="1.0"?>
<purchaseOrder xmlns="http://tempuri.org/po.xsd" orderDate="1999-10-20">
    <shipTo country="US">
        <name>Alice Smith</name>
        <street>123 Maple Street</street>
        <city>Mill Valley</city>
        <state>CA</state>
        <zip>90952</zip>
    </shipTo>
    <billTo country="US">
        <name>Robert Smith</name>
        <street>8 Oak Avenue</street>
        <city>Old Town</city>
        <state>PA</state>
        <zip>95819</zip>
    </billTo>
    <comment>Hurry, my lawn is going wild!</comment>
    <items>
        <item partNum="872-AA">
            <productName>Lawnmower</productName>
            <quantity>1</quantity>
            <USPrice>148.95</USPrice>
            <comment>Confirm this is electric</comment>
        </item>
        <item partNum="926-AA">
            <productName>Baby Monitor</productName>
            <quantity>1</quantity>
            <USPrice>39.98</USPrice>
            <shipDate>1999-05-21</shipDate>
        </item>
				<item/>
				<item/>
    </items>
</purchaseOrder>
`

const testExampleXML_from_msdn_modified = `
<?xml version="1.0"?>
<purchaseOrder xmlns="http://tempuri.org/po.xsd" orderDate="1999-10-20">
    <shipTo country="US">
        <name>Alice Smith</name>
        <street>123 Maple Street</street>
        <city>Mill Valley</city>
        <state>CA</state>
        <zip>90952</zip>
    </shipTo>
    <billTo country="US">
        <name>Robert Smith</name>
        <street>8 Oak Avenue</street>
        <city>Old Town</city>
        <state>PA</state>
        <zip>95819</zip>
    </billTo>
    <comment>Hurry, my lawn is going wild!</comment>
    <items>
        <item partNum="872-AA">
            <productName>Lawnmower</productName>
            <quantity>1</quantity>
            <USPrice>148.95</USPrice>
            <comment>Confirm this is electric</comment>
        </item>
        <item partNum="926-AA">
            <productName>Baby Monitor</productName>
            <quantity>1</quantity>
            <USPrice>39.98</USPrice>
            <shipDate>1999-05-21</shipDate>
        </item>
				  	 <item></item>
				<item>
</item>
    </items>
</purchaseOrder>
`

const testExampleXML_from_msdn_flawed = `
<?xml version="1.0"?>
<purchaseOrder xmlns="http://tempuri.org/po.xsd" orderDate="1999-10-20">
    <shipTo country="US">
        <name>Alice Smith</name>
        <street>123 Maple Street</street>
        <city>Mill Valley</city>
        <state>CA</state>
        <zip>90952</zip>
    </shipTo>
    <billTo country="US">
        <name>Robert Smith</name>
        <street>8 Oak Avenue</street>
        <city>Old Town</city>
        <state>PA</state>
        <zip>95819</zip>
    </billTo>
    <comment>Hurry, my lawn is going wild!</comment>
    <items>
        <item partNum="872-AA">
            <productName>Lawnmower</productName>
            <quantity>1</quantity>
            <USPrice>148.95</USPrice>
            <comment>Confirm this is electric</comment>
        </item>
        <item partNum="926-AA">
            <productName>Baby Monitor</productName>
            <quantity>1</quantity>
            <USPrice>39.98</USPrice>
            <shipDate>1999-05-21</shipDate>
        </item>
				<item>
				flaw
				</item>
    </items>
</purchaseOrder>
`
