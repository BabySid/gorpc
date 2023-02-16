package gin

import (
	"github.com/BabySid/gobase"
	"github.com/BabySid/gorpc/codec"
	r "github.com/gin-gonic/gin/render"
	"net/http"
)

var (
	_                    r.Render = ProtoJson{}
	protoJsonContentType          = []string{"application/json; charset=utf-8"}
)

type ProtoJson struct {
	Data interface{}
}

func (p ProtoJson) Render(w http.ResponseWriter) error {
	if err := WriteJSON(w, p.Data); err != nil {
		gobase.AssertHere()
	}
	return nil
}

func (p ProtoJson) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, protoJsonContentType)
}

// WriteJSON marshals the given interface object and writes it with custom ContentType.
func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	writeContentType(w, protoJsonContentType)
	encoder := codec.GetReplyEncoder(codec.ProtobufCodec)
	jsonBytes, err := encoder(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
