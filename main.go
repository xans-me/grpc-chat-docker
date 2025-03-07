package main

import (
	"log"
	"net"
	"os"
	"sync"

	"github.com/xans-me/grpc-chat-docker/protobuff"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	gLog "google.golang.org/grpc/grpclog"
)

var grpcLog gLog.LoggerV2

func init() {
	grpcLog = gLog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Connection struct {
	stream protobuff.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

type Server struct {
	Connection []*Connection
}

func (s *Server) CreateStream(con *protobuff.Connect, stream protobuff.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		id:     con.User.Id,
		active: true,
		error:  make(chan error),
	}

	s.Connection = append(s.Connection, conn)

	return <-conn.error
}

func (s *Server) BroadcastMessage(_ context.Context, msg *protobuff.Message) (*protobuff.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connection {
		wait.Add(1)

		go func(msg *protobuff.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {
				err := conn.stream.Send(msg)
				grpcLog.Info("Sending message to: ", conn.stream)

				if err != nil {
					grpcLog.Errorf("Error with Stream: %v - Error: %v", conn.stream, err)
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)

	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &protobuff.Close{}, nil
}

func main() {
	var connections []*Connection

	server := &Server{connections}

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("error creating the server %v", err)
	}

	grpcLog.Info("Starting server at port :8080")

	protobuff.RegisterBroadcastServer(grpcServer, server)
	grpcServer.Serve(listener)
}
