package grpc_interceptor

import (
    "google.golang.org/grpc"
)

type streamServerInterceptorGroup struct {
    handlers []grpc.StreamServerInterceptor
    skip map[string]struct{}
}

type StreamServerInterceptors struct {
    global []*streamServerInterceptorGroup
    part  map[string][]grpc.StreamServerInterceptor
}

func (ssi *StreamServerInterceptors) UseGlobal(interceptors []grpc.StreamServerInterceptor, skipMethods ...string) {
    skip := make(map[string]struct{}, len(skipMethods))
    for _, method := range skipMethods {
        skip[method] = struct{}{}
    }
    
    ssi.global = append(ssi.global, &streamServerInterceptorGroup{
        handlers: interceptors,
        skip: skip,
    })
}

func (ssi *StreamServerInterceptors) UseMethod(method string, interceptors ...grpc.StreamServerInterceptor) {
    if ssi.part == nil {
        ssi.part = make(map[string][]grpc.StreamServerInterceptor)
        ssi.part[method] = interceptors
        return
    }
    
    if _, ok := ssi.part[method]; !ok {
        ssi.part[method] = interceptors
        return
    }
    
    ssi.part[method] = append(ssi.part[method], interceptors...)
}

func (ssi *StreamServerInterceptors) StreamServerInterceptor() grpc.StreamServerInterceptor {
    return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
        handler grpc.StreamHandler) error {
        
        globalCount := len(ssi.global)
        methodCount := len(ssi.part[info.FullMethod])
        
        if globalCount + methodCount == 0 {
            return handler(srv, ss)
        }
        
        var (
            groupI, handlerI int
            chainHandler grpc.StreamHandler
        )
        chainHandler = func(srv interface{}, ss grpc.ServerStream) error {
            if groupI < globalCount {
                for {
                    group := ssi.global[groupI]
                    if _, ok := group.skip[info.FullMethod]; !ok {
                        index := handlerI
                        handlerI++
                        if index < len(group.handlers) {
                            return group.handlers[index](srv, ss, info, chainHandler)
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
                special := ssi.part[info.FullMethod]
                index := handlerI
                handlerI++
                return special[index](srv, ss, info, chainHandler)
            }
            
            return handler(srv, ss)
        }
        
        return chainHandler(srv, ss)
    }
}
