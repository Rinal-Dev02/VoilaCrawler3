// Code generated by protoc-gen-desc. DO NOT EDIT.

package item

import (
	pb "github.com/voiladev/protobuf/protoc-gen-go/protobuf"
)

var (
	ServiceDescs []*pb.ServiceDesc
	EnumDescs    []*pb.EnumDesc
)

func init() {
	fileDescs, _ := pb.LoadFileDescriptors(
		
		File_chameleon_smelter_v1_crawl_item_linktree_proto,
		
		File_chameleon_smelter_v1_crawl_item_product_proto,
		
		File_chameleon_smelter_v1_crawl_item_tiktok_proto,
		
		File_chameleon_smelter_v1_crawl_item_youtube_proto,
		
	)

	var err error
	if ServiceDescs, err = pb.LoadServiceDescs(fileDescs...); err != nil {
		panic(err)
	}
	if EnumDescs, err = pb.LoadEnumDescs(fileDescs...); err != nil {
		panic(err)
	}
}