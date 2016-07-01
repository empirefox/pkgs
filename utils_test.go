package pkgs

import (
	"fmt"
	"reflect"
	"testing"
)

type MapTest struct {
	in        map[string]interface{}
	path      string
	separator string
	value     interface{}
	out       interface{}
	err       error
}

func TestGetProperty(t *testing.T) {
	cases := []MapTest{
		{
			in:        setupDocument(),
			path:      ".",
			separator: ".",
			out:       setupDocument(),
			err:       nil,
		},
		{
			in:        setupDocument(),
			path:      "one",
			separator: ".",
			out:       setupDocument()["one"].(map[string]interface{}),
			err:       nil,
		},
		{
			in:        setupDocument(),
			path:      "one.two",
			separator: ".",
			out:       setupDocument()["one"].(map[string]interface{})["two"],
			err:       nil,
		},
		{
			in:        setupDocument(),
			path:      "one.two.three",
			separator: ".",
			out:       setupDocument()["one"].(map[string]interface{})["two"].(map[string]interface{})["three"],
			err:       nil,
		},
		{
			in:        setupDocument(),
			path:      "one.two.three[0]",
			separator: ".",
			out:       setupDocument()["one"].(map[string]interface{})["two"].(map[string]interface{})["three"].([]int)[0],
			err:       nil,
		},
		{
			in:        setupDocument(),
			path:      "one.two.three[1]",
			separator: ".",
			out:       setupDocument()["one"].(map[string]interface{})["two"].(map[string]interface{})["three"].([]int)[1],
			err:       nil,
		},
		{
			in:        setupDocument(),
			path:      "one.two.three[2]",
			separator: ".",
			out:       setupDocument()["one"].(map[string]interface{})["two"].(map[string]interface{})["three"].([]int)[2],
			err:       nil,
		},
		{
			in:        setupDocument(),
			path:      "one.two.four",
			separator: ".",
			out:       nil,
			err:       fmt.Errorf("Property %s does not exist", "four"),
		},
		{
			in:        setupDocument(),
			path:      "one.two.four[0]",
			separator: ".",
			out:       nil,
			err:       fmt.Errorf("Property %s does not exist", "four"),
		},
		{
			in:        setupDocument_I(),
			path:      "one[0]",
			separator: ".",
			out:       setupDocument_I()["one"].([]map[string]interface{})[0],
			err:       nil,
		},
		{
			in:        setupDocument_I(),
			path:      "one[1]",
			separator: ".",
			out:       setupDocument_I()["one"].([]map[string]interface{})[1],
			err:       nil,
		},
		{
			in:        setupDocument_I(),
			path:      "one[2]",
			separator: ".",
			out:       setupDocument_I()["one"].([]map[string]interface{})[2],
			err:       nil,
		},
		{
			in:        setupDocument_I(),
			path:      "one[2].map_c",
			separator: ".",
			out:       setupDocument_I()["one"].([]map[string]interface{})[2]["map_c"],
			err:       nil,
		},
		{
			in:        setupDocument_II(),
			path:      "one[1].two[1]",
			separator: ".",
			out:       setupDocument_II()["one"].([]map[string]interface{})[1]["two"].([]map[string]interface{})[1],
			err:       nil,
		},
		{
			in:        setupDocument_II(),
			path:      "one[2].two[1].eight",
			separator: "",
			out:       setupDocument_II()["one"].([]map[string]interface{})[2]["two"].([]map[string]interface{})[1]["eight"],
			err:       nil,
		},
		{
			in:        setupDocument_II(),
			path:      "one[2].two[1].eight.ten",
			separator: "",
			out:       nil,
			err:       fmt.Errorf("Property %s does not exist", "eight.ten"),
		},
		{
			in:        setupDocument_II(),
			path:      "one[1].two[1].eight",
			separator: ".",
			out:       nil,
			err:       fmt.Errorf("Property %s does not exist", "eight"),
		},
		{
			in:        setupDocument_II(),
			path:      "one[3].three[0].seven.eight",
			separator: ".",
			out:       nil,
			err:       fmt.Errorf("Property %s does not exist", "seven"),
		},
		{
			in:        setupDocumentFalse(),
			path:      "one.two",
			separator: ".",
			out:       false,
			err:       nil,
		},
		{
			in:        setupDocumentFalse(),
			path:      "one.two.three.four",
			separator: ".",
			out:       false,
			err:       nil,
		},
		{
			in:        setupDocumentFalse(),
			path:      "one.two.three",
			separator: ".",
			out:       false,
			err:       nil,
		},
		{
			in:        setupDocumentFalse_I(),
			path:      "one[0]",
			separator: ".",
			out:       setupDocumentFalse_I()["one"].([]map[string]interface{})[0],
			err:       nil,
		},
		{
			in:        setupDocumentFalse_I(),
			path:      "one[1]",
			separator: ".",
			out:       setupDocumentFalse_I()["one"].([]map[string]interface{})[1],
			err:       nil,
		},
		{
			in:        setupDocumentFalse_I(),
			path:      "one[0].map_a",
			separator: ".",
			out:       false,
			err:       nil,
		},
		{
			in:        setupDocumentFalse_I(),
			path:      "one[0].map_a[0]",
			separator: ".",
			out:       false,
			err:       nil,
		},
		{
			in:        setupDocumentFalse_II(),
			path:      "one[1].three",
			separator: ".",
			out:       false,
			err:       nil,
		},
		{
			in:        setupDocumentFalse_II(),
			path:      "one[1].three[0].seven.eight",
			separator: ".",
			out:       false,
			err:       nil,
		},
	}

	num_cases := len(cases)
	for i, c := range cases {
		case_index := i + 1

		out, err_case := GetProperty(c.in, c.path, c.separator)
		if !reflect.DeepEqual(c.err, err_case) {
			t.Errorf("\n[%d of %d: Errors should equal] \n\t%v \n \n\t%v", case_index, num_cases, err_case, c.err)
		}
		if !reflect.DeepEqual(out, c.out) {
			t.Errorf("\n[%d of %d: Results should equal] \n\t%v \n \n\t%v", case_index, num_cases, out, c.out)
		}
	}
}

