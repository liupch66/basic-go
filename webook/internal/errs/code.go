package errs

// 第一种定义方式
const (
	// CommonInvalidInput 任何模块都可以使用的错误码，兜底用的。最好还是设置自己的错误码
	CommonInvalidInput        = 400001
	CommonInternalServerError = 500001
)

// 第二种定义方式
// User 部分，模块代码使用 01
const (
	// UserInvalidInput 含糊的错误码，代表用户相关的 API 参数不对
	UserInvalidInput        = 401001
	UserInternalServerError = 501001

	// UserInvalidOrPassword 假如需要关心别的错误，可以进一步定义
	UserInvalidOrPassword = 401002
	UserDuplicateEmail    = 401003
)

// Article 部分，模块代码使用 02
const (
	ArticleInvalidInput     = 402001
	ArticleInterServerError = 502001
)

// 第三种定义方式

// Code Msg 是你 DEBUG 用的，不是给 C 端用户用的
type Code struct {
	Number int
	Msg    string
}

var (
	UserInvalidInputV1 = Code{Number: 401001, Msg: "用户输入错误"}
)
