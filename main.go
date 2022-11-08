package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	ping "github.com/NaddiNadja/peer-to-peer/grpc"
	"google.golang.org/grpc"
)

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	peer := &peer{
		id:          ownPort,
		timesPinged: make(map[int32]int32),
		clients:     make(map[int32]ping.PingClient),
		ctx:         ctx,
	}

	list, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %v: %v", ownPort, err)
	}

	grpcServer := grpc.NewServer()
	ping.RegisterPingServer(grpcServer, peer)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	for i := 0; i < 3; i++ {
		port := int32(5000) + int32(i)
		if port == ownPort {
			continue
		}

		var conn *grpc.ClientConn
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}

		// Defer means: When this function returns, call this method (meaing, one main is done, close connection)
		defer conn.Close()

		//  Create new Client from generated gRPC code from proto
		peer.clients[port] = ping.NewPingClient(conn)
	}

	fmt.Println("Ping? Press enter")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		peer.SendPingToAllPeers()
		fmt.Println("Ping? Press enter")
	}
}

type peer struct {
	ping.UnimplementedPingServer
	id          int32                     // own id
	timesPinged map[int32]int32           // map of peer id to number of times pinged
	clients     map[int32]ping.PingClient // map of peer id to client
	ctx         context.Context           // context
}

// The peer receives a ping from another peer:
func (p *peer) Ping(ctx context.Context, req *ping.Request) (*ping.Reply, error) {
	id := req.Id
	p.timesPinged[id] += 1
	fmt.Printf("Peer %v has been pinged %v times\n", id, p.timesPinged[id])
	return &ping.Reply{Amount: p.timesPinged[id]}, nil
}

// The peer sends a ping to all other peers:
func (p *peer) SendPingToAllPeers() {
	request := &ping.Request{Id: p.id}
	for id, client := range p.clients {
		reply, err := client.Ping(p.ctx, request)
		if err != nil {
			log.Fatalf("Error when sending ping to peer %v: %v", id, err)
		}
		fmt.Printf("Peer %v has been pinged %v times\n", id, reply.Amount)
	}
}
