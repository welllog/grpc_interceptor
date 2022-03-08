package grpc_interceptor

import (
	"google.golang.org/grpc"
)

type streamServerInterceptorGroup struct {
	handlers []grpc.StreamServerInterceptor
	skip     map[string]struct{}
}

type StreamServerInterceptors struct {
	global  []*streamServerInterceptorGroup
	ggCount int
	part    map[string][]grpc.StreamServerInterceptor
}

func (ssi *StreamServerInterceptors) UseGlobal(interceptors []grpc.StreamServerInterceptor, skipMethods ...string) {
	skip := make(map[string]struct{}, len(skipMethods))
	for _, method := range skipMethods {
		skip[method] = struct{}{}
	}

	ssi.global = append(ssi.global, &streamServerInterceptorGroup{
		handlers: interceptors,
		skip:     skip,
	})
	ssi.ggCount++
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

		mCount := len(ssi.part[info.FullMethod])

		if ssi.ggCount+mCount == 0 {
			return handler(srv, ss)
		}

		curI := handlerCurI{mCount: mCount}
		var chainHandler grpc.StreamHandler
		chainHandler = func(srv interface{}, ss grpc.ServerStream) error {
			if curI.groupI < ssi.ggCount {
				for {
					group := ssi.global[curI.groupI]
					if _, ok := group.skip[info.FullMethod]; !ok {
						index := curI.handlerI
						curI.handlerI++
						if index < len(group.handlers) {
							return group.handlers[index](srv, ss, info, chainHandler)
						}
						curI.handlerI = 0
					}
					curI.groupI++
					if curI.groupI >= ssi.ggCount {
						break
					}
				}
			}

			if curI.handlerI < curI.mCount {
				special := ssi.part[info.FullMethod]
				index := curI.handlerI
				curI.handlerI++
				return special[index](srv, ss, info, chainHandler)
			}

			return handler(srv, ss)
		}

		return chainHandler(srv, ss)
	}
}
