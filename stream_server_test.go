package grpc_interceptor

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestStreamServer(t *testing.T) {
	interceptors := &StreamServerInterceptors{}
	method := "/xx.TestService/TestStream"

	interceptors.Add(streamServerInterceptorFunc(func(exec func() error) error {
		fmt.Println("0")
		err := exec()
		fmt.Println("0")
		return err
	}))

	interceptors.Add(streamServerInterceptorFunc(func(exec func() error) error {
		fmt.Println("- 1")
		err := exec()
		fmt.Println("- 1")
		return err
	}))

	interceptors.AddOnMethod(method, streamServerInterceptorFunc(func(exec func() error) error {
		fmt.Println("-- 2")
		err := exec()
		fmt.Println("-- 2")
		return err
	}), streamServerInterceptorFunc(func(exec func() error) error {
		fmt.Println("--- 3")
		err := exec()
		fmt.Println("--- 3")
		return err
	}))

	interceptors.AddWithoutMethods([]string{method}, streamServerInterceptorFunc(func(exec func() error) error {
		fmt.Println("- 1")
		err := exec()
		fmt.Println("- 1")
		return err
	}))

	interceptors.AddWithoutMethods([]string{"test"}, streamServerInterceptorFunc(func(exec func() error) error {
		fmt.Println("---- 4")
		err := exec()
		fmt.Println("---- 4")
		return err
	}))

	interceptors.AddWithoutMethods([]string{"demo"}, streamServerInterceptorFunc(func(exec func() error) error {
		fmt.Println("----- 5")
		err := exec()
		fmt.Println("----- 5")
		return err
	}))

	handle := interceptors.StreamServerInterceptor()
	_ = handle("service1", &serverStream{}, &grpc.StreamServerInfo{FullMethod: method}, func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	})
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

func streamServerInterceptorFunc(fn func(exec func() error) error) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return fn(func() error {
			return handler(srv, ss)
		})
	}
}
