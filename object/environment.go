package object

import ()

type Environment struct {
	store map[string]Object
	outer *Environment
}

// 一个环境就是一个map
// 用于一个key 和 一个 object 进行关联
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

// 通过传入A *Environment 新建 B *Environment
// A 在 B 的外层
// 通过这种方式模拟闭包: A 是函数定义时的外环境, B 是函数执行时的内环境
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// get : 先从自己找,找不到再向外层找
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// set
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
