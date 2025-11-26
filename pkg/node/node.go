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
	peers      map[string]pb.AuctionServiceClient
	highestBid int32
	winnerID   string
	isOver     bool
	endTime    time.Time
	isLeader   bool
}

func NewNode(id string, peerAddrs []string, isLeader bool) *Node {
	n := &Node{
		id:       id,
		peers:    make(map[string]pb.AuctionServiceClient),
		endTime:  time.Now().Add(100 * time.Second),
		isLeader: isLeader,
	}

	for _, addr := range peerAddrs {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			n.peers[addr] = pb.NewAuctionServiceClient(conn)
		}
	}
	return n
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

	if !n.isLeader {
		for _, client := range n.peers {
			resp, err := client.Bid(ctx, req)
			if err == nil {
				return resp, nil
			}
		}
		n.isLeader = true
	}

	if req.Amount <= n.highestBid {
		return &pb.BidResponse{
			Status: pb.BidResponse_FAIL,
			Reason: fmt.Sprintf("Bid too low. Current: %d", n.highestBid),
		}, nil
	}

	n.highestBid = req.Amount
	n.winnerID = req.BidderId

	n.syncToPeers(ctx)

	log.Printf("Node %s: New highest bid %d from %s", n.id, n.highestBid, n.winnerID)
	return &pb.BidResponse{Status: pb.BidResponse_SUCCESS}, nil
}

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

	n.isLeader = false

	if req.HighestBid > n.highestBid {
		n.highestBid = req.HighestBid
		n.winnerID = req.WinnerId
	}
	return &pb.SyncResponse{Ack: true}, nil
}

func (n *Node) syncToPeers(ctx context.Context) {
	for _, client := range n.peers {
		client.Sync(ctx, &pb.SyncRequest{
			HighestBid: n.highestBid,
			WinnerId:   n.winnerID,
		})
	}
}
