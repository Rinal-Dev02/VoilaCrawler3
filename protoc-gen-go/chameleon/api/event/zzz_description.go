// Code generated by protoc-gen-desc. DO NOT EDIT.\n

package event

import (
    pb "github.com/voiladev/protobuf/protoc-gen-go/protobuf"
)

var (
	ServiceDescs []*pb.ServiceDesc
	EnumDescs    []*pb.EnumDesc
)

func init() {
	fileDescs, _ := pb.LoadFileDescriptors(
		
		File_chameleon_api_event_data_proto,
		
    )

    var err error
	if ServiceDescs, err = pb.LoadServiceDescs(fileDescs...); err != nil {
		panic(err)
	}
	if EnumDescs, err = pb.LoadEnumDescs(fileDescs...); err != nil {
		panic(err)
	}
}