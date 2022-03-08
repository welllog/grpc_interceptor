package grpc_interceptor

import (
	"context"

	"google.golang.org/grpc"
)

type handlerCurI struct {
	groupI   int
	handlerI int
	mCount   int
}

type unaryServerInterceptorGroup struct {
	handlers []grpc.UnaryServerInterceptor
	skip     map[string]struct{}
}

type UnaryServerInterceptors struct {
	global  []*unaryServerInterceptorGroup
	ggCount int
	part    map[string][]grpc.UnaryServerInterceptor
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
	usi.ggCount++
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
		handler grpc.UnaryHandler) (interface{}, error) {

		mCount := len(usi.part[info.FullMethod])

		if usi.ggCount+mCount == 0 {
			return handler(ctx, req)
		}

		curI := handlerCurI{mCount: mCount}
		var chainHandler grpc.UnaryHandler

		chainHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
			if curI.groupI < usi.ggCount {
				for {
					group := usi.global[curI.groupI]
					if _, ok := group.skip[info.FullMethod]; !ok {
						index := curI.handlerI
						curI.handlerI++
						if index < len(group.handlers) {
							return group.handlers[index](ctx, req, info, chainHandler)
						}
						curI.handlerI = 0
					}
					curI.groupI++
					if curI.groupI >= usi.ggCount {
						break
					}
				}
			}

			if curI.handlerI < curI.mCount {
				special := usi.part[info.FullMethod]
				index := curI.handlerI
				curI.handlerI++
				return special[index](ctx, req, info, chainHandler)
			}

			return handler(ctx, req)
		}

		return chainHandler(ctx, req)
	}
}
