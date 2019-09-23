package evaluator

import (
	"fmt"

	"mk/ast"
	"mk/object"
	//"mk/token"
)

var (
	NULL  = &object.Null{}                // null
	TRUE  = &object.Boolean{Value: true}  // true
	FALSE = &object.Boolean{Value: false} // false
)

// 通过 GO 类型 系统的true/false值
// 返回全局构造的object.TRUE/object.FALSE
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// 执行 Node (Statement | Expression)
// 新增一个执行中环境,用于关联变量
func Eval(node ast.Node, env *object.Environment) object.Object {

	switch node := node.(type) {
	// 语句列表
	case *ast.Program:
		return evalProgram(node, env)

	// 表达式语句
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	// 需要检查一下
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	// 整型
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	// 布尔类型
	case *ast.Boolean:
		// 返回全局的引用
		return nativeBoolToBooleanObject(node.Value)

	// 字符串
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	//前缀表达式
	case *ast.PrefixExpression:
		// 这里传进来的可能是很多奇怪的东西(boolen, integer, null ....)
		// 弱类型语言需要兼容这些
		// 所以先把执行出来结果再进行前缀操作
		right := Eval(node.Right, env)

		if isError(right) {
			return right
		}

		return evalPrefix(node.Operator, right)

	// 中缀表达式
	// 先分别求出左，右表达式再进行计算
	case *ast.InfixExpression:
		left := Eval(node.Left, env)

		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)

		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)

	// if 类型表达式
	case *ast.IfExpression:
		return evalIfExpression(node, env)

	// return 语句
	// 返回return类型值
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)

		if isError(val) {
			return val
		}

		return &object.ReturnValue{Value: val}

	// let语句在环境中给变量赋值
	// let语句的返回值就是变量代表的表达式的值
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		return env.Set(node.Name.Value, val)

	// 执行标识符的时候,需要传入环境
	// 在环境中取值然后执行
	case *ast.Identifier:
		return evalIdentifer(node, env)

	// 定义函数
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}

	// 调用函数
	case *ast.CallExpression:
		// 解析出object.Function类型
		function := Eval(node.Function, env)

		if isError(function) {
			return function
		}

		// 运行参数表达式,解析[]object.Object做为参数
		args := evalExpressions(node.Arguments, env)

		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)

	// 解析数组
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	// 解析下标
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	// 解析map类型
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}

	return nil
}

// 使方法作用于参数
func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {

	// 用户定义函数
	case *object.Function:
		extendEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendEnv)
		return unwrapReturnValue(evaluated)

	// 内置函数
	case *object.Builtin:
		return fn.Fn(args...)

	//
	default:
		return newError("not a function %s", fn.Type())
	}
}

// 以函数结构体环境为外环境(函数定义时的环境,定义时传入)
// 以当前参数组成的环境为内环境
// 返回一个新的函数运行时环境
func extendFunctionEnv(fn *object.Function,
	args []object.Object) *object.Environment {

	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

// 剥离return值的包裹
// 不然的话, return结果会一直上抛到最外层
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

// 解析表达式列表
// 用于解析函数的参数列表
// 和数组中表达式列表
func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {

	// 解析表达式的结果列表
	var result []object.Object

	// 挨个解析表达式,并加入到结果列表中
	for _, e := range exps {
		// 执行表达式
		evaluated := Eval(e, env)

		// 执行错误直接返回
		if isError(evaluated) {
			return []object.Object{evaluated}
		}

		result = append(result, evaluated)
	}
	return result
}

//
func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {

		// 如果是return类型,直接返回值
		case *object.ReturnValue:
			return result.Value

		// 如果是错误类型,直接返回错误
		case *object.Error:
			return result
		}
	}
	return result
}

// 解析语句列表
// 返回最后一个语句的值
func evalStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)
	}
	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement, env)

		// 如果是renturn类型的值的话
		// 直接返回return类型,调用方在收到return类型的返回时也会直接return
		// 因为有些语句会嵌套执行,提前返回
		// 例如:if (10 > 1) {if (10 > 1) {return 10;} return 1;};

		// ----------------- 调试专用 -------------------
		// fmt.Printf("\n")
		// println(statement.String())
		// fmt.Printf("%T", result)
		// fmt.Printf("\n")
		// ----------------- 调试专用结束-----------------

		if result.Type() == object.RETURN_VALUE_OBJ ||
			result.Type() == object.ERROR_OBJ {
			return result
		}
	}
	return result
}

