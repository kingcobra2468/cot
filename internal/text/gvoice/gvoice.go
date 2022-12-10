// gvoice manages the communication with GVMS.
package gvoice

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/kingcobra2468/cot/internal/config"
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
	// gRPC client pool for communicating with GVMS
	gvmsPool *sync.Pool
)

var (
	errGVMSConnectionError = errors.New("unable to connect to gvms")
)

// Setup creates a new GVMS client connection pool based on the connection
// configuration.
func Setup(c *config.GVMS) {
	gvmsPool = &sync.Pool{
		New: func() interface{} {
			conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Hostname, c.Port),
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
	// fetch a client from pool
	client, ok := gvmsPool.Get().(GVoiceClient)
	if !ok {
		return &texts, errGVMSConnectionError
	}

	// extract a list of messages which contain at most numMessages messages
	msgList, err := client.GetContactHistory(context.Background(),
		&FetchContactHistoryRequest{GvoicePhoneNumber: &l.GVoiceNumber,
			RecipientPhoneNumber: &l.ClientNumber, NumMessages: &numMessages})
	gvmsPool.Put(client)
	if err != nil || !*msgList.Success {
		return &texts, errors.New(*msgList.Error)
	}

	// creates a Text instance for each raw message
	for _, text := range msgList.Messages {
		if !*text.Source {
			continue
		}
		texts = append(texts, Text{Message: *text.MessageContents, Timestamp: uint64(*text.Timestamp)})
	}

	return &texts, nil
}

// SendText sends a text message to the setup recipient.
func (l Link) SendText(message string) error {
	client, ok := gvmsPool.Get().(GVoiceClient)
	if !ok {
		return errGVMSConnectionError
	}

	msg := &SendSMSRequest{GvoicePhoneNumber: &l.GVoiceNumber, RecipientPhoneNumber: &l.ClientNumber, Message: &message}
	resp, err := client.SendSMS(context.Background(), msg)
	gvmsPool.Put(client)

	if err != nil {
		return err
	}
	if resp.Error != nil {
		return errors.New(*resp.Error)
	}

	return nil
}
