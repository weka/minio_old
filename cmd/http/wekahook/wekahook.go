package wekahook

import (
	"fmt"
	"net"
	"net/http"
	"reflect"
	"time"
	"unsafe"
)

/*
WResponseWriter is a wrapper of http.ResponseWriter
*/
type WResponseWriter struct {
	http.ResponseWriter
	UpgradeMode *bool
}

func (w WResponseWriter) Flush() {
	if flush, ok := w.ResponseWriter.(http.Flusher); ok {
		flush.Flush()
	}
}

func (w WResponseWriter) WriteHeader(statusCode int) {
	pw := w.ResponseWriter
	httpCom := reflect.ValueOf(pw).Elem().Field(0) // We know ResponseWriter is http.response ( private structure)
	// we took the field 0 which is a netcom object
	netComPrivate := httpCom.Elem().Field(2) // the field num 2 is a http com object
	PointerNetCom := netComPrivate.UnsafeAddr()

	netCom := reflect.NewAt(netComPrivate.Type(), unsafe.Pointer(PointerNetCom)).Elem().Interface().(net.Conn)

	netCom.Write([]byte("HTTP/1.1 "))
	fmt.Printf("%T -> %v ->  %v -> %T\n", pw, httpCom.Type(), netComPrivate, netCom)

	if false {
		w.ResponseWriter.WriteHeader(http.StatusGatewayTimeout)
	} else {
		for *w.UpgradeMode {
			netCom.Write([]byte(" "))
			time.Sleep(1 * time.Second)
		}
	}

	netCom.Write([]byte(fmt.Sprintf("%d %s\nProtocol: ", statusCode, http.StatusText(statusCode))))
	w.ResponseWriter.WriteHeader(statusCode)
}
