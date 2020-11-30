package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type server struct {
	proxyServer *proxyServer
	muxServer   *muxServer
}

type muxServer struct {
	sharedReqConfig *sharedConfig
	sharedRspConfig *sharedConfig
}

func (s *muxServer) SetSharedConfig(sharedReqConfig *sharedConfig, sharedRspConfig *sharedConfig) {
	s.sharedReqConfig = sharedReqConfig
	s.sharedRspConfig = sharedRspConfig
}

func NewServer() *server {
	return &server{
		proxyServer: newProxyServer(),
		muxServer:   &muxServer{},
	}
}
func (s *server) getAllConfigs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.muxServer.sharedReqConfig.configuration)
}

func (s *server) createConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	insertID := uuid.New()
	pp := []*ProxyConfiguration{}
	var p ProxyConfiguration
	x := json.NewDecoder(r.Body).Decode(&p)
	pp = append(pp, &p)
	if x != nil {
		panic("Error while receiving request")
	}

	s.muxServer.sharedReqConfig.Lock()
	logrus.Info("inserting config")
	s.muxServer.sharedReqConfig.configuration[insertID.String()] = pp
	s.muxServer.sharedReqConfig.Unlock()
	json.NewEncoder(w).Encode(s.muxServer.sharedReqConfig.configuration)
}

func (s *server) deleteConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	if _, ok := s.muxServer.sharedReqConfig.configuration[params["id"]]; ok {
		s.muxServer.sharedReqConfig.Lock()
		logrus.Info("deleting config")
		delete(s.muxServer.sharedReqConfig.configuration, params["id"])
		s.muxServer.sharedReqConfig.Unlock()
		json.NewEncoder(w).Encode(s.muxServer.sharedReqConfig.configuration)
		return
	}
	json.NewEncoder(w).Encode("")
}
func (s *server) getConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	if _, ok := s.muxServer.sharedReqConfig.configuration[params["id"]]; ok {
		logrus.Info("deleting config")
		json.NewEncoder(w).Encode(s.muxServer.sharedReqConfig.configuration[params["id"]])
		return
	}
	json.NewEncoder(w).Encode("")
}

func (s *server) resetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.muxServer.sharedReqConfig.Lock()
	logrus.Info("resetting config")
	s.muxServer.sharedReqConfig.configuration = make(map[string][]*ProxyConfiguration, 0)
	s.muxServer.sharedReqConfig.Unlock()
}

// Resp Configs
// curl -X POST -d '{"host":"management.azure.com","method":"GET", "path":"virtualMachines", "responseData":"{\"error\":{\"code\":\"AuthorizationFailed\",\"message\":\"Authorization Error\"}}","responseCode":403}' http://localhost:4996/rspConfig
func (s *server) getAllRspConfigs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.muxServer.sharedRspConfig.configuration)
}

func (s *server) createRspConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	insertID := uuid.New()
	pp := []*ProxyConfiguration{}
	var p ProxyConfiguration
	x := json.NewDecoder(r.Body).Decode(&p)
	pp = append(pp, &p)
	if x != nil {
		panic("Error while receiving request")
	}

	s.muxServer.sharedRspConfig.Lock()
	logrus.Info("inserting config")
	s.muxServer.sharedRspConfig.configuration[insertID.String()] = pp
	s.muxServer.sharedRspConfig.Unlock()
	json.NewEncoder(w).Encode(s.muxServer.sharedRspConfig.configuration)
}

func (s *server) deleteRspConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	if _, ok := s.muxServer.sharedRspConfig.configuration[params["id"]]; ok {
		s.muxServer.sharedRspConfig.Lock()
		logrus.Info("deleting config")
		delete(s.muxServer.sharedRspConfig.configuration, params["id"])
		s.muxServer.sharedRspConfig.Unlock()
		json.NewEncoder(w).Encode(s.muxServer.sharedRspConfig.configuration)
		return
	}
	json.NewEncoder(w).Encode("")
}
func (s *server) getRspConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	if _, ok := s.muxServer.sharedRspConfig.configuration[params["id"]]; ok {
		logrus.Info("deleting config")
		json.NewEncoder(w).Encode(s.muxServer.sharedRspConfig.configuration[params["id"]])
		return
	}
	json.NewEncoder(w).Encode("")
}

func (s *server) resetRspConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.muxServer.sharedRspConfig.Lock()
	logrus.Info("resetting config")
	s.muxServer.sharedRspConfig.configuration = make(map[string][]*ProxyConfiguration, 0)
	s.muxServer.sharedRspConfig.Unlock()
}

func (s *server) Run() {
	// Intialize the proxy server
	// then mux server to handle the configuration.
	sharedReqConfig := &sharedConfig{
		configuration: make(map[string][]*ProxyConfiguration, 0),
	}
	sharedRspConfig := &sharedConfig{
		configuration: make(map[string][]*ProxyConfiguration, 0),
	}
	handler := &proxyUpdateHandler{
		sharedConfig: sharedReqConfig,
	}
	handlerResp := &proxyUpdateRespHandler{
		sharedConfig: sharedRspConfig,
	}
	s.proxyServer.AddHandler(handler)
	s.proxyServer.AddRespHandler(handlerResp)

	s.muxServer.SetSharedConfig(sharedReqConfig, sharedRspConfig)
	// run the proxy server

	go s.proxyServer.Run()
	router := mux.NewRouter()
	router.HandleFunc("/reqConfig", s.getAllConfigs).Methods("GET")
	router.HandleFunc("/reqConfig", s.createConfig).Methods("POST")
	router.HandleFunc("/reqConfig", s.resetConfig).Methods("DELETE")
	router.HandleFunc("/reqConfig/{id}", s.deleteConfig).Methods("DELETE")
	router.HandleFunc("/reqConfig/{id}", s.getConfig).Methods("GET")

	router.HandleFunc("/rspConfig", s.getAllRspConfigs).Methods("GET")
	router.HandleFunc("/rspConfig", s.createRspConfig).Methods("POST")
	router.HandleFunc("/rspConfig", s.resetRspConfig).Methods("DELETE")
	router.HandleFunc("/rspConfig/{id}", s.deleteRspConfig).Methods("DELETE")
	router.HandleFunc("/rspConfig/{id}", s.getRspConfig).Methods("GET")
	http.ListenAndServe(":4996", router)
}
