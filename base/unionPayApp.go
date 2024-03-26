package base

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/go-tron/config"
	"github.com/go-tron/crypto/encoding"
	localTime "github.com/go-tron/local-time"
	"github.com/go-tron/logger"
	"github.com/go-tron/random"
	"github.com/go-tron/redis"
	"github.com/go-tron/types/mapUtil"
	"sync"
	"time"
)

func NewWithConfig(c *config.Config, client *redis.Redis) *UnionPayApp {
	return New(&Config{
		AppId:      c.GetString("unionPayApp.appId"),
		Secret:     c.GetString("unionPayApp.secret"),
		EncryptKey: c.GetString("unionPayApp.encryptKey"),
		Redis:      client,
		Logger:     logger.NewZapWithConfig(c, "unionPayApp-base", "info"),
	})
}

func New(c *Config) *UnionPayApp {

	if c == nil {
		panic("config 必须设置")
	}
	if c.AppId == "" {
		panic("AppId 必须设置")
	}
	if c.Secret == "" {
		panic("Secret 必须设置")
	}
	if c.EncryptKey == "" {
		panic("EncryptKey 必须设置")
	}
	if c.Logger == nil {
		panic("Logger 必须设置")
	}
	if c.Redis == nil {
		panic("Redis 必须设置")
	}

	key, err := hex.DecodeString(c.EncryptKey)
	if err != nil {
		panic(err)
	}
	c.encryptKeyByte = key
	return &UnionPayApp{
		Config: c,
	}
}

type UnionPayApp struct {
	*Config
	backendTokenLock sync.Mutex
	frontTokenLock   sync.Mutex
	backendToken     *BackendToken `json:"backendToken"`
	frontToken       *FrontToken   `json:"frontToken"`
}

type Config struct {
	AppId          string        `json:"appId"`
	Secret         string        `json:"secret"`
	EncryptKey     string        `json:"encryptKey"`
	encryptKeyByte []byte        `json:"encryptKeyByte"`
	Logger         logger.Logger `json:"logger"`
	Redis          *redis.Redis  `json:"redis"`
}

type BackendToken struct {
	BackendToken string       `json:"backendToken"`
	ExpiresIn    int64        `json:"expiresIn"`
	ticker       *time.Ticker `json:"-"`
}

type FrontToken struct {
	FrontToken string       `json:"frontToken"`
	ExpiresIn  int64        `json:"expiresIn"`
	ticker     *time.Ticker `json:"-"`
}

const (
	BackendTokenPrefix = "upa-backend-token:"
	FrontTokenPrefix   = "upa-front-token:"
)

type BackendTokenRes struct {
	Resp   string `json:"resp"`
	Msg    string `json:"msg"`
	Params struct {
		BackendToken string      `json:"backendToken"`
		ExpiresIn    json.Number `json:"expiresIn"`
	} `json:"params"`
}

type FrontTokenRes struct {
	Resp   string `json:"resp"`
	Msg    string `json:"msg"`
	Params struct {
		FrontToken string      `json:"frontToken"`
		ExpiresIn  json.Number `json:"expiresIn"`
	} `json:"params"`
}

func (upa *UnionPayApp) Sign(req map[string]interface{}) {
	req["timestamp"] = fmt.Sprint(localTime.Now().Unix())
	req["nonceStr"] = random.String(16)
	req["secret"] = upa.Secret

	signStr := mapUtil.ToSortString(req)
	var hashMethod = sha256.New()
	hashMethod.Write([]byte(signStr))
	signature := (&encoding.Hex{}).EncodeToString(hashMethod.Sum(nil))
	req["signature"] = signature
	delete(req, "secret")
}

func (upa *UnionPayApp) ClearBackendToken() {
	if upa.backendToken == nil {
		return
	}
	upa.backendToken.BackendToken = ""
	upa.backendToken.ExpiresIn = 0
	if upa.backendToken.ticker != nil {
		upa.backendToken.ticker.Stop()
	}
}

func (upa *UnionPayApp) SetBackendToken(backendToken string, expiresIn int64) {

	upa.backendToken = &BackendToken{
		BackendToken: backendToken,
		ExpiresIn:    expiresIn,
		ticker:       time.NewTicker(time.Second),
	}

	go func() {
		for upa.backendToken.ExpiresIn > 0 {
			<-upa.backendToken.ticker.C

			if upa.backendToken == nil {
				break
			}

			upa.backendToken.ExpiresIn--
			upa.Logger.Debug(fmt.Sprintf("backendToken.ExpiresIn:%d", upa.backendToken.ExpiresIn), upa.Logger.Field("appId", upa.AppId))

			if upa.backendToken.ExpiresIn <= 0 {
				upa.ClearBackendToken()
				break
			}
		}
	}()
}

func (upa *UnionPayApp) ClearFrontToken() {
	if upa.frontToken == nil {
		return
	}
	upa.frontToken.FrontToken = ""
	upa.frontToken.ExpiresIn = 0
	if upa.frontToken.ticker != nil {
		upa.frontToken.ticker.Stop()
	}
}

