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

// Text contains the message and the unix-time timestamp on when
// it was received.
type Text struct {
	Message   string
	Timestamp uint64
}

var (
	gvmsPool *sync.Pool
)

var (
	errGVMSConnectionError = errors.New("unable to connect to gvms")
)

// Setup creates a new GVMS client connection pool.
func Setup(host string, port int) {
	gvmsPool = &sync.Pool{
		New: func() interface{} {
			conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil
			}

			return NewGVoiceClient(conn)
		},
	}
}

// Texts fetches the number of messages specified from the message history between
// a client number and an associated gvoice number.
func (l Link) Texts(numMessages uint64) (*[]Text, error) {
	texts := []Text{}
	// fetch client from pool
	client, ok := gvmsPool.Get().(GVoiceClient)
	if !ok {
		return &texts, errGVMSConnectionError
	}

	msgList, err := client.GetContactHistory(context.Background(),
		&FetchContactHistoryRequest{GvoicePhoneNumber: &l.GVoiceNumber,
			RecipientPhoneNumber: &l.ClientNumber, NumMessages: &numMessages})
	gvmsPool.Put(client)
	if err != nil || !*msgList.Success {
		return &texts, err
	}

	// parses each of the messages into a Text instance
	for _, text := range msgList.Messages {
		if *text.Source == false {
			continue
		}
		texts = append(texts, Text{Message: *text.MessageContents, Timestamp: uint64(*text.Timestamp)})
	}

	return &texts, nil
}

func (l Link) SendText(message string) error {
	client, ok := gvmsPool.Get().(GVoiceClient)
	if !ok {
		return errGVMSConnectionError
	}
	resp, err := client.SendSMS(context.Background(),
		&SendSMSRequest{GvoicePhoneNumber: &l.GVoiceNumber, RecipientPhoneNumber: &l.ClientNumber, Message: &message})
	gvmsPool.Put(client)

	if err != nil {
		return err
	}
	if resp.Error != nil {
		return errors.New(*resp.Error)
	}

	return nil
}
