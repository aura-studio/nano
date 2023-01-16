package repl

import (
	"fmt"
	"reflect"

	"github.com/aura-studio/nano/cluster"
)

var localHandler *cluster.LocalHandler

func initLocalHandler() {
	localHandler = cluster.NewHandler()
	for _, c := range options.Components.List() {
		if err := localHandler.Register(c.Comp, c.Opts); err != nil {
			panic(err)
		}
	}
}

// routeMessage creates msg interface from LocalHandler
func routeMessage(route string) (interface{}, error) {
	handler, err := localHandler.RouteHandler(route)
	if err != nil {
		return nil, fmt.Errorf("unexpected route:%s, can not find it's route handler", route)
	}
	if handler.Type == reflect.TypeOf([]byte{}) {
		return reflect.New(handler.Type).Elem().Interface(), nil
	}
	return reflect.New(handler.Type.Elem()).Interface(), nil
}
