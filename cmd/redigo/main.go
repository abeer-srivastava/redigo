package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abeer-srivastava/redigo/persistence"
	"github.com/abeer-srivastava/redigo/server"
	"github.com/abeer-srivastava/redigo/store"
	"github.com/abeer-srivastava/redigo/ttl"
)

func main(){
	addr:=flag.String(
		"addr",
		":8080",
		"http listen address",
	)
	walPath:=flag.String(
		"wal",
		"data.wal",
		"wal file path",
	)
	sweep:=flag.Duration(
		"sweep",
		time.Second,
		"ttl sweep interval",
	)
	flag.Parse()
	// setting up the memory stack
	mem:=store.NewMemoryStore()
	wal,err:=persistence.NewWalStore(
		*walPath,
		mem,
	)
	if(err!=nil){
		log.Fatalf("wal init failed %v",err)
	}
	kvStore:=ttl.NewTTLStore(
		wal,
		*sweep,
	)
	// server setup
	srv:=server.NewServer(*addr,kvStore)
	go func (){
		log.Printf("server started on %s",*addr)
		if err:=srv.Start();err!=nil{
			log.Fatalf("server error %v",err)
		}
	}()
	quit:=make(chan os.Signal,1)
	signal.Notify(quit,syscall.SIGINT,syscall.SIGTERM)
	<-quit
	log.Print("server shut down signal received")
	ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()
	if err:=srv.Stop(ctx);err!=nil{
		log.Printf("http server shutdown error:%v",err)
	}
	if err:=kvStore.Close();err!=nil{
		log.Printf("store close error:%v",err)
	}
	log.Print("server shutdown")
}