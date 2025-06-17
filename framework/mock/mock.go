package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"poker_server/common/pb"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cast"
)

var (
	ip   string = "localhost"
	port int    = 22345
)

func Init(ipp string, portt int) {
	ip = ipp
	port = portt
}

func Request(uid, routeId uint64, cmd pb.CMD, msg proto.Message) error {
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	targetUrl := fmt.Sprintf("http://%s:%d/api", ip, port)
	u, _ := url.ParseRequestURI(targetUrl)

	params := url.Values{}
	params.Set("cmd", cast.ToString(uint32(cmd)))
	params.Set("json", string(buf))
	params.Set("route_id", cast.ToString(routeId))
	params.Set("uid", cast.ToString(uid))

	u.RawQuery = params.Encode()
	_, err = http.Get(u.String())
	return err
}
