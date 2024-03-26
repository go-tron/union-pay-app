package unionPayApp

import (
	"github.com/go-tron/logger"
	"github.com/go-tron/redis"
	"testing"
)

var account = UnionPayApp{
	Config: &Config{
		Username:         "h8ex8ug5yvsnbqz9",
		Password:         "f7d1168d49fafa6fb41edbc67a29ce0d",
		BaseUrl:          "https://unionpay-config.eioos.com/config",
		AppId:            "19d8405c3c3e452ab43d68ce2aa305d3",
		Secret:           "22ce65cacc7147189c849ab8ae8ad62a",
		EncryptKey:       "2ab34a731cae9d7629d3868f16e9e3c72ab34a731cae9d76",
		PlanId:           "9b56b70f48ac43e09980916d67abbee5",
		OAuthRedirectUri: "https://weixin.eioos.com/oauth/return?uri=",
		Logger:           logger.NewZap("unionPayApp", "info"),
		Redis: redis.New(&redis.Config{
			Addr:     "127.0.0.1:6379",
			Password: "GBkrIO9bkOcWrdsC",
		}),
	},
}

func TestUnionPayApp_GetOAuthCode(t *testing.T) {
	info, err := account.GetOAuthCode(&OAuthCodeReq{
		Uri:   "https://unionpay-notice.eioos.com",
		Scope: ScopeMobile,
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(info)
}

func TestUnionPayApp_ContractInfo(t *testing.T) {
	info, err := account.ContractInfo(&ContractInfoReq{
		OpenId: "neFzPSlKkkFIlOGWSu5jYMGTSJWxD4zttwlPi3lIxAuj5Cdy1fjBGnXUPBoZo0XE",
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(info)
}

func TestUnionPayApp_ContractRelieve(t *testing.T) {
	res, err := account.ContractRelieve(&ContractRelieveReq{
		OpenId:       "neFzPSlKkkFIlOGWSu5jYMGTSJWxD4zttwlPi3lIxAuj5Cdy1fjBGnXUPBoZo0XE",
		ContractId:   "6218400000084452092",
		ContractCode: "",
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log("res", res)
}

func TestUnionPayApp_PushMessage(t *testing.T) {
	res, err := account.PushMessage(&PushMessageReq{
		OpenId:  "neFzPSlKkkFIlOGWSu5jYMGTSJWxD4zttwlPi3lIxAuj5Cdy1fjBGnXUPBoZo0XE",
		Content: "测试消息2",
		Url:     "https://unionpay-notice.eioos.com",
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log("res", res)
}
