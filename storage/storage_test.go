package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseType(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name    string
		args    args
		want    Type
		wantErr bool
	}{
		{name: "S3Type",
			args: args{
				t: "s3",
			},
			want:    S3,
			wantErr: false,
		},
		{name: "DOSpaces",
			args: args{
				t: "spaces",
			},
			want:    Spaces,
			wantErr: false,
		},
		{name: "Diskv",
			args: args{
				t: "diskv",
			},
			want:    Diskv,
			wantErr: false,
		},
		{name: "Redis",
			args: args{
				t: "redis",
			},
			want:    Redis,
			wantErr: false,
		},
		{name: "Unsupported Type",
			args: args{
				t: "unsupported",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseType(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewStore(t *testing.T) {
	for _, store := range []Type{S3, Spaces, Diskv, Redis, ""} {
		NewStore(store)
		assert.NotNil(t, store)
	}
}
