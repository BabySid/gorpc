package util

import (
	"context"
	"fmt"
	"google.golang.org/grpc/peer"
	"net"
)

func GetPeerIPFromGRPC(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("[GetPeerIPFromGRPC] invoke FromContext() failed")
	}
	if pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("[GetPeerIPFromGRPC] peer.Addr is nil")
	}
	return pr.Addr.String(), nil
}
