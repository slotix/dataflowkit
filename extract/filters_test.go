package extract

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_filterText(t *testing.T) {
	type args struct {
		data    string
		filters []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "trim + lowercase",
			args: args{
				data: "\n\t 	This is The tEST\t\n",
				filters: []string{"trim", "lowercase"},
			},
			want: "this is the test",
		},
		{name: "trim + uppercase",
			args: args{
				data: "\n\t 	This is The tEST\t\n",
				filters: []string{"trim", "UpPeRcAse"},
			},
			want: "THIS IS THE TEST",
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterText(tt.args.data, tt.args.filters); got != tt.want {
				t.Errorf("filterText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filterTextMW(t *testing.T) {
	filtered := filterTextMW("\n\t Test ChAinEd fILterS \n\t\t", strings.TrimSpace, strings.ToLower)
	assert.Equal(t, "test chained filters", filtered )

}
