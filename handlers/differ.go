package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BouncyElf/differ/config"
	"github.com/BouncyElf/flow"
	"github.com/gin-gonic/gin"
)

type resp struct {
	status int
	header http.Header
	body   []byte
}

var (
	filelocker *sync.Mutex
	once       sync.Once
)

func init() {
	once.Do(func() {
		filelocker = new(sync.Mutex)
	})
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
		if !compareHeader(origin.header, remote.header) {
			log.Println("****************header not equal*********")
			log.Println("=======================origin header:")
			log.Println(origin.header)
			log.Println("=======================remote header:")
			log.Println(remote.header)
		}
		if !compareBody(origin.body, remote.body) {
			log.Println("****************body not equal*********")
			oct := origin.header.Get("Content-Type")
			rct := remote.header.Get("Content-Type")
			if oct == rct && strings.Contains(oct, "json") {
				filelocker.Lock()
				filename := fmt.Sprintf("json_diff_%v_%v.html", rand.Intn(10), time.Now().Unix())
				filelocker.Unlock()
				ojson, rjson := getJSONContent(origin), getJSONContent(remote)
				if ojson == rjson {
					log.Println("false alarm, after ungzip it's fully match")
					return
				}
				html := generateHTML(ojson, rjson)
				err := os.WriteFile(filename, []byte(html), 0644)
				if err != nil {
					panic(err)
				}
				absPath, err := filepath.Abs(filename)
				if err != nil {
					panic(err)
				}
				log.Printf("json diff link: file://%s\n", absPath)
			} else {
				log.Println("=======================origin resp:")
				log.Println(string(origin.body))
				log.Println("=======================remote resp:")
				log.Println(string(remote.body))
			}
		}
	} else {
		log.Println("fully matched")
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
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	ret := &resp{
		status: res.StatusCode,
		header: res.Header,
	}
	bs, err := io.ReadAll(res.Body)
	defer func() { _ = res.Body.Close() }()
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
	return compareHeader(origin.header, remote.header) &&
		compareBody(origin.body, remote.body)
}

func compareBody(l, r []byte) bool {
	if len(l) != len(r) {
		return false
	}
	for i := range l {
		if l[i] != r[i] {
			return false
		}
	}
	return true
}

func compareHeader(l, r http.Header) bool {
	for k, v := range l {
		if config.Conf.ExcludeHeadersMap[k] {
			continue
		}
		vv := r.Values(k)
		if len(v) != len(vv) {
			return false
		}
		for i := range v {
			if v[i] != vv[i] {
				return false
			}
		}
	}
	return true
}

func getJSONContent(data *resp) string {
	if data == nil {
		return ""
	}
	switch data.header.Get("Content-Encoding") {
	case "gzip":
		reader := bytes.NewReader(data.body)

		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return ""
		}
		defer gzReader.Close()

		result, err := io.ReadAll(gzReader)
		if err != nil {
			return ""
		}

		return string(result)
	default:
		return string(data.body)
	}
}
