package unionPayApp

type PushMessageReq struct {
	OpenId  string `json:"openId"`
	Content string `json:"content"`
	Url     string `json:"url"`
}

func (upa *UnionPayApp) PushMessage(params *PushMessageReq) (map[string]interface{}, error) {
	res, err := upa.Request("PushMessage", map[string]interface{}{
		"appId":   upa.AppId,
		"openId":  params.OpenId,
		"content": params.Content,
		"url":     params.Url,
	}, &map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	return *res.(*map[string]interface{}), nil
}
