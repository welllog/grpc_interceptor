package grpc_interceptor

import (
    "context"
    "errors"
    "fmt"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    "log"
    "testing"
)

func TestStreamServer(t *testing.T) {
    streamServMd := &StreamServerInterceptors{}
    
    streamServMd.AddSpecialHandler("/xx.TestService/TestStream", func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        fmt.Println("special --- 0")
        if val, ok := srv.(string); ok {
            if val == "service" {
                return errors.New("srv is string")
            }
        }
        err := handler(srv, ss)
        fmt.Println("special --- 0 --- end")
        return err
    })
    
    streamServMd.AddSpecialHandler("/xx.TestService/TestStream", func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        fmt.Println("special --- 1")
        err := handler(srv, ss)
        fmt.Println("special --- 1 --- end")
        return err
    },func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        fmt.Println("special --- 2")
        err := handler(srv, ss)
        fmt.Println("special --- 2 --- end")
        return err
    })
    
    streamServMd.AddGlobalHandlerGroup("test1", []StreamServerHandler{
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
    
    streamServMd.AddGlobalHandlerGroup("test2", []StreamServerHandler{
        func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
            fmt.Println("global --- 3")
            err := handler(srv, ss)
            fmt.Println("global --- 3 --- end")
            return err
        },
    }, "/xx.TestService/TestStream")
    
    streamServMd.AddGlobalHandler(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        fmt.Println("global --- 2")
        err := handler(srv, ss)
        fmt.Println("global --- 2 --- end")
        return err
    })
    
    funcs := streamServMd.StreamServerInterceptor()
    err := funcs("service1", &serverStream{}, &grpc.StreamServerInfo{FullMethod: "/xx.TestService/TestStream"}, func(srv interface{}, stream grpc.ServerStream) error {
        fmt.Println("stream handle ...")
        return nil
    })
    fmt.Println(err)
}

type serverStream struct {}

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

func ServerDemo2(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
handler grpc.StreamHandler) error {
    log.Printf("before handling. Info: %+v", info)
    err := handler(srv, ss)
    log.Printf("after handling. err: %v", err)
    return err
}
