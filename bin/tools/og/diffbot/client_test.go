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
			args: args{ctx, "https://www.urbanoutfitters.com/shop/hanes-beefy-t-x-heron-hues-future-tee?category=graphic-tees-for-men&color=010&type=REGULAR&quantity=1"},
		},
		{
			name: "princesspollly",
			args: args{ctx, "https://us.princesspolly.com/collections/basics/products/madelyn-top-green"},
		},
		{
			name: "lululemon",
			args: args{ctx, "https://shop.lululemon.com/p/bags/Run-All-Day-Backpack-II/_/prod8450240?color=33928&sz=ONESIZE"},
		},
		{
			name: "lululemon",
			args: args{ctx, "https://shop.lululemon.com/p/men-socks/MicroPillow-Tab-Run-Sock-M/_/prod10590024?color=0001"},
		},
		{
			name: "romwe",
			args: args{ctx, "https://us.romwe.com/Ripped-Boyfriend-Jeans-p-1507845-cat-813.html?scici=navbar_GirlsHomePage~~tab01navbar04~~4~~real_809~~~~0"},
		},
		{
			name: "romwe",
			args: args{ctx, "https://us.romwe.com/Letter-Cartoon-Bear-Graphic-Oversized-Tee-p-1658483-cat-669.html?scici=navbar_GirlsHomePage~~tab01navbar03menu01dir01~~3_1_1~~real_669~~~~0"},
		},
		{
			name: "fentybeauty",
			args: args{ctx, "https://fentybeauty.com/products/pro-filtr-soft-matte-longwear-foundation-310"},
		},
		{
			name: "fentybeauty",
			args: args{ctx, "https://fentybeauty.com/products/cheeks-out-freestyle-cream-blush-petal-poppin"},
		},
		{
			name: "dior",
			args: args{ctx, "https://www.dior.com/en_us/products/couture-15ZOD090I607_C800"},
		},
		{
			name: "dior",
			args: args{ctx, "https://www.dior.com/en_us/products/couture-15ZOD090I607_C900-dior-zodiac-square-scarf-black-silk-twill"},
		},
		{
			name: "j.ing",
			args: args{ctx, "https://jingus.com/products/draping-neckline-beige-mini-dress"},
		},
		{
			name: "j.ing",
			args: args{ctx, "https://jingus.com/collections/dresses/category_-_bb_midi-dresses?sort_by=manual"},
		},
		{
			name: "everlane",
			args: args{ctx, "https://www.everlane.com/products/womens-live-in-pant-black?collection=womens-bottoms"},
		},
		{
			name: "everlane",
			args: args{ctx, "https://www.everlane.com/products/womens-forever-sneaker-sage?collection=womens-shoes"},
		},
		{
			name: "nordstrom",
			args: args{ctx, "https://www.nordstrom.com/s/converse-chuck-taylor-all-star-lugged-boot-unisex/5355855?origin=category-personalizedsort&breadcrumb=Home%2FMen%2FShoes%2FBoots&color=102"},
		},
		{
			name: "nordstrom",
			args: args{ctx, "https://www.nordstrom.com/s/nike-md-valiant-sneaker-baby-walker-toddler-little-kid-big-kid/5508505?origin=category-personalizedsort&breadcrumb=Home%2FKids%2FAll%20Boys%2FTween%20Boys&color=101"},
		},
		{
			name: "dsw",
			args: args{ctx, "https://www.dsw.com/en/us/product/brooks-launch-8-running-shoe---womens/506687?activeColor=096"},
		},
		{
			name: "dsw",
			args: args{ctx, "https://www.dsw.com/en/us/product/olive-and-edie-strappie-boot---kids/512800?activeColor=240"},
		},
		{
			name: "bloomingdales",
			args: args{ctx, "https://www.bloomingdales.com/shop/product/clinique-dramatically-different-moisturizing-lotion?ID=868499&CategoryID=2921&cm_sp=categorysplash_Beauty%20%26%20Cosmetics_Beauty-%26-Cosmetics_1-_-row2_advpp_n-_-Clinique:-$15-Off-Every-$75-%26-7-Piece-Gift"},
		},
		{
			name: "bloomingdales",
			args: args{ctx, "https://www.bloomingdales.com/shop/product/badgley-mischka-womens-cher-crystal-buckle-pumps?ID=3587233&CategoryID=17397"},
		},
		{
			name: "ssense",
			args: args{ctx, "https://www.ssense.com/en-us/women/product/rick-owens/green-down-liner-coat/7251811"},
		},
		{
			name: "ssense",
			args: args{ctx, "https://www.ssense.com/en-us/women/product/balenciaga/black-logo-care-mask/6087621"},
		},
		{
			name: "shopbop",
			args: args{ctx, "https://www.shopbop.com/lily-pistola-denim/vp/v=1/1558776584.htm?folderID=13377&fm=other-shopbysize-viewall&os=false&colorId=12A04&ref_=SB_PLP_PDP_NWL_W_CLOTH_DENIM_13377_NB_1&breadcrumb=Clothing%3EJeans"},
		},
		{
			name: "shopbop",
			args: args{ctx, "https://www.shopbop.com/superstar-sneakers-golden-goose/vp/v=1/1571309185.htm?folderID=5475&fm=other-shopbysize-viewall&os=false&colorId=66053&ref_=SB_PLP_PDP_NWL_W_BRAND_GOOSE_5475_NB_1&breadcrumb=Designers%3EGolden%20Goose%3EShoes"},
		},
		{
			name: "neimanmarcus",
			args: args{ctx, "https://www.neimanmarcus.com/p/pq-swim-gisele-one-off-shoulder-coverup-dress-prod242330482?childItemId=NMT21P6_&navpath=cat000000_cat000001_cat81310738_cat82630739&page=0&position=1"},
		},
		{
			name: "neimanmarcus",
			args: args{ctx, "https://www.neimanmarcus.com/p/area-ribbed-long-sleeve-bodysuit-w-crystal-trim-prod240110130?childItemId=NMI235V_&navpath=cat000000_cat000001_cat81310738_cat82630739&page=0&position=6"},
		},
		{
			name: "revolve",
			args: args{ctx, "https://www.revolve.com/nbd-vanity-mini-dress/dp/NBDR-WD2179/?d=Womens&page=1&lc=1&itrownum=1&itcurrpage=1&itview=05"},
		},
		{
			name: "revolve",
			args: args{ctx, "https://www.revolve.com/child-of-wild-tefnut-pearl-hoops/dp/CHIW-WL97/?d=Womens&page=1&lc=7&itrownum=2&itcurrpage=1&itview=05"},
		},
		{
			name: "other stories",
			args: args{ctx, "https://www.stories.com/en_usd/clothing/blouses-shirts/blouses/product.cropped-criss-cross-tie-blouse-brown.0950156003.html"},
		},
		{
			name: "other stories",
			args: args{ctx, "https://www.stories.com/en_usd/clothing/skirts/mini-skirts/product.mini-wrap-skirt-beige.0979153001.html"},
		},
		{
			name: "shein",
			args: args{ctx, "https://us.shein.com/SHEIN-X-Mackenzie-Miller-Ruched-Drawstring-Solid-Tank-Top-p-2163265-cat-1779.html"},
		},
		{
			name: "shein",
			args: args{ctx, "https://us.shein.com/Asymmetrical-Neck-Form-Fitted-Tee-p-2865134-cat-1738.html?scici=WomenHomePage~~ON_FlashSale,CN_flashsale,HZ_0,HI_0~~17_1~~FlashSale~~~~"},
		},
		{
			name: "nike",
			args: args{ctx, "https://www.nike.com/t/sportswear-collection-essentials-womens-fleece-crew-plus-size-5t689v/DD5632-864"},
		},
		{
			name: "nike",
			args: args{ctx, "https://www.nike.com/t/dri-fit-indy-zip-front-womens-light-support-padded-sports-bra-4lQR8Z/DD1197-529"},
		},
		{
			name: "sephora",
			args: args{ctx, "https://www.sephora.com/product/the-ordinary-deciem-glycolic-acid-7-toning-solution-P427406?icid2=recommended%20for%20you:p427406:product"},
		},
		{
			name: "sephora",
			args: args{ctx, "https://www.sephora.com/product/good-genes-all-in-one-lactic-acid-treatment-P309308?icid2=homepage_bi_rewards_us_d_carousel_080121:p124402:product"},
		},
		{
			name: "reformation",
			args: args{ctx, "https://www.thereformation.com/products/stevie-ultra-high-rise-jean-comfort-stretch?color=Malta&via=Z2lkOi8vcmVmb3JtYXRpb24td2VibGluYy9Xb3JrYXJlYTo6Q2F0YWxvZzo6Q2F0ZWdvcnkvNWE2YWRmZDNmOTJlYTExNmNmMDRlOWQz"},
		},
		{
			name: "reformation",
			args: args{ctx, "https://www.thereformation.com/products/assunta-strappy-block-heel-mule?color=Pecan&via=Z2lkOi8vcmVmb3JtYXRpb24td2VibGluYy9Xb3JrYXJlYTo6Q2F0YWxvZzo6Q2F0ZWdvcnkvNWNiZTRlNzdmMzViZTI0YmUzYWRjMTgx"},
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
