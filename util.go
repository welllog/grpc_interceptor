package grpc_interceptor

type handleCursor struct {
	segment int
	offset  int
}

type domain struct {
	set map[string]struct{}
	typ int
}

func (d domain) isOnMethod(method string) bool {
	switch d.typ {
	case 0:
		return true
	case 1:
		_, ok := d.set[method]
		return ok
	case 2:
		_, ok := d.set[method]
		return !ok
	default:
		return true
	}
}

func newBlackDomain(black []string) domain {
	if len(black) > 0 {
		set := make(map[string]struct{}, len(black))
		for _, method := range black {
			set[method] = struct{}{}
		}
		return domain{
			set: set,
			typ: 2,
		}
	}

	return domain{}
}

func newWhiteDomain(white []string) domain {
	set := make(map[string]struct{}, len(white))
	for _, method := range white {
		set[method] = struct{}{}
	}
	return domain{
		set: set,
		typ: 1,
	}
}

func newDomain() domain {
	return domain{}
}
