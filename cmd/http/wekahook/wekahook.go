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

func Maze(obj interface{}) {
	typeObj := reflect.TypeOf(obj)
	if typeObj.Kind() == reflect.Ptr {
		typeObj = typeObj.Elem()
	}
	fmt.Printf("Maze? %T kind of %v\n", obj , typeObj.Kind() )
	if typeObj.Kind() == reflect.Struct {
		for i := 0; i < typeObj.NumField(); i++ {
			fmt.Printf("Maze? \t %d - %s: %v\n", i, typeObj.Field(i).Name, typeObj.Field(i).Type)
		}
	}
}


func (w WResponseWriter) WriteHeader(statusCode int) {
	pw := w.ResponseWriter

	httpCom := reflect.ValueOf(pw).Elem().Field(0) // We know ResponseWriter is http.response ( private structure)



	netComPrivate := httpCom.Elem().Field(2) // the field num 2 is a http com object
	PointerNetCom := netComPrivate.UnsafeAddr()

	tmpObj := reflect.NewAt(netComPrivate.Type(), unsafe.Pointer(PointerNetCom)).Elem().Interface()
	Maze(tmpObj)
	netCom := tmpObj.(net.Conn)

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
