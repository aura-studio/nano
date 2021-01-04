package repl

import (
	"fmt"
	"reflect"

	"github.com/lonng/nano/cluster"
)

var localHandler *cluster.LocalHandler

func initLocalHandler() {
	localHandler = cluster.NewHandler()
	for _, c := range opt.Components.List() {
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
	return reflect.New(handler.Type.Elem()).Interface(), nil
}
