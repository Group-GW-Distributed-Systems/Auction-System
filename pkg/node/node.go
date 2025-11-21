package node

import (
	"context"
	pb "distributed-auction/grpc"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	pb.UnimplementedAuctionServiceServer

	mu         sync.Mutex
	id         string
	highestBid int32
	winnerID   string
	isOver     bool
	endTime    time.Time

	peers    []string
	isLeader bool
}

func NewNode(id string, peers []string, isLeader bool) *Node {
	return &Node{
		id:       id,
		peers:    peers,
		isLeader: isLeader,
		endTime:  time.Now().Add(100 * time.Second),
	}
}

func (n *Node) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if time.Now().After(n.endTime) {
		n.isOver = true
	}

	if n.isOver {
		return &pb.BidResponse{Status: pb.BidResponse_FAIL, Reason: "Auction is over"}, nil
	}

	// Strict Leader Check
	if !n.isLeader {
		return &pb.BidResponse{
			Status: pb.BidResponse_FAIL,
			Reason: "I am not the leader. Please connect to the leader node.",
		}, nil
	}

	if req.Amount <= n.highestBid {
		return &pb.BidResponse{
			Status: pb.BidResponse_FAIL,
			Reason: fmt.Sprintf("Bid too low. Current: %d", n.highestBid),
		}, nil
	}

	n.highestBid = req.Amount
	n.winnerID = req.BidderId
	log.Printf("Node %s: New highest bid %d from %s", n.id, n.highestBid, n.winnerID)

	n.replicateToPeers(ctx)

	return &pb.BidResponse{Status: pb.BidResponse_SUCCESS}, nil
}

// Result implements AuctionService.Result
func (n *Node) Result(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if time.Now().After(n.endTime) {
		n.isOver = true
	}

	return &pb.ResultResponse{
		HighestBid: n.highestBid,
		WinnerId:   n.winnerID,
		IsOver:     n.isOver,
	}, nil
}

func (n *Node) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if req.HighestBid > n.highestBid {
		n.highestBid = req.HighestBid
		n.winnerID = req.WinnerId
		log.Printf("Node %s: Synced state. High: %d, Winner: %s", n.id, n.highestBid, n.winnerID)
	}
	return &pb.SyncResponse{Ack: true}, nil
}

func (n *Node) replicateToPeers(ctx context.Context) {
	for _, peer := range n.peers {
		conn, err := grpc.Dial(peer, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to connect to peer %s: %v", peer, err)
			continue
		}
		defer conn.Close()

		client := pb.NewAuctionServiceClient(conn)
		_, err = client.Sync(ctx, &pb.SyncRequest{
			HighestBid: n.highestBid,
			WinnerId:   n.winnerID,
		})
		if err != nil {
			log.Printf("Failed to sync with peer %s: %v", peer, err)
		}
	}
}
