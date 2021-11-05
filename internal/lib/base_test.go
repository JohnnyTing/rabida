package lib

import (
	"reflect"
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

func TestFlat(t *testing.T) {
	type args struct {
		data map[string][]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantRet []map[string]interface{}
	}{
		{
			name: "",
			args: args{
				data: map[string][]interface{}{
					"name": {"jack", "lucy", "lily"},
					"age":  {18, 19, 20},
				},
			},
			wantRet: []map[string]interface{}{
				{
					"name": "jack",
					"age":  18,
				},
				{
					"name": "lucy",
					"age":  19,
				},
				{
					"name": "lily",
					"age":  20,
				},
			},
		},
		{
			name: "",
			args: args{
				data: map[string][]interface{}{
					"name": {"jack", []string{"lucy", "green"}, map[string]interface{}{"name": "lily"}},
					"age":  {18, []int{19, 1}, 20},
				},
			},
			wantRet: []map[string]interface{}{
				{
					"name": "jack",
					"age":  18,
				},
				{
					"name": []string{"lucy", "green"},
					"age":  []int{19, 1},
				},
				{
					"name": map[string]interface{}{"name": "lily"},
					"age":  20,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRet := Flat(tt.args.data); !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("Flat() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}
