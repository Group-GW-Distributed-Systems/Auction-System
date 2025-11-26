# Auction-System

## How to Run
An example to start the program is to open three separate terminals to simulate the nodes and the client.

### 1. Start the Leader Node
**Terminal 1:**
```bash
go run cmd/server/server.go -id=1 -port=50051 -leader=true
```

**Terminal 2:**
```bash
go run cmd/server/server.go -id=2 -port=50052 -peers=localhost:50051
```

**Terminal 3:**

Place a bit:
```bash
go run cmd/client/client.go -target=localhost:50051 -action=bid -amount=500 -id=User1
```

Check result: 
```bash
go run cmd/client/client.go -target=localhost:50051 -action=result
```

Test Failover:

1. Stop the Leader (Press Ctrl+C in Terminal 1).

2. Place a bid on the Follower node:
```bash
go run cmd/client.go -target=localhost:50052 -action=bid -amount=600 -id=User2
```

This example simulates a node failure. Thanks to the replication strategy, the backup server automatically takes over, ensuring that no data is lost.
