package internalgrpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func loggingUnaryInterceptor(logger Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		ip := "unknown"
		p, ok := peer.FromContext(ctx)
		if ok {
			ip = p.Addr.String()
		}

		userAgent := "-"
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if len(md.Get("user-agent")) > 0 {
				userAgent = md.Get("user-agent")[0]
			}
		}

		resp, err := handler(ctx, req)
		if err != nil {
			logger.Info("Error " + err.Error())
		}

		stCode := status.Code(err)

		duration := time.Since(start)

		logger.Info(fmt.Sprintf(
			"%s [%s] %s %s %v %s",
			ip,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			info.FullMethod,
			stCode.String(),
			duration,
			userAgent,
		))

		return resp, err
	}
}
