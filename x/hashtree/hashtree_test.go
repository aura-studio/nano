package hashtree

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"testing"

	cli "github.com/lonng/nano/x/hashtree/testdata"
	"github.com/lonng/nano/x/virtualtime"
)

type testA struct {
	Root
	ID Int64
	B  testB
}

type testB struct {
	Time Int64
}

func TestHashtree(t *testing.T) {
	now := virtualtime.Now().Unix()
	a := &testA{}
	hash := map[string]string{
		"ID":     "100",
		"B.Time": strconv.FormatInt(now, 10),
	}
	err := a.Load(a, hash)
	if err != nil {
		t.Error(err)
	}
	if a.ID.Get() != 100 {
		t.Errorf("ID field init failed")
	}
	if a.B.Time.Get() != now {
		t.Errorf("B.Time field init failed")
	}
	a.ID.Set(101)
	if a.ID.Get() != 101 {
		t.Errorf("get not equals set")
	}

	a.B.Time.Set(now + 1)
	hash2, err := a.Dump()
	if err != nil {
		t.Error(err)
	}
	hash3 := map[string]string{
		"ID":     "101",
		"B.Time": strconv.FormatInt(now+1, 10),
	}
	if !reflect.DeepEqual(hash2, hash3) {
		t.Errorf("hash2[%+v] does not equal hash3[%+v]", hash2, hash3)
	}
}

type testC struct {
	Root
	i    Int
	i8   Int8
	i16  Int16
	i32  Int32
	i64  Int64
	u    Uint
	u8   Uint8
	u16  Uint16
	u32  Uint32
	u64  Uint64
	f32  Float32
	f64  Float64
	bi   BigInt
	br   BigRat
	bf   BigFloat
	b    Bool
	s    String
	si   SliceInt
	si8  SliceInt8
	si16 SliceInt16
	si32 SliceInt32
	si64 SliceInt64
	su   SliceUint
	su8  SliceUint8
	su16 SliceUint16
	su32 SliceUint32
	su64 SliceUint64
	st   SliceTime
	t    Time
	j    JSON
	p    Proto
	d    testD
}

type testD struct {
	i    Int
	i8   Int8
	i16  Int16
	i32  Int32
	i64  Int64
	u    Uint
	u8   Uint8
	u16  Uint16
	u32  Uint32
	u64  Uint64
	f32  Float32
	f64  Float64
	bi   BigInt
	br   BigRat
	bf   BigFloat
	b    Bool
	s    String
	si   SliceInt
	si8  SliceInt8
	si16 SliceInt16
	si32 SliceInt32
	si64 SliceInt64
	su   SliceUint
	su8  SliceUint8
	su16 SliceUint16
	su32 SliceUint32
	su64 SliceUint64
	st   SliceTime
	t    Time
	j    JSON
	p    Proto
}

