package dockerignore

import (
	"slices"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	suite := []struct{
		content string
		expect []string
	}{
		{
			content: "",
			expect: []string{},
		},
		{
			content: "\n",
			expect: []string{},
		},
		{
			content: "# foo\n",
			expect: []string{},
		},
		{
			content: "foo\n",
			expect: []string{"foo"},
		},
		{
			content: "# foo\nbar/\n",
			expect: []string{"bar/"},
		},
		{
			content: "# foo\n\n\nbar/\nbar\n\n",
			expect: []string{"bar/", "bar"},
		},
	}

	for _, test := range suite {
		rdr := strings.NewReader(test.content)
		ignores, err := Parse(rdr)
		if err != nil {
			t.Errorf("test %v unexpectedly fail: %v",
				test.content, err)
			continue
		}
		if slices.Compare(ignores, test.expect) != 0 {
			t.Errorf("Unexpected result for %v; want %+v, got %+v",
				test.content, test.expect, ignores)
		}
	}
}
