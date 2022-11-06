package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"

	ping "github.com/NaddiNadja/peer-to-peer/grpc"
)

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := &peer{
		id:          ownPort,
		timesPinged: make(map[int32]int32),
		ctx:         ctx,
	}

	ping.RegisterPingServer(grpcServer, p)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	p.clients = createClients(ownPort)

	fmt.Println("Ping? Press enter")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		p.SendPingToAllPeers()
		fmt.Println("Ping? Press enter")
	}
}

func createClients(ownPort int32) map[int32]ping.PingClient {
	clients := make(map[int32]ping.PingClient)
	for i := int32(5000); i < int32(5003); i++ {
		if i != ownPort {
			address := fmt.Sprintf("localhost:%d", i)
			log.Printf("Node%d dialing port: %v", ownPort-5000, address)
			conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			clients[i] = ping.NewPingClient(conn)
		}
	}
	return clients
}
