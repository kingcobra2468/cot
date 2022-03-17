package gvoice

import (
	"context"
	"errors"
	"fmt"
	"sync"

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

var (
	gvmsPool *sync.Pool
)

var (
	errGVMSConnectionError = errors.New("unable to connect to gvms")
)

func Init(host string, port int) {
	gvmsPool = &sync.Pool{
		New: func() interface{} {
			conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil
			}

			return NewGVoiceClient(conn)
		},
	}
}

func (l Link) Texts(numMessages uint64) (*[]Text, error) {
	texts := []Text{}
	// fetch client from pool
	client, ok := gvmsPool.Get().(GVoiceClient)
	if !ok {
		return &texts, errGVMSConnectionError
	}

	msgList, err := client.GetContactHistory(context.Background(),
		&FetchContactHistoryRequest{GvoicePhoneNumber: &l.GVoiceNumber, RecipientPhoneNumber: &l.ClientNumber, NumMessages: &numMessages})
	gvmsPool.Put(client)
	if err != nil || !*msgList.Success {
		return &texts, err
	}

	for _, text := range msgList.Messages {
		texts = append(texts, Text{Message: *text.MessageContents, Timestamp: uint32(*text.Timestamp)})
	}

	return &texts, nil
}
