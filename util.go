package grpc_interceptor

type handleCursor struct {
	segment int
	offset  int
	ids     []int
}

type domain struct {
	without map[string]struct{}
	method  string
	typ     int
}

func (d domain) isOnMethod(method string) bool {
	switch d.typ {
	case 0:
		return true
	case 1:
		return d.method == method
	case 2:
		_, ok := d.without[method]
		return !ok
	default:
		return true
	}
}

func newBlackDomain(black []string) domain {
	if len(black) > 0 {
		without := make(map[string]struct{}, len(black))
		for _, method := range black {
			without[method] = struct{}{}
		}
		return domain{
			without: without,
			typ:     2,
		}
	}

	return domain{}
}

func newDomain() domain {
	return domain{}
}

func newSpecificDomain(method string) domain {
	return domain{
		method: method,
		typ:    1,
	}
}
