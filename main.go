package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
)

func main() {

	server := Server{
		DbUrl: "",
	}

	strListenPort := os.Getenv("DYRWOOD_PORT")
	if strListenPort == "" {
		strListenPort = "51900"
	}
	listenPort64, err := strconv.ParseInt(strListenPort, 0, 64)
	if err != nil {
		log.Fatal("error getting port number: ", err)
	}

	server.Port = int(listenPort64)

	err = server.Serve()
	if err != nil {
		log.Fatal(err)
	}

}

type Server struct {
	DbUrl string
	Port  int

	rpcServer  *rpc.Server
	l          net.Listener
	shouldStop bool
	stopped    chan struct{}
}

func (self *Server) Serve() (err error) {
	self.stopped = make(chan struct{}, 1)
	defer func() {
		self.stopped <- struct{}{}
	}()

	db, err := sql.Open("postgres", self.DbUrl)
	if err != nil {
		return
	}
	defer db.Close()

	var refreshDelay float64
	refreshDelayStr := os.Getenv("DYRWOOD_REFRESH_DELAY")
	if refreshDelayStr == "" {
		refreshDelayStr = "5"
	}
	refreshDelay, err = strconv.ParseFloat(refreshDelayStr, 64)
	if err != nil {
		return err
	}

	self.rpcServer = rpc.NewServer()

	rpcHandler := NewDyrwood(db, refreshDelay)
	self.rpcServer.Register(rpcHandler)

	self.l, err = net.Listen("tcp", fmt.Sprintf(":%d", self.Port))
	if err != nil {
		return
	}

	log.Printf("dyrwood listening on %d...", self.Port)

	var conn net.Conn
	for {
		conn, err = self.l.Accept()

		if err != nil {
			if self.shouldStop {
				err = nil
			}
			return
		}

		go self.rpcServer.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}

func (self *Server) Stop() (err error) {
	self.shouldStop = true
	err = self.l.Close()
	if err != nil {
		return err
	}
	_ = <-self.stopped
	return nil
}