func (upa *UnionPayApp) SetFrontToken(frontToken string, expiresIn int64) {

	upa.frontToken = &FrontToken{
		FrontToken: frontToken,
		ExpiresIn:  expiresIn,
		ticker:     time.NewTicker(time.Second),
	}

	go func() {
		for upa.frontToken.ExpiresIn > 0 {
			<-upa.frontToken.ticker.C
			if upa.frontToken == nil {
				break
			}

			upa.frontToken.ExpiresIn--
			upa.Logger.Debug(fmt.Sprintf("frontToken.ExpiresIn:%d", upa.frontToken.ExpiresIn), upa.Logger.Field("appId", upa.AppId))

			if upa.frontToken.ExpiresIn <= 0 {
				upa.ClearFrontToken()
				break
			}
		}
	}()
}

func (upa *UnionPayApp) GetBackendToken() (a *BackendToken, err error) {

	upa.backendTokenLock.Lock()
	defer func() {
		if err != nil {
			upa.Logger.Error("GetBackendToken", upa.Logger.Field("error", err), upa.Logger.Field("appId", upa.AppId))
		}
		upa.backendTokenLock.Unlock()
	}()

	if upa.backendToken != nil && upa.backendToken.BackendToken != "" {
		upa.Logger.Debug("GetBackendToken from application", upa.Logger.Field("appId", upa.AppId))
		return upa.backendToken, nil
	}

	backendToken, err := upa.Redis.Get(context.Background(), BackendTokenPrefix+upa.AppId).Result()
	ttl, err := upa.Redis.TTL(context.Background(), BackendTokenPrefix+upa.AppId).Result()
	if backendToken != "" && ttl > 0 {
		upa.SetBackendToken(backendToken, int64(ttl/time.Second))
		upa.Logger.Debug("GetBackendToken from redis", upa.Logger.Field("appId", upa.AppId))
		return upa.backendToken, nil
	}

	req := map[string]interface{}{
		"appId": upa.AppId,
	}
	upa.Sign(req)

	resp, err := resty.New().R().
		SetBody(req).
		Post("https://open.95516.com/open/access/1.0/backendToken")

	if err != nil {
		return nil, err
	}

	upa.Logger.Debug("GetBackendToken", upa.Logger.Field("response", resp.Body()), upa.Logger.Field("appId", upa.AppId))

	var res = &BackendTokenRes{}
	if err := json.Unmarshal(resp.Body(), res); err != nil {
		return nil, err
	}

	if res.Resp != "00" {
		if res.Msg != "" {
			return nil, errors.New(fmt.Sprintf("(%s)%s", res.Resp, res.Msg))
		} else {
			return nil, errors.New("request failed")
		}
	}

	expireIn, err := res.Params.ExpiresIn.Int64()
	if err != nil {
		return nil, err
	}

	expireIn = 3600

	upa.SetBackendToken(res.Params.BackendToken, expireIn)

	upa.Redis.Set(context.Background(), BackendTokenPrefix+upa.AppId, res.Params.BackendToken, time.Second*time.Duration(expireIn)).Result()

	upa.Logger.Debug("GetBackendToken from request", upa.Logger.Field("appId", upa.AppId))
	return upa.backendToken, nil
}

func (upa *UnionPayApp) GetFrontToken() (a *FrontToken, err error) {

	upa.frontTokenLock.Lock()
	defer func() {
		if err != nil {
			upa.Logger.Error("GetFrontToken", upa.Logger.Field("error", err), upa.Logger.Field("appId", upa.AppId))
		}
		upa.frontTokenLock.Unlock()
	}()

	if upa.frontToken != nil && upa.frontToken.FrontToken != "" {
		upa.Logger.Debug("GetFrontToken from application", upa.Logger.Field("appId", upa.AppId))
		return upa.frontToken, nil
	}

	frontToken, err := upa.Redis.Get(context.Background(), FrontTokenPrefix+upa.AppId).Result()
	ttl, err := upa.Redis.TTL(context.Background(), FrontTokenPrefix+upa.AppId).Result()
	if frontToken != "" && ttl > 0 {
		upa.SetFrontToken(frontToken, int64(ttl/time.Second))
		upa.Logger.Debug("GetFrontToken from redis", upa.Logger.Field("appId", upa.AppId))
		return upa.frontToken, nil
	}

	req := map[string]interface{}{
		"appId": upa.AppId,
	}
	upa.Sign(req)

	resp, err := resty.New().R().
		SetBody(req).
		Post("https://open.95516.com/open/access/1.0/frontToken")

	if err != nil {
		return nil, err
	}

	upa.Logger.Debug("GetFrontToken", upa.Logger.Field("response", resp.Body()), upa.Logger.Field("appId", upa.AppId))

	var res = &FrontTokenRes{}
	if err := json.Unmarshal(resp.Body(), res); err != nil {
		return nil, err
	}

	if res.Resp != "00" {
		if res.Msg != "" {
			return nil, errors.New(fmt.Sprintf("(%s)%s", res.Resp, res.Msg))
		} else {
			return nil, errors.New("request failed")
		}
	}

	expireIn, err := res.Params.ExpiresIn.Int64()
	if err != nil {
		return nil, err
	}

	expireIn = 3600

	upa.SetFrontToken(res.Params.FrontToken, expireIn)

	upa.Redis.Set(context.Background(), FrontTokenPrefix+upa.AppId, res.Params.FrontToken, time.Second*time.Duration(expireIn)).Result()

	upa.Logger.Debug("GetFrontToken from request", upa.Logger.Field("appId", upa.AppId))
	return upa.frontToken, nil
}
