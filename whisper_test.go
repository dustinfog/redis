package redis

import (
	"testing"
	"time"
)

func TestWhisperLock(t *testing.T) {
	redis := NewClient(&Options{
		Addr: "127.0.0.1:6379",
	})

	whiperLock := NewWhisperLock(redis, "whisper:lock:test", "123456", 5*time.Second)

	//whiperLock.Acquire()
	cmd := whiperLock.Get("good")
	whiperLock.Release()

	res, err := cmd.Result()
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(res)
	//
	//cmd = whiperLock.Get("whisper:lock:test")
	//res, err = cmd.Result()
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//
	//t.Log(res)
}
