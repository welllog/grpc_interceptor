package grpc_interceptor

import (
	"google.golang.org/grpc"
)

type streamServerInterceptorGroup struct {
	ics    []grpc.StreamServerInterceptor
	domain domain
}

type StreamServerInterceptors struct {
	s []streamServerInterceptorGroup
}

func (ssi *StreamServerInterceptors) Add(interceptors ...grpc.StreamServerInterceptor) *StreamServerInterceptors {
	ssi.s = append(ssi.s, streamServerInterceptorGroup{
		ics:    interceptors,
		domain: newDomain(),
	})
	return ssi
}

func (ssi *StreamServerInterceptors) AddWithoutMethods(methods []string, interceptors ...grpc.StreamServerInterceptor) *StreamServerInterceptors {
	ssi.s = append(ssi.s, streamServerInterceptorGroup{
		ics:    interceptors,
		domain: newBlackDomain(methods),
	})
	return ssi
}

func (ssi *StreamServerInterceptors) AddOnMethods(methods []string, interceptors ...grpc.StreamServerInterceptor) *StreamServerInterceptors {
	if len(methods) > 0 {
		ssi.s = append(ssi.s, streamServerInterceptorGroup{
			ics:    interceptors,
			domain: newWhiteDomain(methods),
		})
	}
	return ssi
}

func (ssi *StreamServerInterceptors) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
		handler grpc.StreamHandler) error {

		if len(ssi.s) == 0 {
			return handler(srv, ss)
		}

		var cursor handleCursor
		var chainHandler grpc.StreamHandler

		chainHandler = func(srv interface{}, ss grpc.ServerStream) error {
			for cursor.segment < len(ssi.s) {
				group := ssi.s[cursor.segment]
				if group.domain.isOnMethod(info.FullMethod) && cursor.offset < len(group.ics) {
					ic := group.ics[cursor.offset]
					cursor.offset++
					return ic(srv, ss, info, chainHandler)
				}

				cursor.offset = 0
				cursor.segment++
			}

			return handler(srv, ss)
		}

		return chainHandler(srv, ss)
	}
}
