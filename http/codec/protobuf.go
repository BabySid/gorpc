package codec

import (
	"bytes"
	"encoding/json"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
)

func ProtobufParamsDecoder(raw interface{}, params interface{}) error {
	bs, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	newReader, err := utilities.IOReaderFactory(bytes.NewReader(bs))
	if err != nil {
		return err
	}

	defaultMarshal := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	err = defaultMarshal.NewDecoder(newReader()).Decode(params)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func ProtobufReplyEncoder(reply interface{}) ([]byte, error) {
	defaultMarshal := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
			UseEnumNumbers:  false,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}
	return defaultMarshal.Marshal(reply)
}
