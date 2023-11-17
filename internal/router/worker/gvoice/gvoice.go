// gvoice manages the communication with GVMS.
package gvoice

import (
	"fmt"

	"github.com/kingcobra2468/cot/internal/config"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Setup creates a new GVMS client connection pool based on the connection
// configuration.
func New(c *config.GVMS) GVoiceClient {

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Hostname, c.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}

	return NewGVoiceClient(conn)

}
