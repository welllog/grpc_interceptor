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

func (usi *UnaryServerInterceptors) AddOnMethod(method string, interceptors ...grpc.UnaryServerInterceptor) *UnaryServerInterceptors {
	usi.s = append(usi.s, unaryServerInterceptorGroup{
		ics:    interceptors,
		domain: newSpecificDomain(method),
	})
	return usi
}

func (usi *UnaryServerInterceptors) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		var cursor handleCursor
		for i := range usi.s {
			if usi.s[i].domain.isOnMethod(info.FullMethod) {
				cursor.ids = append(cursor.ids, i)
			}
		}

		if len(cursor.ids) == 0 {
			return handler(ctx, req)
		}

		var chainHandler grpc.UnaryHandler

		chainHandler = func(ctx context.Context, req interface{}) (interface{}, error) {

			for cursor.segment < len(cursor.ids) {
				group := usi.s[cursor.ids[cursor.segment]]
				if cursor.offset < len(group.ics) {
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
