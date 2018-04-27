package api

import (
	"encoding/xml"
	"testing"
)

type Person struct {
	XMLName xml.Name  `xml:"person"`
	Name    string    `xml:"name,attr,omitempty"`
	Place   *Location `xml:"place,omitempty"`
}

type Location struct {
	Country string `xml:"country,omitempty"`
	City    string `xml:"city,omitempty"`
}

type PersonEmbedded struct {
	XMLName xml.Name `xml:"personEmbedded"`
	Person
	Location
}

type NilStruct *Person

func TestXmlMarshalWithSelfClosingTags(t *testing.T) {
	var tests = []struct {
		value  interface{}
		expect string
	}{
		// Nil values
		{value: nil, expect: ``},
		{value: new(NilStruct), expect: ``},

		// Values
		{value: Person{}, expect: `<person/>`},
		{value: Person{Name: "John Doe"}, expect: `<person name="John Doe"/>`},
		{
			value:  Person{Name: "Jane Doe", Place: &Location{Country: "unknown", City: "unknown"}},
			expect: `<person name="Jane Doe"><place><country>unknown</country><city>unknown</city></place></person>`,
		},
		{
			value:  &PersonEmbedded{Person: Person{Name: "John Doe"}, Location: Location{Country: "unknown"}},
			expect: `<personEmbedded name="John Doe"><country>unknown</country></personEmbedded>`,
		},

		// Pointers to values
		{value: &Person{}, expect: `<person/>`},
		{value: &Person{Name: "John Doe"}, expect: `<person name="John Doe"/>`},
	}

	for _, test := range tests {
		data, err := xmlMarshalWithSelfClosingTags(test.value)
		if err != nil {
			t.Fatalf("Cannot marshal %+v: %s", test.value, err)
		}

		got := string(data)
		if got != test.expect {
			t.Fatalf("Got XML '%s', expect '%s'", got, test.expect)
		}
	}
}
