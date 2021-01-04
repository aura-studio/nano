package coroutine

import (
	"testing"

	"strings"
)

func TestCreate(t *testing.T) {
	exit := make(chan int)
	co := Create(func(co ID, args ...interface{}) error {
		strs := interfaceSliceToStringSlice(args)
		out := strings.Join(strs, " ") // coroutine resume 1
		if out != "ID resume 1" {
			t.Error("ID flow error, should be ID resume 1")
		}
		t.Log(out)

		inData := Yield(co, "ID", "yield", "2")
		strs = interfaceSliceToStringSlice(inData)
		out = strings.Join(strs, " ") // coroutine resume 3
		if out != "ID resume 3" {
			t.Error("ID flow error, should be ID resume 3")
		}
		t.Log(out)

		_ = Yield(co, "ID", "yield", "4")
		return nil
	})

	_, ok := TryResume(co, "ID", "resume", "0")
	if !ok {
		t.Log("Try resume test result Correct")
	}

	_, ok = Resume(co, "ID", "resume", "1")
	if !ok {
		t.Log("Dead ID")
	}

	outData, ok := Resume(co, "ID", "resume", "3")
	if !ok {
		t.Log("Dead ID")
	}
	strs := interfaceSliceToStringSlice(outData)
	out := strings.Join(strs, " ") // coroutine yield 2
	if out != "ID yield 2" {
		t.Error("ID flow error, should be ID yield 2")
	}
	t.Log(out)

	AsyncResume(co, func(outData ...interface{}) {
		strs = interfaceSliceToStringSlice(outData)
		out := strings.Join(strs, " ") // coroutine yield 4
		if out != "ID yield 4" {
			t.Error("ID flow error, should be ID yield 4")
		}
		t.Log(out)

		exit <- 1
	}, "ID", "resume", "3")

	<-exit
}

func TestStart(t *testing.T) {
	co := Wrap(
		func(co ID, args ...interface{}) error {
			strs := interfaceSliceToStringSlice(args)
			out := strings.Join(strs, " ") // ID call 1
			if out != "ID call 1" {
				t.Error("ID flow error, should be ID call 1")
			}
			t.Log(out)

			inData := Yield(co, "ID", "yield", "2")
			strs = interfaceSliceToStringSlice(inData)
			out = strings.Join(strs, " ") // ID resume 3
			if out != "ID resume 3" {
				t.Error("ID flow error, should be ID resume 3")
			}
			t.Log(out)

			_ = Yield(co, "ID", "yield", "4")
			return nil
		})

	go Call(co, "ID", "call", "1")

	outData, ok := Resume(co, "ID", "resume", "3")
	if !ok {
		t.Log("Dead ID")
	}
	strs := interfaceSliceToStringSlice(outData)
	out := strings.Join(strs, " ") // ID yield 2
	if out != "ID yield 2" {
		t.Error("ID flow error, should be ID yield 2")
	}
	t.Log(out)

	outData, ok = Resume(co, "ID", "resume", "5")
	if !ok {
		t.Log("Dead ID")
	}
	strs = interfaceSliceToStringSlice(outData)
	out = strings.Join(strs, " ") // ID yield 4
	if out != "ID yield 4" {
		t.Error("ID flow error, should be ID yield 4")
	}
	t.Log(out)
}
