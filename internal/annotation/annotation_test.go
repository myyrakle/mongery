package annotation

import (
	"reflect"
	"testing"
)

// @Entity(1, "2", 3) => []Entity{"1", "2", "3"}
func Test_ParseParameters(t *testing.T) {
	type testCase struct {
		input string
		want  []string
	}

	testCases := []testCase{
		{
			input: `@Entity(1, "2", 3)`,
			want:  []string{"1", "2", "3"},
		},
		{
			input: `@Entity("foo", 1, 3)`,
			want:  []string{"foo", "1", "3"},
		},
	}

	for _, testCase := range testCases {
		got := ParseParameters(testCase.input)

		if !reflect.DeepEqual(got, testCase.want) {
			t.Errorf("ParseParameters(%s) = %#v, want %#v", testCase.input, got, testCase.want)
		}
	}
}
