package skit

import (
	"fmt"
	"math"
	"time"
)

type Metadata struct {
	val interface{}
}

func NewMetadata(v interface{}) *Metadata {
	m := Metadata{}
	return m.Value(v)
}
func (m *Metadata) Value(v interface{}) *Metadata {
	m.val = v
	return m
}
func (m *Metadata) Int() int {
	return Int(m.val)
}
func (m *Metadata) Int64() int64 {
	return Int64(m.val)
}
func (m *Metadata) Int32() int32 {
	return Int32(m.val)
}
func (m *Metadata) Int16() int16 {
	return Int16(m.val)
}
func (m *Metadata) Int8() int8 {
	return Int8(m.val)
}
func (m *Metadata) Uint() uint {
	return Uint(m.val)
}
func (m *Metadata) Uint64() uint64 {
	return Uint64(m.val)
}
func (m *Metadata) Uint32() uint32 {
	return Uint32(m.val)
}
func (m *Metadata) Uint16() uint16 {
	return Uint16(m.val)
}
func (m *Metadata) Uint8() uint8 {
	return Uint8(m.val)
}
func (m *Metadata) Float() float64 {
	return Float64(m.val)
}
func (m *Metadata) Float32() float32 {
	return Float32(m.val)
}
func (m *Metadata) String() string {
	return String(m.val)
}
func (m *Metadata) Byte() byte {
	return Byte(m.val)
}
func (m *Metadata) Bytes() []byte {
	return Bytes(m.val)
}
func (m *Metadata) Rune() rune {
	return Rune(m.val)
}
func (m *Metadata) Runes() []rune {
	return Runes(m.val)
}
func (m *Metadata) Bool() bool {
	return Bool(m.val)
}
func (m *Metadata) Time() (t time.Time) {
	SetValue(&t, m.val)
	return
}
func (m *Metadata) DateTime() (t DateTime) {
	SetValue(&t, m.val)
	return
}

////////////////////////////////////////////////////////////////////////////////

const (
	TimeFormart = "2006-01-02 15:04:05"
)

type DateTime time.Time

func (dt DateTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf(`"%s"`, time.Time(dt).Format(TimeFormart))
	return []byte(stamp), nil
}
func (dt DateTime) String() string {
	stamp := fmt.Sprintf(`"%s"`, time.Time(dt).Format(TimeFormart))
	return stamp
}

type Money float64

func (m Money) Float64() float64 {
	return (math.Floor(float64(m)*10000+0.5) / 10000)
}
