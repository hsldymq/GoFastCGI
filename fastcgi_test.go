package main

import "testing"

func TestNameValuePairMarshaling(t *testing.T) {
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

		for idx, each := range cases {
			actual := &NameValuePair{}
			err := actual.UnmarshalBinary(each.Data)
			if err != nil && !each.ExpectError {
				t.Fatalf("%d: expect unmarshal success, got error", idx)
			}

			if actual.Name != each.Expect.Name || actual.Value != each.Expect.Value {
				t.Fatalf("%d: expect name and value to be %s and %s, got %s, %s", idx, each.Expect.Name, each.Expect.Value, actual.Name, actual.Value)
			}

		}
	})

	t.Run("test marshal/unmarshal pair", func(t *testing.T) {
		cases := []*NameValuePair{
			{Name: "", Value: ""},
			{Name: "QUERY_STRING", Value: ""},
			{Name: "GATEWAY_INTERFACE", Value: "CGI/1.1"},
		}
		for idx, each := range cases {
			data, err := each.MarshalBinary()
			if err != nil {
				t.Fatalf("%d: marshal, got err", idx)
			}
			unmarshaled := &NameValuePair{}
			if err := unmarshaled.UnmarshalBinary(data); err != nil {
				t.Fatalf("%d: unmarshal, got err", idx)
			}
			if each.Name != unmarshaled.Name || each.Value != unmarshaled.Value {
				t.Fatalf("%d: original name: %s, value: %s, unmarshaled name: %s, value: %s", idx, each.Name, each.Value, unmarshaled.Name, unmarshaled.Value)
			}
		}
	})
}
