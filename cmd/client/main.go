package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	pb "github.com/SphericalPotatoInVacuum/soa-grpc-engine/proto_gen/mafia"
	"google.golang.org/grpc"
)

type ClientState int32

const helpMessage = `Usage:
/help: Print this message
/setUsername [username]: Set username to [username]

Connected to server:
/joinRoom [roomId]: join room with id [roomId]
/createRoom: create a new room

During your turn:
/vote [username]: vote for username when it is your turn.
/endTurn: End your turn.
`

type Client struct {
	username string
	stub     pb.MafiaClient
}

func (c *Client) handleServer(stream pb.Mafia_JoinClient, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			fmt.Printf("Disconnected from server!\n")
			break
		}
		if err != nil {
			log.Fatalf("%v", err)
		}
		switch body := event.EventBody.(type) {
		case *pb.Event_JoinResponse:
			if !body.JoinResponse.Success {
				fmt.Printf("Couldn't join room. Reason: \"%s\"\n", body.JoinResponse.GetReason())
			}
		case *pb.Event_ConnectionEvent:
			if body.ConnectionEvent.Connected {
				fmt.Printf("User %s connected", body.ConnectionEvent.Username)
			} else {
				fmt.Printf("User %s disconnected", body.ConnectionEvent.Username)
			}
		case *pb.Event_ChatEvent:
			fmt.Printf("%s: \"%s\"", body.ChatEvent.Username, body.ChatEvent.Text)
		case *pb.Event_VoteEvent:
			fmt.Printf("%s voted for %s\n", body.VoteEvent.Voter, body.VoteEvent.Target)
		case *pb.Event_EndTurnEvent:
			fmt.Printf("%s voted for turn to end\n", body.EndTurnEvent.Voter)
		default:
			fmt.Printf("Received an unrecognized event: %v\n", body)
		}
	}
}

func (c *Client) handleCommand(cmd string, args []string, wg *sync.WaitGroup) (bool, error) {
	switch cmd {
	case "/help":
		fmt.Printf(helpMessage)
	case "/setUsername":
		if len(args) < 1 {
			fmt.Printf("ERROR: not enough arguments\n")
			break
		}
		c.username = args[0]
		log.Printf("Set username to %s\n", c.username)
	case "/joinRoom":
		if len(args) < 1 {
			fmt.Printf("ERROR: not enough arguments\n")
			break
		}
		stream, err := c.stub.Join(context.Background(), &pb.JoinRequest{Username: c.username, RoomId: args[0]})
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			break
		}
		wg.Add(1)
		go c.handleServer(stream, wg)
	case "/createRoom":
		stream, err := c.stub.Join(context.Background(), &pb.JoinRequest{Username: c.username, RoomId: "create"})
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			break
		}
		wg.Add(1)
		go c.handleServer(stream, wg)
	case "/vote":
		if len(args) < 1 {
			fmt.Printf("ERROR: not enough arguments\n")
			break
		}
		response, err := c.stub.Vote(context.Background(), &pb.VoteRequest{Target: args[1]})
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			break
		}
		if !response.Success {
			fmt.Printf("ERROR: %s\n", response.GetReason())
		}
	case "/exit":
		return false, nil
	default:
		fmt.Printf("Command %s is not recognized\n", cmd)
	}
	return true, nil
}

func (c *Client) Cli(wg *sync.WaitGroup) error {
	defer wg.Done()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		input := scanner.Text()
		if len(input) == 0 {
			continue
		}
		if input[0] == '/' {
			fields := strings.Fields(input)
			cmd, args := fields[0], fields[1:]
			cont, err := c.handleCommand(cmd, args, wg)
			if err != nil {
				return err
			}
			if cont {
				continue
			} else {
				break
			}
		}
	}
	return nil
}

func NewClient(username string, conn *grpc.ClientConn) (*Client, error) {
	c := &Client{
		username: username,
		stub:     pb.NewMafiaClient(conn),
	}
	return c, nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Enter server address: ")
	var conn *grpc.ClientConn
	var err error
	for scanner.Scan() {
		conn, err = grpc.Dial(scanner.Text(), grpc.WithInsecure(), grpc.WithTimeout(time.Second))
		if err != nil {
			continue
		}
		break
	}
	defer conn.Close()
	log.Printf("Using server address: %s", scanner.Text())

	fmt.Printf("Enter username: ")
	scanner.Scan()
	c, err := NewClient(scanner.Text(), conn)
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go c.Cli(wg)
	wg.Wait()
}
