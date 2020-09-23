package grpc_interceptor

import (
    "context"
    "google.golang.org/grpc"
    "strconv"
)

type UnaryServerHandler func(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error)

type unaryServerHandlerGroup struct {
    name string
    handlers []UnaryServerHandler
    skip map[string]struct{}
}

type UnaryServerInterceptors struct {
    global []*unaryServerHandlerGroup
    special  map[string][]UnaryServerHandler
    count int
}

func (interceptors *UnaryServerInterceptors) AddGlobalHandlerGroup(groupName string, handlers []UnaryServerHandler,
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
    interceptors.global = append(interceptors.global, &unaryServerHandlerGroup{
        name: groupName,
        handlers: handlers,
        skip: skipm,
    })
    interceptors.count++
}

func (interceptors *UnaryServerInterceptors) AddGlobalHandler(handler UnaryServerHandler, skip ...string) {
    skipm := make(map[string]struct{}, len(skip))
    for _, s := range skip {
        skipm[s] = struct{}{}
    }
    interceptors.global = append(interceptors.global, &unaryServerHandlerGroup{
        name: "group@" + strconv.Itoa(interceptors.count),
        handlers: []UnaryServerHandler{handler},
        skip: skipm,
    })
    interceptors.count++
}

func (interceptors *UnaryServerInterceptors) AddSpecialHandler(srvMethod string, handlers ...UnaryServerHandler) {
    if interceptors.special == nil {
        interceptors.special = make(map[string][]UnaryServerHandler)
    }
    if _, ok := interceptors.special[srvMethod]; !ok {
        interceptors.special[srvMethod] = handlers
        return
    }
    interceptors.special[srvMethod] = append(interceptors.special[srvMethod], handlers...)
}

func (interceptors *UnaryServerInterceptors) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler)(interface{}, error) {
        
        globalCount := len(interceptors.global)
        specialCount := len(interceptors.special[info.FullMethod])
        
        if globalCount + specialCount == 0 {
            return handler(ctx, req)
        }
        
        var (
            groupI, handlerI int
            chainHandler grpc.UnaryHandler
        )
        chainHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
            if groupI < globalCount {
                for {
                    group := interceptors.global[groupI]
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
            
            if handlerI < specialCount {
                special := interceptors.special[info.FullMethod]
                index := handlerI
                handlerI++
                return special[index](ctx, req, info, chainHandler)
            }
            
            return handler(ctx, req)
        }
        
        return chainHandler(ctx, req)
    }
}



