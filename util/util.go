package util

import (
	"io"
	"log"
)

func HandleClose(closer io.Closer) {
	if closer != nil {
		err := closer.Close()
		if err != nil {
			log.Panicln("error on close: ", err)
		}
	}
}
