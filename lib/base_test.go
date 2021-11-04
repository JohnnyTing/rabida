package lib

import (
	"testing"
	"time"
)

func TestRandDuration(t *testing.T) {
	type args struct {
		min time.Duration
		max time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "",
			args: args{
				min: 3 * time.Second,
				max: 10 * time.Second,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandDuration(tt.args.min, tt.args.max); got != tt.want {
				t.Errorf("RandDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
