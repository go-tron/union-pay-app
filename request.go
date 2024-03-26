package unionPayApp

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	baseError "github.com/go-tron/base-error"
	"github.com/go-tron/types/mapUtil"
	"github.com/tidwall/gjson"
)

var (
	ErrorParam         = baseError.SystemFactory("3101", "云闪付参数错误:{}")
	ErrorMethod        = baseError.SystemFactory("3102", "云闪付方法无效:{}")
	ErrorAuthorize     = baseError.System("3103", "云闪付授权失败")
	ErrorRequest       = baseError.System("3104", "云闪付服务连接失败")
	ErrorUnmarshalBody = baseError.System("3105", "云闪付消息解析失败")
	ErrorCode          = baseError.SystemFactory("3110")
)

func (upa *UnionPayApp) Request(name string, data map[string]interface{}, res interface{}) (result interface{}, err error) {

	request, _ := json.Marshal(data)
	response := ""
	upa.Logger.Info(string(request),
		upa.Logger.Field("openId", data["openId"]),
		upa.Logger.Field("name", name),
		upa.Logger.Field("type", "request"),
	)
	defer func() {
		upa.Logger.Info(response,
			upa.Logger.Field("openId", data["openId"]),
			upa.Logger.Field("name", name),
			upa.Logger.Field("type", "response"),
			upa.Logger.Field("error", err))
	}()

	url := SDKConfig[name]
	if url == "" {
		return nil, ErrorMethod(name)
	}

	backendToken, err := upa.GetBackendToken()
	if err != nil {
		return nil, ErrorAuthorize
	}
	data["backendToken"] = backendToken.BackendToken

	resp, err := resty.New().R().
		SetBody(data).
		Post(url)
	if err != nil {
		return nil, ErrorRequest
	}

	response = string(resp.Body())
	code := gjson.Get(response, "resp").String()
	message := gjson.Get(response, "msg").String()

	if code != "00" {
		if message == "" {
			message = name
		}
		return nil, ErrorCode(fmt.Sprintf("(%s)%s", code, message))
	}

	bodyData := gjson.Get(response, "params").Value()
	if res == nil {
		return bodyData, nil
	}

	if err := mapUtil.ToStruct(bodyData, res); err != nil {
		return nil, ErrorUnmarshalBody
	}

	return res, nil
}
