package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	pb "github.com/SphericalPotatoInVacuum/soa-grpc-engine/proto_gen/mafia"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

type RoomState int

const (
	WAITING RoomState = iota
	TOWNIE_TURN
	MAFIA_TURN
	COP_TURN
	FINISHED
)

type Room struct {
	mu      sync.Mutex
	players map[string]struct{}
	state   RoomState
	size    int
}

type mafiaServer struct {
	pb.UnimplementedMafiaServer

	mu      sync.RWMutex
	rooms   map[string]*Room // room id to room
	mapping map[string]*Room // username to room
}

func (s *mafiaServer) Join(request *pb.JoinRequest, stream pb.Mafia_JoinServer) error {
	username := request.Username
	if _, ok := s.mapping[username]; ok {
		reason := fmt.Sprintf("Username %s already exists", username)
		stream.Send(&pb.Event{EventBody: &pb.Event_JoinResponse{JoinResponse: &pb.Event_JoinResponseMessage{
			Success: false,
			Reason:  &reason,
		}}})
		return nil
	}
	id := request.GetRoomId()
	if id == "create" {
		// create a new room
	} else {
		if room, ok := s.rooms[request.GetRoomId()]; ok {
			// room exists, try to connect
			room.mu.Lock()
			s.mu.Lock()
			if room.state == WAITING {
				room.players[username] = struct{}{}

			}
		} else {

		}
	}
	return nil
}

func (s *mafiaServer) SendChatMessage(ctx context.Context, chatMessage *pb.SendChatMessageRequest) (*pb.SendChatMessageResponse, error) {
	return nil, nil
}

func (s *mafiaServer) Vote(ctx context.Context, request *pb.VoteRequest) (*pb.VoteResponse, error) {
	return nil, nil
}

func newServer() *mafiaServer {
	s := mafiaServer{
		rooms:   make(map[string]*Room),
		mapping: make(map[string]*Room),
	}
	return &s
}

func main() {
	port := os.Getenv("MESSENGER_SERVER_PORT")
	if port == "" {
		port = "51075"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("listening on port: ", lis.Addr())
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterMafiaServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
