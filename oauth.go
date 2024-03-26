package unionPayApp

import (
	"encoding/base64"
	"github.com/forgoer/openssl"
	"github.com/go-tron/crypto/desUtil"
	"github.com/google/go-querystring/query"
)

type Scope string

const (
	ScopeBase   Scope = "upapi_base"
	ScopeMobile Scope = "upapi_mobile"
)

type OAuthCodeReq struct {
	Uri   string `json:"uri"`
	Scope Scope  `json:"scope"`
	State string `json:"state"`
}

type OAuthCodeQuery struct {
	AppId        string `url:"appId"`
	RedirectUri  string `url:"redirectUri"`
	ResponseType string `url:"responseType"`
	Scope        string `url:"scope"`
	State        string `url:"state"`
}

func (upa *UnionPayApp) GetOAuthCode(params *OAuthCodeReq) (string, error) {
	req := OAuthCodeQuery{
		AppId:        upa.AppId,
		RedirectUri:  params.Uri,
		ResponseType: "code",
		Scope:        string(params.Scope),
		State:        params.State,
	}
	if upa.OAuthRedirectUri != "" {
		req.RedirectUri = upa.OAuthRedirectUri + req.RedirectUri
	}
	v, err := query.Values(req)
	if err != nil {
		return "", err
	}
	return "https://open.95516.com/s/open/html/oauth.html?" + v.Encode(), nil
}

type OAuthToken struct {
	AccessToken  string `json:"accessToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	RefreshToken string `json:"refreshToken"`
	OpenId       string `json:"openId"`
	Scope        string `json:"scope"`
}

func (upa *UnionPayApp) GetOAuthToken(code string) (*OAuthToken, error) {
	res, err := upa.Request("OAuthToken", map[string]interface{}{
		"appId":     upa.AppId,
		"code":      code,
		"grantType": "authorization_code",
	}, &OAuthToken{})
	if err != nil {
		return nil, err
	}
	return res.(*OAuthToken), nil
}

type OAuthMobileReq struct {
	OpenId      string `json:"openId"`
	AccessToken string `json:"accessToken"`
}

type OAuthMobile struct {
	Mobile string `json:"mobile"`
}

func (upa *UnionPayApp) GetOAuthMobile(params *OAuthMobileReq) (*OAuthMobile, error) {
	res, err := upa.Request("OAuthMobile", map[string]interface{}{
		"appId":       upa.AppId,
		"accessToken": params.AccessToken,
		"openId":      params.OpenId,
	}, &OAuthMobile{})
	if err != nil {
		return nil, err
	}

	result := res.(*OAuthMobile)
	src, err := base64.StdEncoding.DecodeString(result.Mobile)
	if err != nil {
		return nil, err
	}

	mobile, err := desUtil.Des3ECBDecrypt(src, upa.encryptKeyByte, openssl.PKCS5_PADDING)
	if err != nil {
		return nil, err
	}
	result.Mobile = string(mobile)
	return result, nil
}

func (upa *UnionPayApp) GetOAuthMobileFromCode(code string) (*OAuthMobile, error) {
	res, err := upa.GetOAuthToken(code)
	if err != nil {
		return nil, err
	}
	return upa.GetOAuthMobile(&OAuthMobileReq{
		OpenId:      res.OpenId,
		AccessToken: res.AccessToken,
	})
}
