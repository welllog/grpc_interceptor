package grpc_interceptor

import (
    "google.golang.org/grpc"
    "strconv"
)

type StreamServerHandler func(
    srv interface{},
    ss grpc.ServerStream,
    info *grpc.StreamServerInfo,
    handler grpc.StreamHandler,
) error

type streamServerHandlerGroup struct {
    name string
    handlers []StreamServerHandler
    skip map[string]struct{}
}

type StreamServerInterceptors struct {
    global []*streamServerHandlerGroup
    special  map[string][]StreamServerHandler
    count int
}

func (interceptors *StreamServerInterceptors) AddGlobalHandlerGroup(groupName string, handlers []StreamServerHandler,
    skip ...string) {
    
    skipm := make(map[string]struct{}, len(skip))
    for _, s := range skip {
        skipm[s] = struct{}{}
    }
    for i := range interceptors.global {
        if interceptors.global[i].name == groupName {
            interceptors.global[i].handlers = handlers
            interceptors.global[i].skip = skipm
            return
        }
    }
    interceptors.global = append(interceptors.global, &streamServerHandlerGroup{
        name: groupName,
        handlers: handlers,
        skip: skipm,
    })
    interceptors.count++
}

func (interceptors *StreamServerInterceptors) AddGlobalHandler(handler StreamServerHandler, skip ...string) {
    skipm := make(map[string]struct{}, len(skip))
    for _, s := range skip {
        skipm[s] = struct{}{}
    }
    interceptors.global = append(interceptors.global, &streamServerHandlerGroup{
        name: "group@" + strconv.Itoa(interceptors.count),
        handlers: []StreamServerHandler{handler},
        skip: skipm,
    })
    interceptors.count++
}

func (interceptors *StreamServerInterceptors) AddSpecialHandler(srvMethod string, handlers ...StreamServerHandler) {
    if interceptors.special == nil {
        interceptors.special = make(map[string][]StreamServerHandler)
    }
    if _, ok := interceptors.special[srvMethod]; !ok {
        interceptors.special[srvMethod] = handlers
        return
    }
    interceptors.special[srvMethod] = append(interceptors.special[srvMethod], handlers...)
}

func (interceptors *StreamServerInterceptors) StreamServerInterceptor() grpc.StreamServerInterceptor {
    return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
        handler grpc.StreamHandler) error {
        
        globalCount := len(interceptors.global)
        specialCount := len(interceptors.special[info.FullMethod])
        
        if globalCount + specialCount == 0 {
            return handler(srv, ss)
        }
        
        var (
            groupI, handlerI int
            chainHandler grpc.StreamHandler
        )
        chainHandler = func(srv interface{}, ss grpc.ServerStream) error {
            if groupI < globalCount {
                for {
                    group := interceptors.global[groupI]
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
            
            if handlerI < specialCount {
                special := interceptors.special[info.FullMethod]
                index := handlerI
                handlerI++
                return special[index](srv, ss, info, chainHandler)
            }
            
            return handler(srv, ss)
        }
        
        return chainHandler(srv, ss)
    }
}
