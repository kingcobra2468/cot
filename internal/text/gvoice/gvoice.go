package gvoice

import (
	"context"
	"fmt"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Link binds a given GVoice number with a client number for
// fetching future commands.
type Link struct {
	GVoiceNumber string
	ClientNumber string
}

type Text struct {
	Message   string
	Timestamp uint32
}

var client GVoiceClient

func Init(host string, port int) error {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	client = NewGVoiceClient(conn)
	return nil
}

func (l Link) Texts() []string {
	client.GetContactHistory(context.Background(), &FetchContactHistoryRequest{})
	return []string{"test"}
}
