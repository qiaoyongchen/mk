package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"

	"mk/ast"
)

const (
	NULL_OBJ         = "NULL"         // null
	INTEGER_OBJ      = "INTEGER"      // 整型
	BOOLEAN_OBJ      = "BOOLEAN"      // 布尔
	RETURN_VALUE_OBJ = "RETURN_VALUE" // return
	ERROR_OBJ        = "ERROR"        // error
	FUNCTION_OBJ     = "FUNCTION"     // user defined function
	STRING_OBJ       = "STRING"       // string
	BUILTIN_OBJ      = "BUILTIN"      // buildin function
	ARRAY_OBJ        = "ARRAY"
	HASH_OBJ         = "HASH"
)

type ObjectType string

type Object interface {
	Type() ObjectType // 类型
	Inspect() string  // 检查
}

type Hashable interface {
	HashKey() HashKey
}

// 整数类型
type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

//布尔类型
type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}

// 空指针类型,哈哈
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (N *Null) Inspect() string  { return "null" }

// return值(可包含任何类型的值)
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}

// 错误类型
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// 函数类型
// 因为该语音支持闭包
// 所以需要带上函数定义时的环境
type Function struct {
	Parameters []*ast.Identifier   //语法树里面的变量
	Body       *ast.BlockStatement //语法树里面的方法体
	Env        *Environment        //函数定义时的环境
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}

	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

// 字符串
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// 内置函数
type Builtin struct {
	Fn BuiltinFunction
}
type BuiltinFunction func(args ...Object) Object

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin funciton" }

// 数组
type Array struct {
	Elements []Object //包含任何类型的列表
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// 用于Hash.Pairs中的key
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// 单个的 k - v 对
type HashPair struct {
	Key   Object
	Value Object
}

// map 类型
type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}

	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
