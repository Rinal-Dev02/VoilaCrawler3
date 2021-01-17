package context

import (
	"context"
	"reflect"
	"testing"
)

func TestRetrieveAllValues(t *testing.T) {
	withCancel := func(ctx context.Context) context.Context {
		ctx, _ = context.WithCancel(ctx)
		return ctx
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want map[interface{}]interface{}
	}{
		{
			name: "WithValue depth-1",
			args: args{
				ctx: context.WithValue(context.Background(), "12345", "12345"),
			},
			want: map[interface{}]interface{}{
				"12345": "12345",
			},
		}, {
			name: "WithValue depth-2",
			args: args{
				ctx: context.WithValue(context.WithValue(context.Background(), "12345", "12345"), "234", "234"),
			},
			want: map[interface{}]interface{}{
				"12345": "12345",
				"234":   "234",
			},
		}, {
			name: "WithValues depth-1",
			args: args{
				ctx: WithValues(context.Background(), "12345", "12345", "234", "234", "345", "345"),
			},
			want: map[interface{}]interface{}{
				"12345": "12345",
				"234":   "234",
				"345":   "345",
			},
		}, {
			name: "WithValues depth-2",
			args: args{
				ctx: WithValues(WithValues(context.Background(), "12345", "12345", "234", "234", "345", "345"), "345", "678", "789", "789"),
			},
			want: map[interface{}]interface{}{
				"12345": "12345",
				"234":   "234",
				"345":   "678",
				"789":   "789",
			},
		},
		{
			name: "Mix depth-2",
			args: args{
				ctx: WithValues(context.WithValue(context.Background(), "12345", "12345"), "234", "234", "345", "345"),
			},
			want: map[interface{}]interface{}{
				"12345": "12345",
				"234":   "234",
				"345":   "345",
			},
		},
		{
			name: "Mix depth-3 without overwrite",
			args: args{
				ctx: context.WithValue(WithValues(context.WithValue(context.Background(), "12345", "12345"), "234", "234", "345", "345"), "678", "678"),
			},
			want: map[interface{}]interface{}{
				"12345": "12345",
				"234":   "234",
				"345":   "345",
				"678":   "678",
			},
		},
		{
			name: "Mix depth-3 overwrite",
			args: args{
				ctx: context.WithValue(WithValues(context.WithValue(context.Background(), "12345", "12345"), "234", "234", "345", "345"), "345", "678"),
			},
			want: map[interface{}]interface{}{
				"12345": "12345",
				"234":   "234",
				"345":   "678",
			},
		},
		{
			name: "Mix depth-4 overwrite",
			args: args{
				ctx: context.WithValue(withCancel(WithValues(context.WithValue(context.Background(), "12345", "12345"), "234", "234", "345", "345")), "345", "678"),
			},
			want: map[interface{}]interface{}{
				"12345": "12345",
				"234":   "234",
				"345":   "678",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RetrieveAllValues(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveAllValues() = %v, want %v", got, tt.want)
			}
		})
	}
}
