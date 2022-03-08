package grpc_interceptor

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestStreamServer(t *testing.T) {
	streamServMd := &StreamServerInterceptors{}

	streamServMd.UseMethod("/xx.TestService/TestStream", func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		fmt.Println("part --- 0")
		if val, ok := srv.(string); ok {
			if val == "service" {
				return errors.New("srv is string")
			}
		}
		err := handler(srv, ss)
		fmt.Println("part --- 0 --- end")
		return err
	})

	streamServMd.UseMethod("/xx.TestService/TestStream", func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		fmt.Println("part --- 1")
		err := handler(srv, ss)
		fmt.Println("part --- 1 --- end")
		return err
	}, func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		fmt.Println("part --- 2")
		err := handler(srv, ss)
		fmt.Println("part --- 2 --- end")
		return err
	})

	streamServMd.UseGlobal([]grpc.StreamServerInterceptor{
		func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			fmt.Println("global --- 0")
			err := handler(srv, ss)
			fmt.Println("global --- 0 --- end")
			return err
		},
		func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			fmt.Println("global --- 1")
			err := handler(srv, ss)
			fmt.Println("global --- 1 --- end")
			return err
		},
	})

	streamServMd.UseGlobal([]grpc.StreamServerInterceptor{
		func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			fmt.Println("global --- 2")
			err := handler(srv, ss)
			fmt.Println("global --- 2 --- end")
			return err
		},
	}, "/xx.TestService/TestStream")

	funcs := streamServMd.StreamServerInterceptor()
	err := funcs("service1", &serverStream{}, &grpc.StreamServerInfo{FullMethod: "/xx.TestService/TestStream"}, func(srv interface{}, stream grpc.ServerStream) error {
		fmt.Println("stream handle ...")
		return nil
	})
	fmt.Println(err)
}

type serverStream struct{}

func (s *serverStream) SetHeader(md metadata.MD) error {
	return nil
}

func (s *serverStream) SendHeader(metadata.MD) error {
	return nil
}

func (s *serverStream) SetTrailer(metadata.MD) {
}

func (s *serverStream) Context() context.Context {
	return context.Background()
}

func (s *serverStream) SendMsg(m interface{}) error {
	return nil
}

func (s *serverStream) RecvMsg(m interface{}) error {
	return nil
}
