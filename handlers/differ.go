package handlers

import (
	"context"
	"differ/config"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/BouncyElf/flow"
	"github.com/gin-gonic/gin"
)

type resp struct {
	status int
	header http.Header
	body   []byte
}

func Differ(c *gin.Context) {
	var origin, remote *resp
	var oerr, rerr error
	f := flow.NewWithLimit(2).With(func() {
		origin, oerr = call(c, config.Conf.OriginSchemeAndHost)
	})
	if config.Conf.AsyncCall {
		f.With(func() {
			remote, rerr = call(c, config.Conf.RemoteSchemeAndHost)
		})
	} else {
		f.Next(func() {
			remote, rerr = call(c, config.Conf.RemoteSchemeAndHost)
		})
	}
	f.Run()
	if oerr != nil || rerr != nil {
		log.Println("****************something went wrong*********")
		log.Println("=======================origin error:")
		log.Println(oerr)
		log.Println("=======================remote error:")
		log.Println(rerr)
		return
	}
	if ok := compareResult(origin, remote); !ok {
		log.Println("****************not equal*********")
		log.Println("=======================origin resp:")
		log.Println(origin)
		log.Println("=======================remote resp:")
		log.Println(remote)
	}
	modifyResp(c, origin)
}

func modifyResp(c *gin.Context, origin *resp) {
	if origin == nil {
		c.String(200, "something wrong")
		return
	}
	for k, v := range origin.header {
		for _, vv := range v {
			c.Writer.Header().Add(k, vv)
		}
	}
	c.Status(origin.status)
	c.Data(origin.status, origin.header.Get("content-type"), origin.body)
}

func call(c *gin.Context, scheme_and_host string) (*resp, error) {
	if scheme_and_host == "" {
		return nil, fmt.Errorf("invalid uri: %v", scheme_and_host)
	}
	u, _ := url.Parse(scheme_and_host)
	scheme, host := u.Scheme, u.Host
	req := c.Request.Clone(context.Background())
	req.URL = &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     c.Request.URL.Path,
		RawPath:  c.Request.URL.RawPath,
		RawQuery: c.Request.URL.RawQuery,
	}
	req.RequestURI = ""
	req.RemoteAddr = ""
	req.Host = ""
	req.TLS = nil
	req.Close = true
	for k := range req.Header {
		if strings.Contains(strings.ToLower(k), "x-forward") {
			req.Header.Del(k)
		}
	}
	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	ret := &resp{
		status: res.StatusCode,
		header: res.Header,
	}
	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return ret, err
	}
	ret.body = bs
	return ret, nil
}

func compareResult(origin, remote *resp) (res bool) {
	if origin == nil {
		return remote == nil
	}
	if len(origin.header) != len(remote.header) {
		return false
	}
	for k, v := range origin.header {
		vv := remote.header.Values(k)
		if len(v) != len(vv) {
			return false
		}
		for i := range v {
			if v[i] != vv[i] {
				return false
			}
		}
	}
	return string(origin.body) == string(remote.body)
}
