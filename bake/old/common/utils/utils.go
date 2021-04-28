package utils

import "fmt"

func SafeSend(msgPush chan []byte, mm []byte) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("safeSend; recovered from ", r)
		}
	}()
	msgPush <- mm
}

func SafeClose(ch chan []byte) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("safeClose; recovered from ", r)
		}
	}()
	close(ch)
}