package main

import (
	"context"
	"fmt"
	"log"

	ping "github.com/NaddiNadja/peer-to-peer/grpc"
)

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
