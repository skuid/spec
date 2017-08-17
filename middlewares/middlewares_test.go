package middlewares

import (
	"net/http"
	"testing"
)

func TestGetRemoteAddr(t *testing.T) {

	cases := []struct {
		name       string
		header     map[string]string
		remoteAddr string
		want       string
	}{
		{
			"getRemoteAddr consumes X-Real-IP",
			map[string]string{
				"X-Real-IP":       "11.22.33.44:1234",
				"X-Forwarded-For": "11.22.33.44:1234",
			},
			"11.22.33.44:1234",
			"11.22.33.44",
		},
		{
			"getRemoteAddr consumes X-Forwarded-For",
			map[string]string{
				"X-Forwarded-For": "11.22.33.44:1234",
			},
			"11.22.33.44:1234",
			"11.22.33.44",
		},
		{
			"getRemoteAddr strips port",
			map[string]string{
				"X-Real-IP":       "11.22.33.44:1234",
				"X-Forwarded-For": "11.22.33.44:1234",
			},
			"11.22.33.44:12312",
			"11.22.33.44",
		},
		{
			"getRemoteAddr handles no header",
			map[string]string{},
			"11.22.33.44:2000",
			"11.22.33.44",
		},
		{
			"getRemoteAddr handles no port",
			map[string]string{},
			"11.22.33.44",
			"11.22.33.44",
		},
		{
			"getRemoteAddr handles ipv6",
			map[string]string{},
			"[::]:1234",
			"[::]",
		},
	}
	for _, c := range cases {
		request, err := http.NewRequest("GET", "http://localhost", nil)
		if err != nil {
			t.Errorf("Failed %s: Unable to create a new http.Request{}", c.name)
		}
		request.RemoteAddr = c.remoteAddr
		for header, value := range c.header {
			request.Header.Set(header, value)
		}
		if got := getRemoteAddr(request); got != c.want {
			t.Errorf("Failed %s: getRemoteAddr() Expected: %v, got: %v", c.name, c.want, got)
		}
	}
}
