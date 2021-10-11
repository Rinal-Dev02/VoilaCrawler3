package brand

import (
	"testing"
)

func TestGetBrand(t *testing.T) {
	type args struct {
		domain string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "acne studios",
			args: args{domain: "www.acnestudios.com"},
			want: "Acne Studios",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBrand(tt.args.domain); got != tt.want {
				t.Errorf("GetBrand() = %v, want %v", got, tt.want)
			}
		})
	}
}
