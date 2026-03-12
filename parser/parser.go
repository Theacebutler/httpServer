
// request-line   = method SP request-target SP HTTP-version
	n := bytes.Index(rl, SP)
	if n == -1 {
		r.State = ParserError
		return nil, fmt.Errorf("Request line parser Error: No empty space found in request line")
	}
	// DEBUG: read up to n, where method_idx is the idx of the first SP
	method := rl[:n]
	n += len(SP)
	if bytes.HasPrefix(method, SP) {
		r.State = ParserError
		return nil, fmt.Errorf("Request line parser Error: No empty space allowed before the METHOD")
	}
	method = bytes.ToUpper(method)

	// DEBUG: read from n to the end and find the next idx of SP, target_idx is the idx of the next SP
	target_idx := bytes.Index(rl[n:], SP)
	if target_idx == -1 {
		r.State = ParserError
		return nil, fmt.Errorf("Request line parser Error: No empty space found in request line")
	}
	target := rl[n : target_idx+n]
	if len(target) < 1 {
		r.State = ParserError
		return nil, fmt.Errorf("Request line parser Error: No empty space allowed before the request-target")
	}
	after_target_idx := target_idx + n + len(SP)

	version_idx := bytes.Index(rl[after_target_idx:], RN)
	if version_idx == -1 {
		r.State = ParserError
		return nil, fmt.Errorf("Request line parser Error: Invelid version")
	}
	version, err := r.RequestLine.ParseVersion(rl[after_target_idx : after_target_idx+version_idx])
	if err != nil {
		r.State = ParserError
		return nil, err
	}
	return &RequestLine{
		Method:  method,
		Target:  target,
		Version: version,
	}, nil

func (l *RequestLine) ParseVersion(v []byte) ([]byte, error) {
	http, v, ok := bytes.Cut(v, []byte("/"))
	if !ok || http == nil || v == nil {
		return nil, fmt.Errorf("Cant parse HTTP-version")
	}

	for _, x := range v {
		if x == '.' && v[len(v)-1] != '.' {
			continue
		}
		if x > '9' || x < '0' {
			return nil, fmt.Errorf("Cant parse HTTP-version number")
		}
	}

	return fmt.Appendf(nil, "%s/%d", http, v), nil
}
