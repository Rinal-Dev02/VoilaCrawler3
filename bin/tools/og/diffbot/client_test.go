package diffbot

import (
	"context"
	"testing"

	"github.com/voiladev/go-framework/glog"
)

func TestDiffbotCient_Fetch(t *testing.T) {
	c, _ := New("4f738a5f64cfa3f43dda5a1845c246d4", glog.New(glog.LogLevelDebug))
	type args struct {
		ctx    context.Context
		rawurl string
	}
	ctx := context.Background()
	tests := []struct {
		name string
		args args
	}{
		{
			name: "urbanoutfitters",
			args: args{
				ctx:    ctx,
				rawurl: "https://www.urbanoutfitters.com/shop/hanes-beefy-t-x-heron-hues-future-tee?category=graphic-tees-for-men&color=010&type=REGULAR&quantity=1",
			},
		},
		{
			name: "princesspollly",
			args: args{
				ctx:    ctx,
				rawurl: "https://us.princesspolly.com/collections/basics/products/madelyn-top-green",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prods, err := c.Fetch(tt.args.ctx, tt.args.rawurl)
			if err != nil {
				t.Error(err)
			} else if len(prods) == 0 {
				t.Errorf("no product info found")
			}
		})
	}
}
