package web

import (
	"errors"
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/service"
)

// 确保 UserHandler 实现了 handle 接口
var _ handler = &UserHandler{}
var _ handler = (*UserHandler)(nil)

const (
	biz                  = "login"
	emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	phoneRegexPattern    = `^1[3-9]\d{9}$`
)

type UserHandler struct {
	emailRegexp    *regexp.Regexp
	passwordRegexp *regexp.Regexp
	phoneRegex     *regexp.Regexp
	userSvc        service.UserService
	codeSvc        service.CodeService
}

func NewUserHandler(userSvc service.UserService, codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		emailRegexp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		phoneRegex:     regexp.MustCompile(phoneRegexPattern, regexp.None),
		userSvc:        userSvc,
		codeSvc:        codeSvc,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	{
		ug.POST("/signup", u.Signup)
		// session 机制
		// ug.POST("/login", u.Login)
		// jwt 机制
		ug.POST("/login", u.LoginJWT)
		ug.POST("/edit", u.Edit)
		// ug.GET("/profile", u.Profile)
		ug.GET("/profile", u.ProfileJWT)
		ug.POST("/login_sms/code/send", u.SendSmsLoginCode)
		ug.POST("/login_sms", u.LoginSms)
	}
}

func (u *UserHandler) Signup(ctx *gin.Context) {
	type SignupReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignupReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 校验邮箱密码
	isEmail, err := u.emailRegexp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "邮箱格式错误")
		return
	}
	isPassword, err := u.passwordRegexp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须大于 8 位，并且包含数字，字母和特殊符号")
		return
	}
	// 校验两次输入密码是否一致
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}
	// 调用 service 层方法
	err = u.userSvc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		// 单独区分邮箱重复错误
		if errors.Is(err, service.ErrUserDuplicate) {
			ctx.String(http.StatusOK, "邮箱重复，请换一个邮箱")
		}
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "注册成功！")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	user, err := u.userSvc.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEmailOrPassword) {
			ctx.String(http.StatusOK, "邮箱或密码不对")
			return
		}
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// 登录成功，设置 session
	sess := sessions.Default(ctx)
	// 设置 session 的键值对， userId: user.Id
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		// HttpOnly	禁止 JavaScript 访问 Cookie		防止 XSS 窃取 Cookie
		// Secure	限制 Cookie 仅通过 HTTPS 传输		防止中间人攻击
		// Secure:   true,
		// HttpOnly: true,
		MaxAge: 60,
	})
	err = sess.Save()
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}

	ctx.String(http.StatusOK, "登录成功")
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserId int64
	// 利用 UserAgent 增强登录安全性
	UserAgent string
}

func (u *UserHandler) setJwtToken(ctx *gin.Context, userId int64) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		UserId:    userId,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("C%B|]SiozBE,S)X>ru,3Uu0+rl1Lj.@O"))
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	user, err := u.userSvc.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEmailOrPassword) {
			ctx.String(http.StatusOK, "邮箱或密码不对")
			return
		}
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// 设置登录态
	if err = u.setJwtToken(ctx, user.Id); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) Edit(ctx *gin.Context) {

}

func (u *UserHandler) Profile(ctx *gin.Context) {
	userId := sessions.Default(ctx).Get("userId")
	ctx.String(http.StatusOK, "This is your profile，your userId: %d", userId.(int64))
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	c, _ := ctx.Get("claims")
	// 忽略 exist 也行，因为下一步还有一个类型断言，nil 也会导致断言失败 !ok
	// if !exist {
	// 	ctx.String(http.StatusOK, "系统错误")
	// 	return
	// }
	claims, ok := c.(UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "This is your profile，your userId: %d", claims.UserId)
}

func (u *UserHandler) SendSmsLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	// 验证手机号码格式是否正确
	ok, err := u.phoneRegex.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "请输入正确的手机号码"})
		return
	}
	// 发送验证码
	err = u.codeSvc.Send(ctx, biz, req.Phone)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "发送成功"})
	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "发送验证码太频繁"})
	default:
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
	}
}

func (u *UserHandler) LoginSms(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	// 验证手机号码格式是否正确
	ok, err := u.phoneRegex.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "请输入正确的手机号码"})
		return
	}
	// 校验验证码
	ok, err = u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		if errors.Is(err, service.ErrCodeVerifyExpired) {
			ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "验证码已过期"})
		} else {
			ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		}
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "验证码错误"})
		return
	}
	// 验证成功设置登录态，要先拿到 userId
	user, err := u.userSvc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	err = u.setJwtToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "验证码验证成功"})
}
