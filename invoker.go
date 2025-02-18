package revel

import (
	"io"
	"reflect"

	"github.com/golang/net/websocket"
)

var (
	controllerType    = reflect.TypeOf(Controller{})
	controllerPtrType = reflect.TypeOf(&Controller{})
	websocketType     = reflect.TypeOf((*websocket.Conn)(nil))
)

func ActionInvoker(c *Controller, _ []Filter) {
	// Instantiate the method.
	methodValue := reflect.ValueOf(c.AppController).MethodByName(c.MethodType.Name)

	// Collect the values for the method's arguments.
	var methodArgs []reflect.Value
	for _, arg := range c.MethodType.Args {
		// If they accept a websocket connection, treat that arg specially.
		var boundArg reflect.Value
		if arg.Type == websocketType {
			boundArg = reflect.ValueOf(c.Request.Websocket)
		} else {
			boundArg = Bind(c.Params, arg.Name, arg.Type)
			// #756 - If the argument is a closer, defer a Close call,
			// so we don't risk on leaks.
			if closer, ok := boundArg.Interface().(io.Closer); ok {
				defer closer.Close()
			}
		}
		methodArgs = append(methodArgs, boundArg)
	}

	var resultValue reflect.Value
	if methodValue.Type().IsVariadic() {
		resultValue = methodValue.CallSlice(methodArgs)[0]
	} else {
		resultValue = methodValue.Call(methodArgs)[0]
	}
	if resultValue.Kind() == reflect.Interface && !resultValue.IsNil() {
		c.Result = resultValue.Interface().(Result)
	}
}
