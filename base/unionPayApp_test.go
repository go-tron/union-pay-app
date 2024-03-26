package base

import (
	localTime "github.com/go-tron/local-time"
	"github.com/go-tron/logger"
	"github.com/go-tron/redis"
	"sync"
	"testing"
	"time"
)

var base = UnionPayApp{
	Config: &Config{
		AppId:      "19d8405c3c3e452ab43d68ce2aa305d3",
		Secret:     "22ce65cacc7147189c849ab8ae8ad62a",
		EncryptKey: "2ab34a731cae9d7629d3868f16e9e3c72ab34a731cae9d76",
		Redis: redis.New(&redis.Config{
			Addr:     "127.0.0.1:6379",
			Password: "GBkrIO9bkOcWrdsC",
		}),
		Logger: logger.NewZap("unionPayApp", "info"),
	},
}

func TestGetBackendToken(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	var i = 0
	for i < 1000 {
		i++
		go func() {
			result, err := base.GetBackendToken()
			if err != nil {
				t.Fatal(err)
			}
			t.Log(localTime.Now(), "result", result)
		}()

	}
	wg.Wait()
}

func TestGetFrontToken(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	result, err := base.GetFrontToken()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("result", result)

	time.Sleep(time.Second * 6)
	base.frontToken = nil

	time.Sleep(time.Second * 6)
	result, err = base.GetFrontToken()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("result", result)

	wg.Wait()
}
