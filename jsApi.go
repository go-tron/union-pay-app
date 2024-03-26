package unionPayApp

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/go-tron/local-time"
	"github.com/go-tron/random"
	"strconv"
)

type JsApiConfig struct {
	Debug     bool   `json:"debug"`
	AppId     string `json:"appId"`
	Timestamp int64  `json:"timestamp"`
	NonceStr  string `json:"nonceStr"`
	Signature string `json:"signature"`
}

func (upa *UnionPayApp) GetJsApiConfig(url string) (*JsApiConfig, error) {

	frontToken, err := upa.GetFrontToken()
	if err != nil {
		return nil, err
	}

	jsApiConfig := &JsApiConfig{
		AppId:     upa.AppId,
		Timestamp: localTime.Now().Unix(),
		NonceStr:  random.String(10),
		Signature: "",
	}
	var signStr = "appId=" + jsApiConfig.AppId + "&frontToken=" + frontToken.FrontToken + "&nonceStr=" + jsApiConfig.NonceStr + "&timestamp=" + strconv.FormatInt(jsApiConfig.Timestamp, 10) + "&url=" + url
	hash := sha256.New()
	hash.Write([]byte(signStr))
	jsApiConfig.Signature = hex.EncodeToString(hash.Sum(nil))
	return jsApiConfig, nil
}
