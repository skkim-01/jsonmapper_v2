package jsonmapper_v2

import (
	"testing"
)

var test_json_string string = `
{
	"root.int" : 0,
	"root.string" : "root_value",
	"root.bool" : true,
	"child.1.map" : {
		"child.1.subint" : 1,
		"child.1.substring" : "child_1_value",
		"child.1.subslice" : [
			{ "id": "5001", "type": "None" },
			{ "id": "5002", "type": "Glazed" },
			{ "id": "5005", "type": "Sugar" }
		],
		"child.1.submap" : {
			"child.1.1.subint" : 2,
			"child.1.1.substring" : "child_1_1_value"
		}
	}
}
`

func BenchmarkFindString(b *testing.B) {
	j, _ := NewJsonMapStr(test_json_string)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = j.FindStringOr("child.1.map.child.1.subslice.1.id", "")
	}
}

func BenchmarkRemoveSubslice(b *testing.B) {
	j, _ := NewJsonMapStr(test_json_string)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = j.Remove("child.1.map")
	}
}

func BenchmarkRemoveSubmap(b *testing.B) {
	j, _ := NewJsonMapStr(test_json_string)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = j.Remove("child.1.map.child.1.subslice.1")
	}
}
