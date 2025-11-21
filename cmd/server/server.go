package main

import (
	pb "distributed-auction/grpc"
	"distributed-auction/pkg/node"
	"flag"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "50051", "Port to listen on")
	id := flag.String("id", "1", "Node ID")
	leader := flag.Bool("leader", false, "Is this node the leader?")
	peersRaw := flag.String("peers", "", "Comma-separated peer addresses")
	flag.Parse()

	var peers []string

	if *peersRaw != "" {
		peers = strings.Split(*peersRaw, ",")
	}

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	auctionNode := node.NewNode(*id, peers, *leader)

	pb.RegisterAuctionServiceServer(grpcServer, auctionNode)

	log.Printf("Node %s listening on port %s (Leader: %v)", *id, *port, *leader)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
