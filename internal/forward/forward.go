package forward

import (
	"bytes"
	"log"

	"github.com/nayukidayo/tjg-go/env"
	"github.com/nayukidayo/tjg-go/fetch"
)

var (
	url    = env.GetStr("FORWARD_URL", "")
	header = map[string]string{"Content-Type": "application/json"}
)

func Send(data []byte) {
	_, err := fetch.Fetch("POST", url, bytes.NewReader(data), header)
	if err != nil {
		log.Println(err)
	}
}
