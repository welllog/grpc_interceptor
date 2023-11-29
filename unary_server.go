package grpc_interceptor

import (
	"context"

	"google.golang.org/grpc"
)

type unaryServerInterceptorGroup struct {
	ics    []grpc.UnaryServerInterceptor
	domain domain
}

type UnaryServerInterceptors struct {
	s []unaryServerInterceptorGroup
}

func (usi *UnaryServerInterceptors) Add(interceptors ...grpc.UnaryServerInterceptor) *UnaryServerInterceptors {
	usi.s = append(usi.s, unaryServerInterceptorGroup{
		ics:    interceptors,
		domain: newDomain(),
	})
	return usi
}

func (usi *UnaryServerInterceptors) AddWithoutMethods(methods []string, interceptors ...grpc.UnaryServerInterceptor) *UnaryServerInterceptors {
	usi.s = append(usi.s, unaryServerInterceptorGroup{
		ics:    interceptors,
		domain: newBlackDomain(methods),
	})
	return usi
}

func (usi *UnaryServerInterceptors) AddOnMethods(methods []string, interceptors ...grpc.UnaryServerInterceptor) *UnaryServerInterceptors {
	if len(methods) > 0 {
		usi.s = append(usi.s, unaryServerInterceptorGroup{
			ics:    interceptors,
			domain: newWhiteDomain(methods),
		})
	}
	return usi
}

func (usi *UnaryServerInterceptors) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		if len(usi.s) == 0 {
			return handler(ctx, req)
		}

		var cursor handleCursor
		var chainHandler grpc.UnaryHandler

		chainHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
			for cursor.segment < len(usi.s) {
				group := usi.s[cursor.segment]
				if group.domain.isOnMethod(info.FullMethod) && cursor.offset < len(group.ics) {
					ic := group.ics[cursor.offset]
					cursor.offset++
					return ic(ctx, req, info, chainHandler)
				}

				cursor.offset = 0
				cursor.segment++
			}

			return handler(ctx, req)
		}

		return chainHandler(ctx, req)
	}
}
