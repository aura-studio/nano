package codec

import (
	"reflect"
	"testing"

	. "github.com/lonng/nano/packet"
)

func TestPack(t *testing.T) {
	data := []byte("hello world")
	p1 := &Packet{Data: data, Length: len(data)}
	pp1, err := Encode(data)
	if err != nil {
		t.Error(err.Error())
	}

	d1 := NewDecoder()
	packets, err := d1.Decode(pp1)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(packets) < 1 {
		t.Fatal("packets should not empty")
	}
	if !reflect.DeepEqual(p1, packets[0]) {
		t.Fatalf("expect: %v, got: %v", p1, packets[0])
	}

	p2 := &Packet{Data: data, Length: len(data)}
	pp2, err := Encode(data)
	if err != nil {
		t.Error(err.Error())
	}

	d2 := NewDecoder()
	upp2, err := d2.Decode(pp2)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(upp2) < 1 {
		t.Fatal("packets should not empty")
	}
	if !reflect.DeepEqual(p2, upp2[0]) {
		t.Fatalf("expect: %v, got: %v", p2, upp2[0])
	}

	_ = &Packet{Data: data, Length: len(data)}
	if _, err := Encode(data); err != nil {
		t.Error("cannot be err")
	}

	_ = &Packet{Data: data, Length: len(data)}
	if _, err = Encode(data); err != nil {
		t.Error("cannot be err")
	}

	p5 := &Packet{Data: data, Length: len(data)}
	pp5, err := Encode(data)
	if err != nil {
		t.Fatal(err.Error())
	}
	d3 := NewDecoder()
	upp5, err := d3.Decode(append(pp5, []byte{0x00, 0x00, 0x00, 0x00}...))
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(upp5) < 1 {
		t.Fatal("packets should not empty")
	}

	if !reflect.DeepEqual(p5, upp5[0]) {
		t.Fatalf("expect: %v, got: %v", p2, upp5[0])
	}
}

func BenchmarkDecoder_Decode(b *testing.B) {
	data := []byte("hello world")
	pp1, err := Encode(data)
	if err != nil {
		b.Error(err.Error())
	}

	b.ReportAllocs()
	d1 := NewDecoder()
	for i := 0; i < b.N; i++ {
		packets, err := d1.Decode(pp1)
		if err != nil {
			b.Fatal(err)
		}
		if len(packets) != 1 {
			b.Fatal("decode error")
		}
	}
}