func setupDocument() (document map[string]interface{}) {
	document = map[string]interface{}{
		"one": map[string]interface{}{
			"two": map[string]interface{}{
				"three": []int{
					1, 2, 3,
				},
			},
			"four": map[string]interface{}{
				"five": []int{
					11, 22, 33,
				},
			},
		},
	}

	return
}

func setupDocumentFalse() (document map[string]interface{}) {
	document = map[string]interface{}{
		"one": map[string]interface{}{
			"two": false,
			"four": map[string]interface{}{
				"five": []int{
					11, 22, 33,
				},
			},
		},
	}

	return
}

func setupDocument_I() (document_I map[string]interface{}) {
	document_I = map[string]interface{}{
		"one": []map[string]interface{}{
			{"map_a": []int{1, 2, 3}},
			{"map_b": []int{4, 5, 6}},
			{"map_c": []int{7, 8, 9}},
		},
	}
	return
}

func setupDocumentFalse_I() (document_I map[string]interface{}) {
	document_I = map[string]interface{}{
		"one": []map[string]interface{}{
			{"map_a": false},
			{"map_b": []int{4, 5, 6}},
		},
	}
	return
}

func setupDocument_II() (document_II map[string]interface{}) {
	document_II = map[string]interface{}{
		"one": []map[string]interface{}{
			{
				"two": []map[string]interface{}{
					{"three": "got three"},
					{"four": "got four"},
				},
			},
			{
				"two": []map[string]interface{}{
					{"five": "got five"},
					{"six": "got six"},
				},
			},
			{
				"two": []map[string]interface{}{
					{"seven": "got seven"},
					{"eight": "got eight"},
				},
			},
			{
				"three": []map[string]interface{}{
					{"four": map[string]interface{}{
						"five": "six",
					}},
					{"seven": map[string]interface{}{
						"eight": "ten",
					}},
				},
			},
		},
	}
	return
}

func setupDocumentFalse_II() (document_II map[string]interface{}) {
	document_II = map[string]interface{}{
		"one": []map[string]interface{}{
			{
				"two": []map[string]interface{}{
					{"three": "got three"},
					{"four": "got four"},
				},
			},
			{
				"three": false,
			},
		},
	}
	return
}
