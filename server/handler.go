package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/abeer-srivastava/redigo/store"
)



type Handler struct{
	store KVStore
}


func storeToHttp(w http.ResponseWriter,err error){
	switch{
		case errors.Is(err,store.ErrEmptyKey):
			http.Error(w,"key is empty",http.StatusBadRequest)
		case errors.Is(err,store.ErrKeyNotFound):
			http.Error(w,"key not found",http.StatusNotFound)
		default:
			http.Error(w,"internal error",http.StatusInternalServerError)
	}
	
}

func (h *Handler) GetKey(w http.ResponseWriter,r *http.Request){
	key:=r.PathValue("key")
	if(key==""){
		storeToHttp(w,store.ErrEmptyKey)
		return
	}
	value,err:=h.store.Get(key)
	if(err!=nil){
		storeToHttp(w,err)
		return
	}
	w.Header().Set("Content-Type","application/octet-stream")
	w.WriteHeader(200)
	w.Write(value)
}

func (h *Handler) SetKey(w http.ResponseWriter,r *http.Request){
	key:=r.PathValue("key")
	if(key==""){
		storeToHttp(w,store.ErrEmptyKey)
		return
	}
	body,err:=io.ReadAll(io.LimitReader(r.Body,1<<20))
	if(err!=nil){
		http.Error(w,"invalid request",http.StatusBadRequest)
		return 
	}
	ttlStr:=r.URL.Query().Get("ttl")
	if(ttlStr!=""){
		ttl,err:=time.ParseDuration(ttlStr)
		if(err!=nil){
			storeToHttp(w,err)
			return 
		}
		err=h.store.SetWithTtl(key,body,ttl)
		if(err!=nil){
			storeToHttp(w,err)
			return 
		}
	}else{
		err:=h.store.Set(key,body)
		if(err!=nil){
			storeToHttp(w,err)
			return 
		}
	}
	w.WriteHeader(200)
}

func (h *Handler) DeleteKey(w http.ResponseWriter,r *http.Request){
	key:=r.PathValue("key")
	if(key==""){
		storeToHttp(w,store.ErrEmptyKey)
		return
	}
	err:=h.store.Delete(key)
	if(err!=nil){
		storeToHttp(w,err)
		return
	}
	w.WriteHeader(200)
}
func (h *Handler) ExistsKey(w http.ResponseWriter,r *http.Request){
	key:=r.PathValue("key")
	if(key==""){
		storeToHttp(w,store.ErrEmptyKey)
		return
	}
	if(h.store.Exists(key)){
		w.WriteHeader(200)
		return
	}
	w.WriteHeader(404)
}

func (h *Handler) ScanKeys(w http.ResponseWriter,r *http.Request){
	start:=r.URL.Query().Get("start")
	end:=r.URL.Query().Get("end")
	if(start=="" && end==""){
		http.Error(w,"start and end query params required",http.StatusBadRequest)
		return
	}
	if(start==""){
		start=end
	}
	if(end==""){
		end=start
	}
	result,err:=h.store.Scan(start,end)
	if(err!=nil){
		storeToHttp(w,err)
		return
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(result)
}
