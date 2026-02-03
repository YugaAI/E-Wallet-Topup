package external

import (
	"context"
	notificationpb "ewallet-topup/external/proto/notification"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type External struct {
	NotificationClient *NotificationClient
}

// Init client sekali di startup
func NewExternal(grpcAddr string) (*External, error) {
	client, err := NewNotificationClient(grpcAddr)
	if err != nil {
		return nil, err
	}
	return &External{NotificationClient: client}, nil
}

func (e *External) NotifyUserRegistered(userID int64, email, fullName string) error {
	if e.NotificationClient == nil {
		return fmt.Errorf("notification client not initialized")
	}
	return e.NotificationClient.SendNotification(userID, email, fullName)
}

// --- NotificationClient ---
type NotificationClient struct {
	Conn   *grpc.ClientConn
	Client notificationpb.NotificationServiceClient
}

func NewNotificationClient(addr string) (*NotificationClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := notificationpb.NewNotificationServiceClient(conn)
	return &NotificationClient{Conn: conn, Client: client}, nil
}

func (n *NotificationClient) Close() error {
	return n.Conn.Close()
}

func (n *NotificationClient) SendNotification(userID int64, email, fullName string) error {
	req := &notificationpb.SendNotificationRequest{
		Event:    "user_registered",
		UserId:   userID,
		Channels: []string{"email", "push"},
		Payload: &notificationpb.NotificationPayload{
			Email: &notificationpb.EmailPayload{
				To:      email,
				Subject: "Welcome to Ewallet!",
				Body:    "Hi " + fullName + ", thanks for registering!",
			},
			Push: &notificationpb.PushPayload{
				Title: "Welcome!",
				Body:  "Hi " + fullName + ", thanks for registering!",
				Data:  map[string]string{},
			},
		},
	}

	resp, err := n.Client.SendNotification(context.Background(), req)
	if err != nil {
		return fmt.Errorf("grpc send failed: %w", err)
	}

	switch resp.Status {
	case "PENDING", "PROCESSING", "SUCCESS":
		return nil
	default:
		return fmt.Errorf("notification rejected: %s", resp.Status)
	}
	log.Warn().
		Str("status", resp.Status).
		Msg("notification accepted but not finished yet")

	return nil

}
