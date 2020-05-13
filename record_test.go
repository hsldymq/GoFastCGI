package main

import (
	"strings"
	"testing"
)

func TestNameValuePair_Marshaling(t *testing.T) {
	t.Run("test NameValuePair.UnmarshalBinary", func(t *testing.T) {
		cases := []struct {
			Data        []byte
			Expect      *NameValuePair
			ExpectError bool
		}{
			{Data: []byte{12, 0, 81, 85, 69, 82, 89, 95, 83, 84, 82, 73, 78, 71}, Expect: &NameValuePair{Name: "QUERY_STRING", Value: ""}},
			{Data: []byte{14, 3, 82, 69, 81, 85, 69, 83, 84, 95, 77, 69, 84, 72, 79, 68, 71, 69, 84}, Expect: &NameValuePair{Name: "REQUEST_METHOD", Value: "GET"}},
			{Data: []byte{17, 7, 71, 65, 84, 69, 87, 65, 89, 95, 73, 78, 84, 69, 82, 70, 65, 67, 69, 67, 71, 73, 47, 49, 46, 49}, Expect: &NameValuePair{Name: "GATEWAY_INTERFACE", Value: "CGI/1.1"}},
			{Data: []byte{128, 12}, Expect: &NameValuePair{}, ExpectError: true},
			{Data: []byte{12, 128}, Expect: &NameValuePair{}, ExpectError: true},
			{Data: []byte{2, 2, 48}, Expect: &NameValuePair{}, ExpectError: true},
		}

		for idx, eachCase := range cases {
			actual := &NameValuePair{}
			err := actual.UnmarshalBinary(eachCase.Data)
			if err != nil && !eachCase.ExpectError {
				t.Fatalf("%d: expect unmarshal success, got error", idx)
			}

			if actual.Name != eachCase.Expect.Name || actual.Value != eachCase.Expect.Value {
				t.Fatalf("%d: expect name and value to be %s and %s, got %s, %s", idx, eachCase.Expect.Name, eachCase.Expect.Value, actual.Name, actual.Value)
			}

		}
	})

	t.Run("test marshal/unmarshal pair", func(t *testing.T) {
		cases := []*NameValuePair{
			{Name: "", Value: ""},
			{Name: "QUERY_STRING", Value: ""},
			{Name: "GATEWAY_INTERFACE", Value: "CGI/1.1"},
		}
		for idx, eachCase := range cases {
			data, err := eachCase.MarshalBinary()
			if err != nil {
				t.Fatalf("%d: marshal, got err", idx)
			}
			unmarshaled := &NameValuePair{}
			if err := unmarshaled.UnmarshalBinary(data); err != nil {
				t.Fatalf("%d: unmarshal, got err", idx)
			}
			if eachCase.Name != unmarshaled.Name || eachCase.Value != unmarshaled.Value {
				t.Fatalf("%d: original name: %s, value: %s, unmarshaled name: %s, value: %s", idx, eachCase.Name, eachCase.Value, unmarshaled.Name, unmarshaled.Value)
			}
		}
	})
}

func TestNameValuePair_Length(t *testing.T) {
	cases := []struct {
		Name   string
		Value  string
		Expect uint16
	}{
		{Name: "A", Value: "", Expect: 3},
		{Name: "A", Value: "A", Expect: 4},
		{Name: strings.Repeat("A", 127), Value: strings.Repeat("A", 127), Expect: 2 + 127*2},
		{Name: strings.Repeat("A", 128), Value: strings.Repeat("A", 127), Expect: 5 + 255},
		{Name: strings.Repeat("A", 127), Value: strings.Repeat("A", 128), Expect: 5 + 255},
		{Name: strings.Repeat("A", 128), Value: strings.Repeat("A", 128), Expect: 8 + 256},
	}

	for idx, eachCase := range cases {
		nv := &NameValuePair{Name: eachCase.Name, Value: eachCase.Value}
		actual := nv.Length()
		if actual != eachCase.Expect {
			t.Fatalf("%d: expect %d, got %d", idx, eachCase.Expect, actual)
		}
	}
}
