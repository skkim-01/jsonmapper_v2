package main

import (
	"fmt"

	jsonmapper_v2 "github.com/skkim-01/jsonmapper-v2"
)

func main() {
	// 샘플 JSON 데이터
	jsonStr := `{
		"testData": {
			"number": 25,
			"string": "hello",
			"bool": true,
			"nested": {
				"number": 15,
				"string": "world"
			},
			"sliced": [
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
				100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111
			],
			"s2": [
				{"id": 1, "name": "alice"},
				{"id": 2, "name": "bob"},
				{"id": 3, "name": "cindy"}
			]
		}
	}`

	jm, err := jsonmapper_v2.NewJsonMapStr(jsonStr)
	if err != nil {
		fmt.Println("Error creating JsonMapper:", err)
		return
	}

	// 테스트 케이스 정의
	testCases := []struct {
		description  string
		conditions   interface{}
		expectedKeys []string
	}{
		{
			description: "Single eq condition (Find 'hello')",
			conditions: map[string]interface{}{
				"eq": "hello",
			},
			expectedKeys: []string{"testData.string"}, // "hello"는 testData.string에만 해당
		},
		{
			description: "Single neq condition (Not equal to 'world')",
			conditions: map[string]interface{}{
				"neq": "world",
			},
			expectedKeys: []string{"testData.string", "testData.bool", "testData.number", "testData.nested.number"}, // 30이 아닌 숫자는 testData.number와 testData.nested.number
		},
		{
			description: "And condition (Greater than 20 and less than 30)",
			conditions: map[string][]map[string]interface{}{
				"and": {
					{"gt": 20},
					{"lt": 30},
				},
			},
			expectedKeys: []string{"testData.number"}, // 25는 20보다 크고 30보다 작음
		},
		{
			description: "Or condition (Equal to 15 or 'world')",
			conditions: map[string][]map[string]interface{}{
				"or": {
					{"eq": 15},
					{"eq": "world"},
				},
			},
			expectedKeys: []string{"testData.nested.number", "testData.nested.string"}, // 15와 "world"는 testData.nested 아래에 있음
		},
		{
			description: "Xor condition (Either true or 'hello', but not both)",
			conditions: map[string][]map[string]interface{}{
				"xor": {
					{"eq": true},
					{"eq": "hello"},
				},
			},
			expectedKeys: []string{"testData.bool", "testData.string"}, // true와 "hello"는 각각 다른 키에 있으므로 xor 조건 만족
		},
		{
			description: "Nor condition (Neither 25 nor 'hello')",
			conditions: map[string][]map[string]interface{}{
				"nor": {
					{"neq": 25},
					{"neq": "hello"},
				},
			},
			expectedKeys: []string{}, // 25와 "hello"는 포함되므로 해당하는 키 없음
		},
		{
			description: "gt condition (greater than 100)",
			conditions: map[string]interface{}{
				"gt": 100,
			},
			expectedKeys: []string{}, // 25와 "hello"는 포함되므로 해당하는 키 없음
		},
	}

	// 테스트 케이스 실행
	for _, tc := range testCases {
		fmt.Printf("Testing: %s\n", tc.description)
		fmt.Printf("Expected keys (not correct): %v\n", tc.expectedKeys)
		results, err := jm.FindAllWithCondition("testData", tc.conditions)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("Found keys: %v\n\n", results)

	}

	////////////////////////

	fmt.Println("#DBG \t Find \t testData.sliced[0] \t expected: 1")
	fmt.Println(jm.Find("testData.sliced[0]"))
	fmt.Println()

	fmt.Println("#DBG \t Remove \t testData.sliced[1] \t expected: [1, 3, 4, ...]")
	fmt.Println(jm.Remove("testData.sliced[1]"))
	fmt.Println(jm.Find("testData.sliced"))
	fmt.Println()

	fmt.Println("#DBG \t Add \t testData.sliced[-1] \t expected: [..., 150]")
	fmt.Println(jm.Add("testData.sliced[-1]", 150))
	fmt.Println(jm.Find("testData.sliced"))
	fmt.Println()

	fmt.Println(jm.Find("testData.s2"))
	fmt.Println(jm.Find("testData.s2.0"))
	fmt.Println(jm.Find("testData.s2[1]"))
	fmt.Println(jm.Find("testData.s2[2]"))
	fmt.Println(jm.Find("testData.s2[3]"))
	fmt.Println()

	fmt.Println("#DBG \t Add \t testData.s2[-1] \t expected: [..., {\"id\": 4, \"name\": \"diana\"}]")
	fmt.Println(jm.Add("testData.s2[-1]", map[string]interface{}{"id": 4, "name": "diana"}))
	fmt.Println(jm.Find("testData.s2"))
	fmt.Println()

	fmt.Println("#DBG \t Remove \t testData.s2[0] \t expected: [id:2, 3, 4]")
	fmt.Println(jm.Remove("testData.s2[0]"))
	fmt.Println(jm.Find("testData.s2"))
	fmt.Println()

	fmt.Println("#DBG \t Remove \t testData.s2[-1] \t expected: [id:2, 3]")
	fmt.Println(jm.Remove("testData.s2[-1]"))
	fmt.Println(jm.Find("testData.s2"))
	fmt.Println()

	fmt.Println("#DBG \t Add \t testData.s3 \t expected: s3: {}")
	fmt.Println(jm.Add("testData.s3", map[string]interface{}{"id": 4, "name": "diana"}))
	fmt.Println(jm.Find("testData.s3"))
	fmt.Println()

	fmt.Println(jm.Print())
}
