package backend

import (
	"errors"
)

type HttpCases[K, V JsonPayload] struct {
	RequestsToSend    []K
	ExpectedResponses []V
	HttpMethod        string
	UrlTarget         string
	ExpectedHttpCode  int
}

type ServerMuxHttpCases[K, V JsonPayload] struct {
	RequestsToSend    []K
	ExpectedResponses []V
	HttpMethod        string
	UrlTemplate       string
	UrlTarget         string
	ExpectedHttpCode  int
}

type ByteCase struct {
	ToOutput []byte
	Expected []byte
}

func ConvertToByteCases[K, V JsonPayload](reqs []K, resps []V) (result []ByteCase, err error) {
	if len(reqs) != len(resps) {
		err = errors.New("reqs и resps должны быть одной длины")
		return
	}
	for ind := 0; ind < len(reqs); ind++ {
		var (
			reqBuf  []byte
			respBuf []byte
		)
		reqBuf, err = reqs[ind].Marshal()
		if err != nil {
			return
		}
		respBuf, err = resps[ind].Marshal()
		if err != nil {
			return
		}
		result = append(result, ByteCase{ToOutput: reqBuf, Expected: respBuf})
	}
	return
}
