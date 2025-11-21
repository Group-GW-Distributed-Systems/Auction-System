package main

import (
	"context"
	"distributed-auction/grpc"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	target := flag.String("target", "localhost:50051", "Server address")
	action := flag.String("action", "result", "Action: bid or result")
	amount := flag.Int("amount", 0, "Bid amount")
	id := flag.String("id", "user1", "Bidder ID")
	flag.Parse()

	conn, err := grpc.Dial(*target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := auction.NewAuctionServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if *action == "bid" {
		res, err := client.Bid(ctx, &auction.BidRequest{
			BidderId: *id,
			Amount:   int32(*amount),
		})
		if err != nil {
			log.Fatalf("Error bidding: %v", err)
		}
		fmt.Printf("Bid Response: %s (Reason: %s)\n", res.Status, res.Reason)

	} else if *action == "result" {
		res, err := client.Result(ctx, &auction.ResultRequest{})
		if err != nil {
			log.Fatalf("Error getting result: %v", err)
		}
		fmt.Printf("Result: High Bid: %d, Winner: %s, Over: %v\n",
			res.HighestBid, res.WinnerId, res.IsOver)
	}
}
