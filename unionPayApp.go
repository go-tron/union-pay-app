package unionPayApp

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/go-tron/config"
	"github.com/go-tron/logger"
	"github.com/go-tron/redis"
	"github.com/go-tron/union-pay-app/base"
	"sync"
	"time"
)

func NewWithConfig(c *config.Config, client *redis.Redis) *UnionPayApp {
	return New(&Config{
		Username:         c.GetString("application.id"),
		Password:         c.GetString("application.secret"),
		BaseUrl:          c.GetString("unionPayApp.baseUrl"),
		AppId:            c.GetString("unionPayApp.appId"),
		Secret:           c.GetString("unionPayApp.secret"),
		EncryptKey:       c.GetString("unionPayApp.encryptKey"),
		OAuthRedirectUri: c.GetString("unionPayApp.oAuthRedirectUri"),
		Redis:            client,
		Logger:           logger.NewZapWithConfig(c, "unionPayApp", "info"),
	})
}

func New(c *Config) *UnionPayApp {

	if c == nil {
		panic("config 必须设置")
	}
	if c.Username == "" {
		panic("Username 必须设置")
	}
	if c.Password == "" {
		panic("Password 必须设置")
	}
	if c.BaseUrl == "" {
		panic("BaseUrl 必须设置")
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

type UnionPayApp struct {
	*Config
	lock         sync.Mutex
	backendToken *BackendToken `json:"backendToken"`
	frontToken   *FrontToken   `json:"frontToken"`
}

type Config struct {
	Username         string        `json:"username"`
	Password         string        `json:"password"`
	BaseUrl          string        `json:"baseUrl"`
	AppId            string        `json:"appId"`
	Secret           string        `json:"secret"`
	EncryptKey       string        `json:"encryptKey"`
	PlanId           string        `json:"planId"`
	encryptKeyByte   []byte        `json:"encryptKeyByte"`
	OAuthRedirectUri string        `json:"oAuthRedirectUri"`
	Logger           logger.Logger `json:"logger"`
	Redis            *redis.Redis  `json:"redis"`
}

type BackendTokenRes struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Data    BackendToken `json:"data"`
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

			if upa.backendToken.ExpiresIn == 0 {
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

			if upa.frontToken.ExpiresIn == 0 {
				upa.ClearFrontToken()
				break
			}
		}
	}()
}

func (upa *UnionPayApp) GetBackendToken() (a *BackendToken, err error) {

	upa.lock.Lock()
	defer func() {
		if err != nil {
			upa.Logger.Error("GetBackendToken", upa.Logger.Field("error", err), upa.Logger.Field("appId", upa.AppId))
		}
		upa.lock.Unlock()
	}()

	if upa.backendToken != nil && upa.backendToken.BackendToken != "" {
		upa.Logger.Debug("GetBackendToken from application", upa.Logger.Field("appId", upa.AppId))
		return upa.backendToken, nil
	}

	backendToken, err := upa.Redis.Get(context.Background(), base.BackendTokenPrefix+upa.AppId).Result()
	ttl, err := upa.Redis.TTL(context.Background(), base.BackendTokenPrefix+upa.AppId).Result()
	if backendToken != "" && ttl > 0 {
		upa.SetBackendToken(backendToken, int64(ttl/time.Second))
		upa.Logger.Debug("GetBackendToken from redis", upa.Logger.Field("appId", upa.AppId))
		return upa.backendToken, nil
	}

	resp, err := resty.New().R().
		SetBody(map[string]string{
			"appId":  upa.AppId,
			"secret": upa.Secret,
		}).
		SetBasicAuth(upa.Username, upa.Password).
		Post(upa.BaseUrl + "/backendToken")
	if err != nil {
		return nil, err
	}

	upa.Logger.Debug("GetBackendToken", upa.Logger.Field("response", resp.Body()), upa.Logger.Field("appId", upa.AppId))

	var res = &BackendTokenRes{}
	if err := json.Unmarshal(resp.Body(), res); err != nil {
		return nil, err
	}

	if res.Code != "00" {
		if res.Message != "" {
			return nil, errors.New(fmt.Sprintf("(%s)%s", res.Code, res.Message))
		} else {
			return nil, errors.New("getBackendTokenError")
		}
	}

	upa.SetBackendToken(res.Data.BackendToken, res.Data.ExpiresIn)
	upa.Logger.Debug("GetBackendToken from request", upa.Logger.Field("appId", upa.AppId))

	return upa.backendToken, nil
}

type FrontTokenRes struct {
	Code    string     `json:"code"`
	Message string     `json:"message"`
	Data    FrontToken `json:"data"`
}

func (upa *UnionPayApp) GetFrontToken() (j *FrontToken, err error) {

	upa.lock.Lock()
	defer func() {
		if err != nil {
			upa.Logger.Error("GetFrontToken", upa.Logger.Field("error", err), upa.Logger.Field("appId", upa.AppId))
		}
		upa.lock.Unlock()
	}()

	if upa.frontToken != nil && upa.frontToken.FrontToken != "" {
		upa.Logger.Debug("GetFrontToken from application", upa.Logger.Field("appId", upa.AppId))
		return upa.frontToken, nil
	}

	frontToken, err := upa.Redis.Get(context.Background(), base.FrontTokenPrefix+upa.AppId).Result()
	ttl, err := upa.Redis.TTL(context.Background(), base.FrontTokenPrefix+upa.AppId).Result()
	if frontToken != "" && ttl > 0 {
		upa.SetFrontToken(frontToken, int64(ttl/time.Second))
		upa.Logger.Debug("GetFrontToken from redis", upa.Logger.Field("appId", upa.AppId))
		return upa.frontToken, nil
	}

	resp, err := resty.New().R().
		SetBody(map[string]string{
			"appId":  upa.AppId,
			"secret": upa.Secret,
		}).
		SetBasicAuth(upa.Username, upa.Password).
		Post(upa.BaseUrl + "/frontToken")
	if err != nil {
		return nil, err
	}
	upa.Logger.Debug("GetFrontToken", upa.Logger.Field("response", resp.Body()), upa.Logger.Field("appId", upa.AppId))

	var res = &FrontTokenRes{}
	if err := json.Unmarshal(resp.Body(), res); err != nil {
		return nil, err
	}

	if res.Code != "00" {
		if res.Message != "" {
			return nil, errors.New(fmt.Sprintf("(%s)%s", res.Code, res.Message))
		} else {
			return nil, errors.New("getFrontTokenError")
		}
	}

	upa.SetFrontToken(res.Data.FrontToken, res.Data.ExpiresIn)
	upa.Logger.Debug("GetFrontToken from request", upa.Logger.Field("appId", upa.AppId))

	return upa.frontToken, nil
}
