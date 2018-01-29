package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	for _, store := range []string{"S3", "Spaces", "Diskv", "Redis", ""} {
		NewStore(store)
		assert.NotNil(t, store)
	}
}

/* func TestParseType(t *testing.T) {
	type args struct {
		t string
	}
	//var tp Type
	tests := []struct {
		name string
		args args
		want *Type
	}{
		{name: "S3Type",
			args: args{
				t: "s3",
			},
			want: nil,
		},
		{name: "DOSpaces",
			args: args{
				t: "spaces",
			},
			want:    nil,
			
		},
		{name: "Diskv",
			args: args{
				t: "diskv",
			},
			want:    nil,
	
		},
		{name: "Redis",
			args: args{
				t: "redis",
			},
			want:    nil,

		},
		{name: "Unsupported Type",
			args: args{
				t: "unsupported",
			},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := ParseType(tt.args.t); got != tt.want {
				t.Errorf("ParseType() = %v, want %v", got, tt.want)
			}
		})
	}
}
 */