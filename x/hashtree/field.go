package hashtree

import "strings"

// Int returns field value as int
func (r *Root) Int(field string) *Int {
	return (*Int)(r.Field(strings.Split(field, ".")...))
}

// Int8 returns field value as int8
func (r *Root) Int8(field string) *Int8 {
	return (*Int8)(r.Field(strings.Split(field, ".")...))
}

// Int16 returns field value as int16
func (r *Root) Int16(field string) *Int16 {
	return (*Int16)(r.Field(strings.Split(field, ".")...))
}

// Int32 returns field value as int32
func (r *Root) Int32(field string) *Int32 {
	return (*Int32)(r.Field(strings.Split(field, ".")...))
}

// Int64 returns field value as int64
func (r *Root) Int64(field string) *Int64 {
	return (*Int64)(r.Field(strings.Split(field, ".")...))
}

// Uint returns field value as uint
func (r *Root) Uint(field string) *Uint {
	return (*Uint)(r.Field(strings.Split(field, ".")...))
}

// Uint8 returns field value as uint8
func (r *Root) Uint8(field string) *Uint8 {
	return (*Uint8)(r.Field(strings.Split(field, ".")...))
}

// Uint16 returns field value as uint16
func (r *Root) Uint16(field string) *Uint16 {
	return (*Uint16)(r.Field(strings.Split(field, ".")...))
}

// Uint32 returns field value as Uint32
func (r *Root) Uint32(field string) *Uint32 {
	return (*Uint32)(r.Field(strings.Split(field, ".")...))
}

// Uint64 returns field value as uint64
func (r *Root) Uint64(field string) *Uint64 {
	return (*Uint64)(r.Field(strings.Split(field, ".")...))
}

// Float32 returns field value as float32
func (r *Root) Float32(field string) *Float32 {
	return (*Float32)(r.Field(strings.Split(field, ".")...))
}

// Float64 returns field value as float64
func (r *Root) Float64(field string) *Float64 {
	return (*Float64)(r.Field(strings.Split(field, ".")...))
}

// BigInt returns field value as big.Int
func (r *Root) BigInt(field string) *BigInt {
	return (*BigInt)(r.Field(strings.Split(field, ".")...))
}

// BigRat returns field value as big.Rat
func (r *Root) BigRat(field string) *BigRat {
	return (*BigRat)(r.Field(strings.Split(field, ".")...))
}

// BigFloat returns field value as big.Float
func (r *Root) BigFloat(field string) *BigFloat {
	return (*BigFloat)(r.Field(strings.Split(field, ".")...))
}

// String returns field value as string
func (r *Root) String(field string) *String {
	return (*String)(r.Field(strings.Split(field, ".")...))
}

// SliceInt returns field value as int
func (r *Root) SliceInt(field string) *SliceInt {
	return (*SliceInt)(r.Field(strings.Split(field, ".")...))
}

// SliceInt8 returns field value as int8
func (r *Root) SliceInt8(field string) *SliceInt8 {
	return (*SliceInt8)(r.Field(strings.Split(field, ".")...))
}

// SliceInt16 returns field value as int16
func (r *Root) SliceInt16(field string) *SliceInt16 {
	return (*SliceInt16)(r.Field(strings.Split(field, ".")...))
}

// Int32 returns field value as int32
func (r *Root) SliceInt32(field string) *SliceInt32 {
	return (*SliceInt32)(r.Field(strings.Split(field, ".")...))
}

// SliceInt64 returns field value as int64
func (r *Root) SliceInt64(field string) *SliceInt64 {
	return (*SliceInt64)(r.Field(strings.Split(field, ".")...))
}

// SliceUint returns field value as uint
func (r *Root) SliceUint(field string) *SliceUint {
	return (*SliceUint)(r.Field(strings.Split(field, ".")...))
}

// SliceUint8 returns field value as uint8
func (r *Root) SliceUint8(field string) *SliceUint8 {
	return (*SliceUint8)(r.Field(strings.Split(field, ".")...))
}

// SliceUint16 returns field value as uint16
func (r *Root) SliceUint16(field string) *SliceUint16 {
	return (*SliceUint16)(r.Field(strings.Split(field, ".")...))
}

// SliceUint32 returns field value as Uint32
func (r *Root) SliceUint32(field string) *SliceUint32 {
	return (*SliceUint32)(r.Field(strings.Split(field, ".")...))
}

// SliceUint64 returns field value as uint64
func (r *Root) SliceUint64(field string) *SliceUint64 {
	return (*SliceUint64)(r.Field(strings.Split(field, ".")...))
}

// SliceFloat32 returns field value as uint64
func (r *Root) SliceFloat32(field string) *SliceFloat32 {
	return (*SliceFloat32)(r.Field(strings.Split(field, ".")...))
}

// SliceFloat64 returns field value as uint64
func (r *Root) SliceFloat64(field string) *SliceFloat64 {
	return (*SliceFloat64)(r.Field(strings.Split(field, ".")...))
}

// Time returns field value as time
func (r *Root) Time(field string) *Time {
	return (*Time)(r.Field(strings.Split(field, ".")...))
}

// Bool returns field value as Bool
func (r *Root) Bool(field string) *Bool {
	return (*Bool)(r.Field(strings.Split(field, ".")...))
}

// JSON returns field value as JSON
func (r *Root) JSON(field string) *JSON {
	return (*JSON)(r.Field(strings.Split(field, ".")...))
}

// Proto returns field value as Proto
func (r *Root) Proto(field string) *Proto {
	return (*Proto)(r.Field(strings.Split(field, ".")...))
}

// PropField returns field value as uint64
func (r *Root) PropField(field string) *PropField {
	return (*PropField)(r.Field(strings.Split(field, ".")...))
}

// SliceFloat64 returns field value as uint64
func (r *Root) PropBag(field string) *PropBag {
	return (*PropBag)(r.Field(strings.Split(field, ".")...))
}
