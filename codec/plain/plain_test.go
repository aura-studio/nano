package plain_test

import (
	"reflect"
	"testing"

	"github.com/aura-studio/nano/codec/plain"
	"github.com/aura-studio/nano/packet"
)

func TestPack(t *testing.T) {
	var e = plain.NewEncoder()
	data := []byte("hello world")
	p1 := &packet.Packet{Data: data, Length: len(data)}
	pp1, err := e.Encode(p1)
	if err != nil {
		t.Error(err.Error())
	}

	d1 := plain.NewDecoder()
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

	p2 := &packet.Packet{Data: data, Length: len(data)}
	pp2, err := e.Encode(p2)
	if err != nil {
		t.Error(err.Error())
	}

	d2 := plain.NewDecoder()
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

	p3 := &packet.Packet{Length: len(data), Data: data}
	pp3, err := e.Encode(p3)
	if err != nil {
		t.Fatal(err.Error())
	}
	d3 := plain.NewDecoder()
	upp3, err := d3.Decode(append(pp3, []byte{0x00, 0x00, 0x00, 0x00}...))
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(upp3) < 1 {
		t.Fatal("packets should not empty")
	}

	if !reflect.DeepEqual(p3, upp3[0]) {
		t.Fatalf("expect: %v, got: %v", p2, upp3[0])
	}
}

func BenchmarkDecoder_Decode(b *testing.B) {
	var e = plain.NewEncoder()
	data := []byte("hello world")
	pp1, err := e.Encode(&packet.Packet{Length: len(data), Data: data})
	if err != nil {
		b.Error(err.Error())
	}

	b.ReportAllocs()
	d1 := plain.NewDecoder()
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
