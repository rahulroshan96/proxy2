package server

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"net/http/httputil"

	//"github.com/avinetworks/avi-internal/avitest/reverse-proxy/proxy"

	"github.com/elazarl/goproxy"
)

type ProxyConfiguration struct {
	Method         string `json:"method"`
	Path           string `json:"path"`
	Host           string `json:"host"`
	ResponseCode   int32  `json:"responseCode"`
	QueryParameter string `json:"queryParam"`
	PostBody       string `json:"postBody"`
	ResponseData   string `json:"responseData"`
}

type proxyConfig struct {
	InsertID string
	Config   []*ProxyConfiguration
}

type proxyServer struct {
	server *goproxy.ProxyHttpServer
}

func newProxyServer() *proxyServer {
	server := &proxyServer{
		server: goproxy.NewProxyHttpServer(),
	}
	return server
}

func (s *proxyServer) AddHandler(h goproxy.ReqHandler) {
	s.server.OnRequest().Do(h)
}
func (s *proxyServer) AddRespHandler(h goproxy.RespHandler) {
	s.server.OnResponse().Do(h)
}

func (s *proxyServer) Run() {
	s.server.Verbose = true
	s.server.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*$"))).
		HandleConnect(goproxy.AlwaysMitm)
	if err := http.ListenAndServe(":5996", s.server); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

type proxyUpdateHandler struct {
	sharedConfig *sharedConfig
}
type proxyUpdateRespHandler struct {
	sharedConfig *sharedConfig
}

type sharedConfig struct {
	sync.RWMutex
	configuration map[string][]*ProxyConfiguration
}

func (handler *proxyUpdateHandler) Handle(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	handler.sharedConfig.RLock()
	defer handler.sharedConfig.RUnlock()
	reqBody, err := httputil.DumpRequest(req, true)
	if err != nil {
		panic(err)
	}
	for _, configs := range handler.sharedConfig.configuration {
		for _, config := range configs {
			if req.Method == config.Method && req.Host == config.Host {
				if strings.Contains(req.URL.RawQuery, config.QueryParameter) && strings.Contains(string(reqBody), config.PostBody) && strings.Contains(req.URL.Path, config.Path) {
					logrus.Warnf("denying Host: %s , Path: %s, Method: %s", config.Host, config.Path, config.Method)
					return req, goproxy.NewResponse(req,
						goproxy.ContentTypeText, int(config.ResponseCode),
						config.ResponseData)
				}
			}
		}
	}
	logrus.Infof("allowing Host: %s , Path: %s, Method: %s, Query: %s, body: %s", req.Host, req.URL.Path, req.Method, req.URL.RawQuery, reqBody)
	return req, nil
}

// resp *http.Response, ctx *ProxyCtx) *http.Response

func (handler *proxyUpdateRespHandler) Handle(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	//buf := new(strings.Builder)
	//n, _ := io.Copy(buf, resp.Body)
	//logrus.Info(n)
	//logrus.Info("Origional Response Body is %s", buf.String())
	logrus.Info("This is the response")
	req := resp.Request
	handler.sharedConfig.RLock()
	defer handler.sharedConfig.RUnlock()
	reqBody, err := httputil.DumpRequest(req, true)
	if err != nil {
		panic(err)
	}
	for _, configs := range handler.sharedConfig.configuration {
		for _, config := range configs {
			if req.Method == config.Method && req.Host == config.Host {
				if strings.Contains(req.URL.RawQuery, config.QueryParameter) && strings.Contains(string(reqBody), config.PostBody) && strings.Contains(req.URL.Path, config.Path) {
					logrus.Warnf("Response denying Host: %s , Path: %s, Method: %s", config.Host, config.Path, config.Method)
					rr := ioutil.NopCloser(bytes.NewReader([]byte(config.ResponseData)))
					resp.Body = rr
					resp.StatusCode = int(config.ResponseCode)
					return resp
				}
			}
		}
	}
	logrus.Infof("allowing Host: %s , Path: %s, Method: %s, Query: %s, body: %s", req.Host, req.URL.Path, req.Method, req.URL.RawQuery, reqBody)
	return resp
}
