package unionPayApp

type accounts interface {
	GetAccountById(string) (*UnionPayApp, error)
}

type Accounts struct {
	Accounts accounts
}

func (u *Accounts) GetBackendToken(appId string) (*BackendToken, error) {
	account, err := u.Accounts.GetAccountById(appId)
	if err != nil {
		return nil, err
	}
	return account.GetBackendToken()
}

func (u *Accounts) GetFrontToken(appId string) (*FrontToken, error) {
	account, err := u.Accounts.GetAccountById(appId)
	if err != nil {
		return nil, err
	}
	return account.GetFrontToken()
}

func (u *Accounts) GetJsApiConfig(appId string, url string) (*JsApiConfig, error) {
	account, err := u.Accounts.GetAccountById(appId)
	if err != nil {
		return nil, err
	}
	return account.GetJsApiConfig(url)
}

func (u *Accounts) GetOAuthCode(appId string, params *OAuthCodeReq) (string, error) {
	account, err := u.Accounts.GetAccountById(appId)
	if err != nil {
		return "", err
	}
	return account.GetOAuthCode(params)
}

func (u *Accounts) GetOAuthToken(appId string, code string) (*OAuthToken, error) {
	account, err := u.Accounts.GetAccountById(appId)
	if err != nil {
		return nil, err
	}
	return account.GetOAuthToken(code)
}

func (u *Accounts) GetOAuthMobile(appId string, params *OAuthMobileReq) (*OAuthMobile, error) {
	account, err := u.Accounts.GetAccountById(appId)
	if err != nil {
		return nil, err
	}
	return account.GetOAuthMobile(params)
}

func (u *Accounts) GetOAuthMobileFromCode(appId string, code string) (*OAuthMobile, error) {
	account, err := u.Accounts.GetAccountById(appId)
	if err != nil {
		return nil, err
	}
	return account.GetOAuthMobileFromCode(code)
}

func (u *Accounts) ContractCode(appId string, params *ContractCodeReq) (string, error) {
	account, err := u.Accounts.GetAccountById(appId)
	if err != nil {
		return "", err
	}
	return account.ContractCode(params)
}

func (u *Accounts) ContractApply(appId string, params *ContractApplyReq) (*ContractApply, error) {
	account, err := u.Accounts.GetAccountById(appId)
	if err != nil {
		return nil, err
	}
	return account.ContractApply(params)
}

func (u *Accounts) ContractRelieve(appId string, params *ContractRelieveReq) (*ContractRelieve, error) {
	account, err := u.Accounts.GetAccountById(appId)
	if err != nil {
		return nil, err
	}
	return account.ContractRelieve(params)
}

func (u *Accounts) ContractInfo(appId string, params *ContractInfoReq) (*ContractInfo, error) {
	account, err := u.Accounts.GetAccountById(appId)
	if err != nil {
		return nil, err
	}
	return account.ContractInfo(params)
}
