package message

import (
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	dict := map[string]uint16{
		"test.test.test":  100,
		"test.test.test1": 101,
		"test.test.test2": 102,
		"test.test.test3": 103,
	}
	routes, codes := ParseDictionary(dict)
	m1 := &Message{
		Type:       Request,
		ID:         100,
		Route:      "test.test.test",
		Data:       []byte(`hello world`),
		compressed: true,
	}
	em1, err := Encode(m1, routes)
	if err != nil {
		t.Error(err.Error())
	}
	dm1, _, err := Decode(em1, codes)
	if err != nil {
		t.Error(err.Error())
	}

	t.Log("encode:", m1)
	t.Log("decode:", dm1)

	if !reflect.DeepEqual(m1, dm1) {
		t.Error("not equal")
	}

	m2 := &Message{
		Type:  Request,
		ID:    100,
		Route: "test.test.test4",
		Data:  []byte(`hello world`),
	}
	em2, err := Encode(m2, routes)
	if err != nil {
		t.Error(err.Error())
	}
	dm2, _, err := Decode(em2, codes)
	if err != nil {
		t.Error(err.Error())
	}

	t.Log("encode:", m2)
	t.Log("decode:", dm2)

	if !reflect.DeepEqual(m2, dm2) {
		t.Error("not equal")
	}

	m3 := &Message{
		Type: Response,
		ID:   100,
		Data: []byte(`hello world`),
	}
	em3, err := Encode(m3, routes)
	if err != nil {
		t.Error(err.Error())
	}
	dm3, _, err := Decode(em3, codes)
	if err != nil {
		t.Error(err.Error())
	}

	t.Log("encode:", m3)
	t.Log("decode:", dm3)

	if !reflect.DeepEqual(m3, dm3) {
		t.Error("not equal")
	}

	m4 := &Message{
		Type: Response,
		ID:   100,
		Data: []byte(`hello world`),
	}
	em4, err := Encode(m4, routes)
	if err != nil {
		t.Error(err.Error())
	}
	dm4, _, err := Decode(em4, codes)
	if err != nil {
		t.Error(err.Error())
	}

	t.Log("encode:", m4)
	t.Log("decode:", dm4)

	if !reflect.DeepEqual(m4, dm4) {
		t.Error("not equal")
	}

	m5 := &Message{
		Type:       Notify,
		Route:      "test.test.test",
		Data:       []byte(`hello world`),
		compressed: true,
	}
	em5, err := Encode(m5, routes)
	if err != nil {
		t.Error(err.Error())
	}
	dm5, _, err := Decode(em5, codes)
	if err != nil {
		t.Error(err.Error())
	}

	t.Log("encode:", m5)
	t.Log("decode:", dm5)

	if !reflect.DeepEqual(m5, dm5) {
		t.Error("not equal")
	}

	m6 := &Message{
		Type:  Notify,
		Route: "test.test.test20",
		Data:  []byte(`hello world`),
	}
	em6, err := Encode(m6, routes)
	if err != nil {
		t.Error(err.Error())
	}
	dm6, _, err := Decode(em6, codes)
	if err != nil {
		t.Error(err.Error())
	}

	t.Log("encode:", m6)
	t.Log("decode:", dm6)

	if !reflect.DeepEqual(m6, dm6) {
		t.Error("not equal")
	}

	m7 := &Message{
		Type:  Push,
		Route: "test.test.test9",
		Data:  []byte(`hello world`),
	}
	em7, err := Encode(m7, routes)
	if err != nil {
		t.Error(err.Error())
	}
	dm7, _, err := Decode(em7, codes)
	if err != nil {
		t.Error(err.Error())
	}

	t.Log("encode:", m7)
	t.Log("decode:", dm7)

	if !reflect.DeepEqual(m7, dm7) {
		t.Error("not equal")
	}

	m8 := &Message{
		Type:       Push,
		Route:      "test.test.test3",
		Data:       []byte(`hello world`),
		compressed: true,
	}
	em8, err := Encode(m8, routes)
	if err != nil {
		t.Error(err.Error())
	}
	dm8, _, err := Decode(em8, codes)
	if err != nil {
		t.Error(err.Error())
	}

	t.Log("encode:", m8)
	t.Log("decode:", dm8)

	if !reflect.DeepEqual(m8, dm8) {
		t.Error("not equal")
	}
}
