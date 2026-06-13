package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/abeer-srivastava/redigo/store"
)
type KVStore interface{
	store.Store
	SetWithTtl(key string,value []byte,ttl time.Duration) error
}
type Server struct{
	srv *http.Server
}

func NewServer(addr string,store KVStore)*Server{
	h:=&Handler{store: store}
	
	mux:=http.NewServeMux()
	mux.HandleFunc("GET /keys/{key}",h.GetKey)
	mux.HandleFunc("PUT /keys/{key}",h.SetKey)
	mux.HandleFunc("DELETE /keys/{keys}",h.DeleteKey)
	mux.HandleFunc("HEAD /keys/{key}",h.ExistsKey)
	return &Server{
		srv:&http.Server{
			Addr: addr,
			Handler: mux,
		},
	}
}

func (s *Server) Start() error{
	log.Print("starting server at port...",s.srv.Addr)
	return s.srv.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error{
	log.Print("server shutting down...")
	return s.srv.Shutdown(ctx)
}
