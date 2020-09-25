package grpc_interceptor

import (
    "context"
    "google.golang.org/grpc"
)

type unaryServerInterceptorGroup struct {
    handlers []grpc.UnaryServerInterceptor
    skip map[string]struct{}
}

type UnaryServerInterceptors struct {
    global []*unaryServerInterceptorGroup
    part  map[string][]grpc.UnaryServerInterceptor
}

func (usi *UnaryServerInterceptors) UseGlobal(interceptors []grpc.UnaryServerInterceptor, skipMethods ...string) {
    skip := make(map[string]struct{}, len(skipMethods))
    for _, method := range skipMethods {
        skip[method] = struct{}{}
    }
    
    usi.global = append(usi.global, &unaryServerInterceptorGroup{
        handlers: interceptors,
        skip:     skip,
    })
}

func (usi *UnaryServerInterceptors) UseMethod(method string, interceptors ...grpc.UnaryServerInterceptor) {
    if usi.part == nil {
        usi.part = make(map[string][]grpc.UnaryServerInterceptor)
        usi.part[method] = interceptors
        return
    }
    
    if _, ok := usi.part[method]; !ok {
        usi.part[method] = interceptors
        return
    }
    
    usi.part[method] = append(usi.part[method], interceptors...)
}

func (usi *UnaryServerInterceptors) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler)(interface{}, error) {
        
        globalCount := len(usi.global)
        methodCount := len(usi.part[info.FullMethod])
    
        if globalCount + methodCount == 0 {
            return handler(ctx, req)
        }
        
        var (
            groupI, handlerI int
            chainHandler grpc.UnaryHandler
        )
        chainHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
            if groupI < globalCount {
                for {
                    group := usi.global[groupI]
                    if _, ok := group.skip[info.FullMethod]; !ok {
                        index := handlerI
                        handlerI++
                        if index < len(group.handlers) {
                            return group.handlers[index](ctx, req, info, chainHandler)
                        }
                        handlerI = 0
                    }
                    groupI++
                    if groupI >= globalCount {
                        break
                    }
                }
            }
            
            if handlerI < methodCount {
                special := usi.part[info.FullMethod]
                index := handlerI
                handlerI++
                return special[index](ctx, req, info, chainHandler)
            }
            
            return handler(ctx, req)
        }
        
        return chainHandler(ctx, req)
    }
}