func TestAllTypes(t *testing.T) {
	ts := virtualtime.Now().Unix()

	kickOff := &cli.KickPush{Code: 10}
	bi := new(big.Int)
	bi.SetString("111111111111111111111111111111111111111", 10)
	br := new(big.Rat)
	br.SetString("11111111111111111111111111111.11111111111111111")
	bf := new(big.Float)
	bf.SetString("1.111111111e+32")

	c := &testC{}
	if err := c.Load(c, nil); err != nil {
		t.Error(err)
	}
	c.i.Set(-256 * 256 * 256 * 256 * 256 * 256 * 256 * 256 / 2)
	c.i8.Set(-256 / 2)
	c.i16.Set(-256 * 256 / 2)
	c.i32.Set(-256 * 256 * 256 * 256 / 2)
	c.i64.Set(-256 * 256 * 256 * 256 * 256 * 256 * 256 * 256 / 2)
	c.u.Set(256*256*256*256*256*256*256*256 - 1)
	c.u8.Set(256 - 1)
	c.u16.Set(256*256 - 1)
	c.u32.Set(256*256*256*256 - 1)
	c.u64.Set(256*256*256*256*256*256*256*256 - 1)
	c.f32.Set(1.234234)
	c.f64.Set(-2.423904234324)
	c.bi.SetBig(bi)
	c.br.SetBig(br)
	c.bf.SetBig(bf)
	c.b.Set(false)
	c.s.Set("string")
	c.si.Set([]int{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.si8.Set([]int8{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.si16.Set([]int16{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.si32.Set([]int32{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.si64.Set([]int64{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.su.Set([]uint{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.su8.Set([]uint8{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.su16.Set([]uint16{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.su32.Set([]uint32{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.su64.Set([]uint64{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.st.Set([]int64{123423123, 123312321})
	c.t.Set(ts)
	c.j.Set(kickOff)
	c.p.Set(kickOff)

	c.d.i.Set(-256 * 256 * 256 * 256 * 256 * 256 * 256 * 256 / 2)
	c.d.i8.Set(-256 / 2)
	c.d.i16.Set(-256 * 256 / 2)
	c.d.i32.Set(-256 * 256 * 256 * 256 / 2)
	c.d.i64.Set(-256 * 256 * 256 * 256 * 256 * 256 * 256 * 256 / 2)
	c.d.si64.Set([]int64{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.u.Set(256*256*256*256*256*256*256*256 - 1)
	c.d.u8.Set(256 - 1)
	c.d.u16.Set(256*256 - 1)
	c.d.u32.Set(256*256*256*256 - 1)
	c.d.u64.Set(256*256*256*256*256*256*256*256 - 1)
	c.d.f32.Set(1.234234)
	c.d.f64.Set(-2.423904234324)
	c.d.bi.SetBig(bi)
	c.d.br.SetBig(br)
	c.d.bf.SetBig(bf)
	c.d.b.Set(false)
	c.d.s.Set("string")
	c.d.si.Set([]int{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.si8.Set([]int8{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.si16.Set([]int16{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.si32.Set([]int32{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.si64.Set([]int64{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.su.Set([]uint{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.su8.Set([]uint8{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.su16.Set([]uint16{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.su32.Set([]uint32{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.su64.Set([]uint64{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.st.Set([]int64{123423123, 123312321})
	c.d.t.Set(ts)
	c.d.j.Set(kickOff)
	c.d.p.Set(kickOff)

	dump, err := c.Dump()
	if err != nil {
		t.Error(err)
	}
	c2 := &testC{}
	if err = c2.Load(c2, dump); err != nil {
		t.Error(err)
	}
	if c.i.Get() != c2.i.Get() {
		t.Error("i not equal")
	}
	if c.i8.Get() != c2.i8.Get() {
		t.Error("i8 not equal")
	}
	if c.i16.Get() != c2.i16.Get() {
		t.Error("i16 not equal")
	}
	if c.i32.Get() != c2.i32.Get() {
		t.Error("i32 not equal")
	}
	if c.i64.Get() != c2.i64.Get() {
		t.Error("i64 not equal")
	}
	if c.u.Get() != c2.u.Get() {
		t.Error("u not equal")
	}
	if c.u8.Get() != c2.u8.Get() {
		t.Error("u8 not equal")
	}
	if c.u16.Get() != c2.u16.Get() {
		t.Error("u16 not equal")
	}
	if c.u32.Get() != c2.u32.Get() {
		t.Error("u32 not equal")
	}
	if c.u64.Get() != c2.u64.Get() {
		t.Error("u64 not equal")
	}
	if c.f32.Get() != c2.f32.Get() {
		t.Error("f32 not equal")
	}
	if c.f64.Get() != c2.f64.Get() {
		t.Error("f64 not equal")
	}
	if c.bi.GetBig().Cmp(bi) != 0 {
		t.Error("bi not equal")
	}
	if c.br.GetBig().Cmp(br) != 0 {
		t.Error("br not equal")
	}
	if c.bf.GetBig().Cmp(bf) != 0 {
		t.Error("bf not equal")
	}
	if c.t.Get() != c2.t.Get() {
		t.Error("time not equal")
	}
	if c.b.Get() != c2.b.Get() {
		t.Error("bool not equal")
	}
	if c.s.Get() != c2.s.Get() {
		t.Error("string not equal")
	}
	if fmt.Sprint(c.si.Get()) != fmt.Sprint(c2.si.Get()) {
		t.Error("si not equal")
	}
	if fmt.Sprint(c.si8.Get()) != fmt.Sprint(c2.si8.Get()) {
		t.Error("si8 not equal")
	}
	if fmt.Sprint(c.si16.Get()) != fmt.Sprint(c2.si16.Get()) {
		t.Error("si16 not equal")
	}
	if fmt.Sprint(c.si32.Get()) != fmt.Sprint(c2.si32.Get()) {
		t.Error("si32 not equal")
	}
	if fmt.Sprint(c.si64.Get()) != fmt.Sprint(c2.si64.Get()) {
		t.Error("si64 not equal")
	}
	if fmt.Sprint(c.su.Get()) != fmt.Sprint(c2.su.Get()) {
		t.Error("su not equal")
	}
	if fmt.Sprint(c.su8.Get()) != fmt.Sprint(c2.su8.Get()) {
		t.Error("su8 not equal")
	}
	if fmt.Sprint(c.su16.Get()) != fmt.Sprint(c2.su16.Get()) {
		t.Error("su16 not equal")
	}
	if fmt.Sprint(c.su32.Get()) != fmt.Sprint(c2.su32.Get()) {
		t.Error("su32 not equal")
	}
	if fmt.Sprint(c.su64.Get()) != fmt.Sprint(c2.su64.Get()) {
		t.Error("su64 not equal")
	}
	if fmt.Sprint(c.st.Get()) != fmt.Sprint(c2.st.Get()) {
		t.Error("st not equal")
	}
	j := &cli.KickPush{}
	j2 := &cli.KickPush{}
	c.j.Get(j)
	c2.j.Get(j2)
	if !reflect.DeepEqual(j, j2) {
		t.Error("JSON not equal")
	}
	p := &cli.KickPush{}
	p2 := &cli.KickPush{}
	c.p.Get(p)
	c2.p.Get(p2)
	if !reflect.DeepEqual(p, p2) {
		t.Error("proto not equal")
	}
	if c.d.i.Get() != c2.d.i.Get() {
		t.Error("d.i not equal")
	}
	if c.d.i8.Get() != c2.d.i8.Get() {
		t.Error("d.i8 not equal")
	}
	if c.d.i16.Get() != c2.d.i16.Get() {
		t.Error("d.i16 not equal")
	}
	if c.d.i32.Get() != c2.d.i32.Get() {
		t.Error("d.i32 not equal")
	}
	if c.d.i64.Get() != c2.d.i64.Get() {
		t.Error("d.i64 not equal")
	}
	if fmt.Sprint(c.d.si64.Get()) != fmt.Sprint(c2.d.si64.Get()) {
		t.Error("d.si64 not equal")
	}
	if c.d.u.Get() != c2.d.u.Get() {
		t.Error("d.u not equal")
	}
	if c.d.u8.Get() != c2.d.u8.Get() {
		t.Error("d.u8 not equal")
	}
	if c.d.u16.Get() != c2.d.u16.Get() {
		t.Error("d.u16 not equal")
	}
	if c.d.u32.Get() != c2.d.u32.Get() {
		t.Error("d.u32 not equal")
	}
	if c.d.u64.Get() != c2.d.u64.Get() {
		t.Error("d.u64 not equal")
	}
	if c.d.f32.Get() != c2.d.f32.Get() {
		t.Error("d.f32 not equal")
	}
	if c.d.f64.Get() != c2.d.f64.Get() {
		t.Error("d.f64 not equal")
	}
	if c.d.bi.GetBig().Cmp(bi) != 0 {
		t.Error("d.bi not equal")
	}
	if c.d.br.GetBig().Cmp(br) != 0 {
		t.Error("d.br not equal")
	}
	if c.d.bf.GetBig().Cmp(bf) != 0 {
		t.Error("d.bf not equal")
	}
	if c.d.b.Get() != c2.d.b.Get() {
		t.Error("d.bool not equal")
	}
	if c.d.s.Get() != c2.d.s.Get() {
		t.Error("d.string not equal")
	}
	if c.d.t.Get() != c2.d.t.Get() {
		t.Error("d.time not equal")
	}
	if fmt.Sprint(c.d.si.Get()) != fmt.Sprint(c2.d.si.Get()) {
		t.Error("d.si not equal")
	}
	if fmt.Sprint(c.d.si8.Get()) != fmt.Sprint(c2.d.si8.Get()) {
		t.Error("d.si8 not equal")
	}
	if fmt.Sprint(c.d.si16.Get()) != fmt.Sprint(c2.d.si16.Get()) {
		t.Error("d.si16 not equal")
	}
	if fmt.Sprint(c.d.si32.Get()) != fmt.Sprint(c2.d.si32.Get()) {
		t.Error("d.si32 not equal")
	}
	if fmt.Sprint(c.d.si64.Get()) != fmt.Sprint(c2.d.si64.Get()) {
		t.Error("d.si64 not equal")
	}
	if fmt.Sprint(c.d.su.Get()) != fmt.Sprint(c2.d.su.Get()) {
		t.Error("d.su not equal")
	}
	if fmt.Sprint(c.d.su8.Get()) != fmt.Sprint(c2.d.su8.Get()) {
		t.Error("d.su8 not equal")
	}
	if fmt.Sprint(c.d.su16.Get()) != fmt.Sprint(c2.d.su16.Get()) {
		t.Error("d.su16 not equal")
	}
	if fmt.Sprint(c.d.su32.Get()) != fmt.Sprint(c2.d.su32.Get()) {
		t.Error("d.su32 not equal")
	}
	if fmt.Sprint(c.d.su64.Get()) != fmt.Sprint(c2.d.su64.Get()) {
		t.Error("d.su64 not equal")
	}
	if fmt.Sprint(c.d.st.Get()) != fmt.Sprint(c2.d.st.Get()) {
		t.Error("d.st not equal")
	}
	dj := &cli.KickPush{}
	dj2 := &cli.KickPush{}
	c.d.j.Get(dj)
	c2.d.j.Get(dj2)
	if !reflect.DeepEqual(dj, dj2) {
		t.Error("d.JSON not equal")
	}
	dp := &cli.KickPush{}
	dp2 := &cli.KickPush{}
	c.d.p.Get(dp)
	c2.d.p.Get(dp2)
	if !reflect.DeepEqual(dp, dp2) {
		t.Error("d.proto not equal")
	}
}

func TestRecover(t *testing.T) {
	ts := virtualtime.Now().Unix()

	kickOff := &cli.KickPush{Code: 10}
	c := &testC{}
	if err := c.Load(c, nil); err != nil {
		t.Error(err)
	}
	c.i8.Set(-256 / 2)
	c.i16.Set(-256 * 256 / 2)
	c.i32.Set(-256 * 256 * 256 * 256 / 2)
	c.i64.Set(-256 * 256 * 256 * 256 * 256 * 256 * 256 * 256 / 2)
	c.u8.Set(256 - 1)
	c.u16.Set(256*256 - 1)
	c.u32.Set(256*256*256*256 - 1)
	c.u64.Set(256*256*256*256*256*256*256*256 - 1)
	c.f32.Set(1.234234)
	c.f64.Set(-2.423904234324)
	c.b.Set(false)
	c.s.Set("string")
	c.si.Set([]int{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.si8.Set([]int8{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.si16.Set([]int16{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.si32.Set([]int32{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.si64.Set([]int64{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.su.Set([]uint{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.su8.Set([]uint8{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.su16.Set([]uint16{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.su32.Set([]uint32{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.su64.Set([]uint64{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.st.Set([]int64{123423123, 123312321})
	c.t.Set(ts)
	c.j.Set(kickOff)
	c.p.Set(kickOff)

	c.d.i8.Set(-256 / 2)
	c.d.i16.Set(-256 * 256 / 2)
	c.d.i32.Set(-256 * 256 * 256 * 256 / 2)
	c.d.i64.Set(-256 * 256 * 256 * 256 * 256 * 256 * 256 * 256 / 2)
	c.d.u8.Set(256 - 1)
	c.d.u16.Set(256*256 - 1)
	c.d.u32.Set(256*256*256*256 - 1)
	c.d.u64.Set(256*256*256*256*256*256*256*256 - 1)
	c.d.f32.Set(1.234234)
	c.d.f64.Set(-2.423904234324)
	c.d.b.Set(false)
	c.d.s.Set("string")
	c.d.si.Set([]int{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.si8.Set([]int8{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.si16.Set([]int16{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.si32.Set([]int32{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.si64.Set([]int64{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.su.Set([]uint{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.su8.Set([]uint8{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.su16.Set([]uint16{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.su32.Set([]uint32{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.su64.Set([]uint64{1, 2, 3, 4, 5, 6, 6, 76, 2, 32, 45, 5, 2})
	c.d.st.Set([]int64{123423123, 123312321})
	c.d.t.Set(ts)
	c.d.j.Set(kickOff)
	c.d.p.Set(kickOff)

	if err := c.Revert(); err != nil {
		t.Error(err)
	}
	if c.i8.Get() != 0 {
		t.Error("i8 not equal to the value in bak")
	}
	if c.i16.Get() != 0 {
		t.Error("i16 not equal to the value in bak")
	}
	if c.i32.Get() != 0 {
		t.Error("i32 not equal to the value in bak")
	}
	if c.i64.Get() != 0 {
		t.Error("i64 not equal to the value in bak")
	}
	if c.u8.Get() != 0 {
		t.Error("u8 not equal to the value in bak")
	}
	if c.u16.Get() != 0 {
		t.Error("u16 not equal to the value in bak")
	}
	if c.u32.Get() != 0 {
		t.Error("u32 not equal to the value in bak")
	}
	if c.u64.Get() != 0 {
		t.Error("u64 not equal to the value in bak")
	}
	if c.f32.Get() != 0 {
		t.Error("f32 not equal to the value in bak")
	}
	if c.f64.Get() != 0 {
		t.Error("f64 not equal to the value in bak")
	}
	if c.b.Get() != false {
		t.Error("b not equal to the value in bak")
	}
	if c.s.Get() != "" {
		t.Error("s not equal to the value in bak")
	}
	if fmt.Sprint(c.si.Get()) != fmt.Sprint([]int{}) {
		t.Error("si not equal to the value in bak")
	}
	if fmt.Sprint(c.si8.Get()) != fmt.Sprint([]int8{}) {
		t.Error("si64 not equal to the value in bak")
	}
	if fmt.Sprint(c.si16.Get()) != fmt.Sprint([]int16{}) {
		t.Error("si64 not equal to the value in bak")
	}
	if fmt.Sprint(c.si32.Get()) != fmt.Sprint([]int32{}) {
		t.Error("si64 not equal to the value in bak")
	}
	if fmt.Sprint(c.si64.Get()) != fmt.Sprint([]int64{}) {
		t.Error("si64 not equal to the value in bak")
	}
	if fmt.Sprint(c.su.Get()) != fmt.Sprint([]uint{}) {
		t.Error("su not equal to the value in bak")
	}
	if fmt.Sprint(c.su8.Get()) != fmt.Sprint([]uint{}) {
		t.Error("su8 not equal to the value in bak")
	}
	if fmt.Sprint(c.su16.Get()) != fmt.Sprint([]uint16{}) {
		t.Error("su16 not equal to the value in bak")
	}
	if fmt.Sprint(c.su32.Get()) != fmt.Sprint([]uint32{}) {
		t.Error("su32 not equal to the value in bak")
	}
	if fmt.Sprint(c.su64.Get()) != fmt.Sprint([]uint64{}) {
		t.Error("su64 not equal to the value in bak")
	}
	if fmt.Sprint(c.st.Get()) != fmt.Sprint([]string{}) {
		t.Error("st not equal to the value in bak")
	}
	if c.t.Get() != 0 {
		t.Error("s not equal to the value in bak")
	}

	j := &cli.KickPush{}
	j2 := &cli.KickPush{}
	c.j.Get(j2)
	if !reflect.DeepEqual(j, j2) {
		t.Error("j not equal to the value in bak")
	}

	p := &cli.KickPush{}
	p2 := &cli.KickPush{}
	c.p.Get(p2)
	t.Log(p, p2)
	if !reflect.DeepEqual(p, p2) {
		t.Error("p not equal to the value in bak")
	}

	if c.d.i8.Get() != 0 {
		t.Error("d.i8 not equal to the value in bak")
	}
	if c.d.i16.Get() != 0 {
		t.Error("d.i16 not equal to the value in bak")
	}
	if c.d.i32.Get() != 0 {
		t.Error("d.i32 not equal to the value in bak")
	}
	if c.d.i64.Get() != 0 {
		t.Error("d.i64 not equal to the value in bak")
	}
	if fmt.Sprint(c.d.si64.Get()) != fmt.Sprint([]int64{}) {
		t.Error("d.si64 not equal to the value in bak")
	}
	if c.d.u8.Get() != 0 {
		t.Error("d.u8 not equal to the value in bak")
	}
	if c.d.u16.Get() != 0 {
		t.Error("d.u16 not equal to the value in bak")
	}
	if c.d.u32.Get() != 0 {
		t.Error("d.u32 not equal to the value in bak")
	}
	if c.d.u64.Get() != 0 {
		t.Error("d.u64 not equal to the value in bak")
	}
	if c.d.f32.Get() != 0 {
		t.Error("d.f32 not equal to the value in bak")
	}
	if c.d.f64.Get() != 0 {
		t.Error("d.f64 not equal to the value in bak")
	}
	if c.d.b.Get() != false {
		t.Error("d.b not equal to the value in bak")
	}
	if c.d.s.Get() != "" {
		t.Error("d.s not equal to the value in bak")
	}
	if fmt.Sprint(c.d.si.Get()) != fmt.Sprint([]int{}) {
		t.Error("d.si not equal to the value in bak")
	}
	if fmt.Sprint(c.d.si8.Get()) != fmt.Sprint([]int8{}) {
		t.Error("d.si8 not equal to the value in bak")
	}
	if fmt.Sprint(c.d.si16.Get()) != fmt.Sprint([]int16{}) {
		t.Error("d.si16 not equal to the value in bak")
	}
	if fmt.Sprint(c.d.si32.Get()) != fmt.Sprint([]int32{}) {
		t.Error("d.si32 not equal to the value in bak")
	}
	if fmt.Sprint(c.d.si64.Get()) != fmt.Sprint([]int64{}) {
		t.Log(c.d.si64.Get(), fmt.Sprint([]int64{}))
		t.Error("d.si64 not equal to the value in bak")
	}
	if fmt.Sprint(c.d.su.Get()) != fmt.Sprint([]uint{}) {
		t.Error("d.su not equal to the value in bak")
	}
	if fmt.Sprint(c.d.su8.Get()) != fmt.Sprint([]uint{}) {
		t.Error("d.su8 not equal to the value in bak")
	}
	if fmt.Sprint(c.d.su16.Get()) != fmt.Sprint([]uint16{}) {
		t.Error("d.su16 not equal to the value in bak")
	}
	if fmt.Sprint(c.d.su32.Get()) != fmt.Sprint([]uint32{}) {
		t.Error("d.su32 not equal to the value in bak")
	}
	if fmt.Sprint(c.d.su64.Get()) != fmt.Sprint([]uint64{}) {
		t.Error("d.su64 not equal to the value in bak")
	}
	if fmt.Sprint(c.d.st.Get()) != fmt.Sprint([]string{}) {
		t.Error("d.st not equal to the value in bak")
	}
	if c.d.t.Get() != 0 {
		t.Error("d.s not equal to the value in bak")
	}

	dj := &cli.KickPush{}
	dj2 := &cli.KickPush{}
	c.d.j.Get(dj2)
	if !reflect.DeepEqual(dj, dj2) {
		t.Error("dj not equal to the value in bak")
	}

	dp := &cli.KickPush{}
	dp2 := &cli.KickPush{}
	c.d.p.Get(dp2)
	t.Log(dp, dp2)
	if !reflect.DeepEqual(dp, dp2) {
		t.Error("d.p not equal to the value in bak")
	}
}
