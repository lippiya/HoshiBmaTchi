package grpcclient

import (
	"context"
	"fmt"
	"time"

	emailpb "hoshiBmaTchi/back-end/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn *grpc.ClientConn
	cli  emailpb.EmailServiceClient
}

func New(target string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	state := conn.GetState()
	if state != connectivity.Ready {
		for conn.WaitForStateChange(ctx, state) {
			state = conn.GetState()
			if state == connectivity.Ready {
				break
			}
		}
		if state != connectivity.Ready {
			conn.Close()
			return nil, fmt.Errorf("gRPC connection to %s not ready", target)
		}
	}
	return &Client{conn: conn, cli: emailpb.NewEmailServiceClient(conn)}, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) QueueOTP(ctx context.Context, recipient, code string) error {
	ctx2, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := c.cli.QueueOTPEmail(ctx2, &emailpb.OTPRequest{Recipient: recipient, Code: code})
	return err
}

func (c *Client) QueueWelcome(ctx context.Context, recipient string, name string) error {
	ctx2, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := c.cli.QueueWelcomeEmail(ctx2, &emailpb.GenericEmail{
		Recipient: recipient,
		Name: name,
	})
	return err
}
