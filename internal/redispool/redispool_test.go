package redispool

import (
	"flag"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/log/logtest"
)

func TestSchemeMatcher(t *testing.T) {
	tests := []struct {
		urlMaybe  string
		hasScheme bool
	}{
		{"redis://foo.com", true},
		{"https://foo.com", true},
		{"redis://:password@foo.com/0", true},
		{"redis://foo.com/0?password=foo", true},
		{"foo:1234", false},
	}
	for _, test := range tests {
		hasScheme := schemeMatcher.MatchString(test.urlMaybe)
		if hasScheme != test.hasScheme {
			t.Errorf("for string %q, exp != got: %v != %v", test.urlMaybe, test.hasScheme, hasScheme)
		}
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	logtest.Init(m)
	os.Exit(m.Run())
}

func TestDeleteAllKeysWithPrefix(t *testing.T) {
	t.Helper()

	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "127.0.0.1:6379")
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	c := pool.Get()
	defer c.Close()

	// If we are not on CI, skip the test if our redis connection fails.
	if os.Getenv("CI") == "" {
		_, err := c.Do("PING")
		if err != nil {
			t.Skip("could not connect to redis", err)
		}
	}

	kv := RedisKeyValue(pool)
	var aKeys, bKeys []string
	var key string
	for i := range 10 {
		if i%2 == 0 {
			key = "a:" + strconv.Itoa(i)
			aKeys = append(aKeys, key)
		} else {
			key = "b:" + strconv.Itoa(i)
			bKeys = append(bKeys, key)
		}

		if err := kv.Set(key, []byte(strconv.Itoa(i))); err != nil {
			t.Fatalf("could not set key %s: %v", key, err)
		}
	}

	if err := DeleteAllKeysWithPrefix(kv, "a"); err != nil {
		t.Fatal(err)
	}

	getMulti := func(keys ...string) []string {
		t.Helper()
		var vals []string
		for _, k := range keys {
			v, _ := kv.Get(k).String()
			vals = append(vals, v)
		}
		return vals
	}

	vals := getMulti(aKeys...)
	if got, exp := vals, []string{"", "", "", "", ""}; !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}

	vals = getMulti(bKeys...)
	if got, exp := vals, []string{"1", "3", "5", "7", "9"}; !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
}
