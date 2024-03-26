package base

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
