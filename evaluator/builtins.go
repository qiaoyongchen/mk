package evaluator

import (
	"fmt"
	"time"

	"mk/object"
)

// 内置函数
var builtins = map[string]*object.Builtin{

	// 解析字符串长度
	// 解析数组长度
	// 解析map长度
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			switch arg := args[0].(type) {

			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}

			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}

			case *object.Hash:
				return &object.Integer{Value: int64(len(arg.Pairs))}

			default:
				return newError("argument to `len` not supported, got=%s",
					args[0].Type())
			}
		},
	},

	// 取数组第一个元素
	"first": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			// 限制参数个数
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			// 检查参数类型为 object.Array
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}

			// 强制转换
			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			// 默认返回NULL值
			return NULL
		},
	},

	// 取数组最后一个元素
	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			// 检查参数个数
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			// 检查类型
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)

			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return NULL
		},
	},

	// 去除第一个取剩余部分
	"rest": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			// 检查参数个数
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			// 检查参数类型
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)

			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]object.Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])
				return &object.Array{Elements: newElements}
			}

			return NULL
		},
	},

	// 压入一个值
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			// 检查参数个数
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			// 第一个参数为*object.Array
			// 第一个参数可以为任何值
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)

			length := len(arr.Elements)
			newElements := make([]object.Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &object.Array{Elements: newElements}
		},
	},

	// 打印任何值
	"puts": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},

	// 显示当前时间
	"now": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			// 检查参数个数
			if len(args) != 0 {
				return newError("too many parameters, expect :0, given :%d", len(args))
			}

			// 打印当前时间
			return &object.String{Value: time.Now().Format("2006-01-02 15:04:05")}
		},
	},
}
