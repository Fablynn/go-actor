package httplib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cast"
)

var (
	defaultClient = &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     30 * time.Second,
		},
	}
)

func POST(url string, params map[string]interface{}, rsp proto.Message) error {
	payload := bytes.Buffer{}
	writer := multipart.NewWriter(&payload)
	for key, val := range params {
		if err := writer.WriteField(key, cast.ToString(val)); err != nil {
			return err
		}
	}
	defer writer.Close()

	req, err := http.NewRequest("POST", url, &payload)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 请求http服务
	resp, err := defaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 解析返回值
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求未成功: 状态码=%d, 响应=%s", resp.StatusCode, string(body))
	}
	return json.Unmarshal(body, rsp)
}
