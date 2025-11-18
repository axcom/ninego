package syscl

import (
	"bytes"
	"log"
	"strconv"

	"strings"
)

// 内嵌bytes.Buffer，支持连写
type Buffer struct {
	*bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{Buffer: new(bytes.Buffer)}
}

func (b *Buffer) Append(i interface{}) *Buffer {
	switch val := i.(type) {
	case int:
		b.append(strconv.Itoa(val))
	case int64:
		b.append(strconv.FormatInt(val, 10))
	case uint:
		b.append(strconv.FormatUint(uint64(val), 10))
	case uint64:
		b.append(strconv.FormatUint(val, 10))
	case string:
		b.append(val)
	case []byte:
		b.Write(val)
	case rune:
		b.WriteRune(val)
	}

	return b
}

func (b *Buffer) append(s string) *Buffer {
	defer func() {
		if err := recover(); err != nil {
			log.Println("*****内存不够了！******")
		}
	}()

	b.WriteString(s)
	return b
}

/*
strings.Builder和bytes.Buffer底层都是使用[]byte实现的， 但是性能测试的结果显示， 执行String()函数的时候，strings.Builder却比bytes.Buffer快很多。

区别就在于 bytes.Buffer 是重新申请了一块空间，存放生成的string变量， 而strings.Builder直接将底层的[]byte转换成了string类型返回了回来。

在bytes.Buffer中也说明了，如果想更有效率地(efficiently)构建字符串，请使用strings.Builder类型
*/

type Builder struct {
	*strings.Builder
}

func NewBuilder() *Builder {
	return &Builder{Builder: new(strings.Builder)}
}

func (b *Builder) Append(i interface{}) *Builder {
	switch val := i.(type) {
	case int:
		b.append(strconv.Itoa(val))
	case int64:
		b.append(strconv.FormatInt(val, 10))
	case uint:
		b.append(strconv.FormatUint(uint64(val), 10))
	case uint64:
		b.append(strconv.FormatUint(val, 10))
	case string:
		b.append(val)
	case []byte:
		b.Write(val)
	case rune:
		b.WriteRune(val)
	}

	return b
}

func (b *Builder) append(s string) *Builder {
	defer func() {
		if err := recover(); err != nil {
			log.Println("*****内存不够了！******")
		}
	}()

	b.WriteString(s)
	return b
}