// 解析前缀表达式
func evalPrefix(operator string, right object.Object) object.Object {
	switch operator {

	case "!":
		return evalBangOperatorExpression(right)

	case "-":
		return evalMinusPrefixOperatorExpression(right)

	// 如果暂时无法处理,返回一个错误
	default:
		return newError("unknown operator: %s %s", operator, right.Type())
	}
}

// 解析'!'前缀表达式
// true => false, false => true, null => true, other => false
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

// 解析'-'前缀表达式
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {

	if right.Type() != object.INTEGER_OBJ {
		return newError("unknon operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

// 解析中缀表达式
func evalInfixExpression(operator string, left object.Object,
	right object.Object) object.Object {

	switch {

	// 左右都是数值类型
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)

	// 左右都是string类型
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)

	// "==" 还能判断更多的类型,比如boolean
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)

	// "!=" 还能判断更多的类型,比如boolean
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)

	// 如果暂时无法处理,返回一个错误
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator,
			right.Type())
	}
}

// 解析处理integer类型的中缀表达式
func evalIntegerInfixExpression(operator string,
	left object.Object, right object.Object) object.Object {

	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {

	case "+":
		return &object.Integer{Value: leftVal + rightVal}

	case "-":
		return &object.Integer{Value: leftVal - rightVal}

	case "*":
		return &object.Integer{Value: leftVal * rightVal}

	case "/":
		return &object.Integer{Value: leftVal / rightVal}

	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)

	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)

	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)

	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)

	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator,
			right.Type())
	}
}

// 处理string类型中缀表达式
// 暂时只有连字符'+'
func evalStringInfixExpression(operator string, left, right object.Object) object.Object {

	if operator != "+" {
		return newError("unknow operator: %s %s %s",
			left.Type(), operator, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

// 解析if表达式
func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {

	// 执行条件表达式
	condition := Eval(ie.Condition, env)

	//println(condition.Type())
	//println(condition.Inspect())

	if isError(condition) {
		return condition
	}

	// 条件表达式为真,执行then部分
	if isTruthy(condition) {
		return Eval(ie.Consequence, env)

		// 条件表达式为假,执行else部分
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)

		// 其他情况直接返回null
	} else {
		return NULL
	}
}

// 检查object是不是boolean
// 需要兼容其他类型
func isTruthy(obj object.Object) bool {
	switch obj {

	case NULL:
		return false

	case TRUE:
		return true

	case FALSE:
		return false

	default:
		return true
	}
}

// 生成错误(辅助函数)
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// 检查是不是错误
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

// 运行标识符表达式
// 从环境中取值然后执行
// 添加内置函数后还需要查看标识符是不是内置函数的函数名
func evalIdentifer(node *ast.Identifier, env *object.Environment) object.Object {

	// 先搜索执行环境,查看执行环境中是否保存该值
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	// 再搜索内置方法
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	// 如果都查找不到则返回错误
	return newError("idenfier not found: " + node.Value)
}

// 解析下标表达式
func evalIndexExpression(left, index object.Object) object.Object {
	switch {

	// 左值是数组,index是数字,则解析的是数组表达式
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)

	// map类型没有要求,map类型的key可以是任何类型,只要HashKey()相同即可
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)

	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

// 解析数组类型下标表达式
func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)

	idx := index.(*object.Integer).Value

	// 检查下标是否越界
	max := int64(len(arrayObject.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

// 解析map类型
func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {

	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		// 因为key也可以是表达式,所以先执行获取key的值
		// 例如: let a = {11+22 : "33"};最终会被解析为{33: "33"}
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		// 检查下标是否可hash
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		// 执行value表达式
		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}
	return &object.Hash{Pairs: pairs}
}

// 解析map下标
func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	// 检查下标是否可hash
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	// 通过下标获取值
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}
