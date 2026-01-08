package internalgrpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	eventpb "github.com/perc400/hw/hw12_13_14_15_calendar/api/event"                     //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/app"                          //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/logger"                       //nolint:depguard
	internalgrpc "github.com/perc400/hw/hw12_13_14_15_calendar/internal/server/grpc"     //nolint:depguard
	memorystorage "github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage/memory" //nolint:depguard
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const bufSize = 1048576

func newBufConnServer(t *testing.T) (*grpc.ClientConn, func()) {
	t.Helper()

	lsn := bufconn.Listen(bufSize)
	logg, err := logger.New("Info")
	require.NoError(t, err)

	strg := memorystorage.New()
	a := app.New(logg, strg)

	server := internalgrpc.NewServer(logg, a)

	go func() {
		if err := server.ServeListener(lsn); err != nil {
			t.Errorf("server exited with error: %v", err)
		}
	}()

	//nolint:revive
	cli, err := grpc.NewClient(
		"127.0.0.1",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lsn.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		cli.Close()
		server.Stop()
	}

	return cli, cleanup
}

func TestGRPCCreateAndListDay(t *testing.T) {
	conn, cleanup := newBufConnServer(t)
	defer cleanup()

	client := eventpb.NewEventServiceClient(conn)

	ctx := context.Background()

	userID := uint64(42)
	date := time.Date(2025, 1, 8, 12, 0, 0, 0, time.UTC)

	_, err := client.CreateEvent(ctx, &eventpb.CreateEventRequest{
		Event: &eventpb.Event{
			Id:                "event-1",
			Title:             "Test event",
			Datetime:          timestamppb.New(date),
			Duration:          int64(time.Hour.Seconds()),
			Description:       "test",
			UserId:            userID,
			NotificationDelay: int64(time.Minute.Seconds()),
		},
	})
	require.NoError(t, err)

	resp, err := client.ListDay(ctx, &eventpb.ListDayRequest{
		UserId: userID,
		Date:   timestamppb.New(date),
	})
	require.NoError(t, err)

	require.Len(t, resp.Events, 1)

	if resp.Events[0].Id != "event-1" {
		t.Fatalf("unexpected event id: %s", resp.Events[0].Id)
	}
}
