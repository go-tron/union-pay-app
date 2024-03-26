package unionPayApp

import (
	"github.com/google/go-querystring/query"
)

const (
	ContractStatusNotOpen  = "0"
	ContractStatusOpened   = "1"
	ContractStatusRelieved = "3"
)

type ContractCodeReq struct {
	Uri   string `json:"uri"`
	State string `json:"state"`
}

type ContractCodeQuery struct {
	AppId        string `url:"appId"`
	RedirectUri  string `url:"redirectUri"`
	ResponseType string `url:"responseType"`
	Scope        string `url:"scope"`
	State        string `url:"state"`
	PlanId       string `url:"planId"`
}

func (upa *UnionPayApp) ContractCode(params *ContractCodeReq) (string, error) {

	req := ContractCodeQuery{
		AppId:        upa.AppId,
		RedirectUri:  params.Uri,
		ResponseType: "code",
		Scope:        "upapi_contract",
		State:        params.State,
		PlanId:       upa.PlanId,
	}
	if upa.OAuthRedirectUri != "" {
		req.RedirectUri = upa.OAuthRedirectUri + req.RedirectUri
	}
	v, err := query.Values(req)
	if err != nil {
		return "", err
	}

	return "https://open.95516.com/s/open/noPwd/html/open.html?" + v.Encode(), nil
}

type ContractApplyReq struct {
	OpenId       string `json:"openId"`
	AccessToken  string `json:"accessToken"`
	ContractCode string `json:"contractCode"`
}

type ContractApply struct {
	ContractCode string `json:"contractCode"`
	ContractId   string `json:"contractId"`
	OperateTime  string `json:"operateTime"`
}

func (upa *UnionPayApp) ContractApply(params *ContractApplyReq) (*ContractApply, error) {
	res, err := upa.Request("ContractApply", map[string]interface{}{
		"appId":        upa.AppId,
		"accessToken":  params.AccessToken,
		"openId":       params.OpenId,
		"planId":       upa.PlanId,
		"contractCode": params.ContractCode,
	}, &ContractApply{})
	if err != nil {
		return nil, err
	}
	return res.(*ContractApply), nil
}

type ContractRelieveReq struct {
	OpenId       string `json:"openId"`
	ContractId   string `json:"contractId"`
	ContractCode string `json:"contractCode"`
}

type ContractRelieve struct {
	ContractCode string `json:"contractCode"`
	OperateTime  string `json:"operateTime"`
}

func (upa *UnionPayApp) ContractRelieve(params *ContractRelieveReq) (*ContractRelieve, error) {
	res, err := upa.Request("ContractRelieve", map[string]interface{}{
		"appId":        upa.AppId,
		"openId":       params.OpenId,
		"planId":       upa.PlanId,
		"contractId":   params.ContractId,
		"contractCode": params.ContractCode,
	}, &ContractRelieve{})
	if err != nil {
		return nil, err
	}
	return res.(*ContractRelieve), nil
}

type ContractInfoReq struct {
	OpenId string `json:"openId"`
}

type ContractInfo struct {
	ContractId     string `json:"contractId"`
	ContractStatus string `json:"contractStatus"`
}

func (upa *UnionPayApp) ContractInfo(params *ContractInfoReq) (*ContractInfo, error) {
	res, err := upa.Request("ContractInfo", map[string]interface{}{
		"appId":  upa.AppId,
		"openId": params.OpenId,
		"planId": upa.PlanId,
	}, &ContractInfo{})
	if err != nil {
		return nil, err
	}
	return res.(*ContractInfo), nil
}
