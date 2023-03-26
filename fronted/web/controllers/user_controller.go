package controllers

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"go-shop-demo/datamodels"
	"go-shop-demo/services"
	"go-shop-demo/tool"
	"strconv"
)

type UserController struct {
	Ctx     iris.Context
	Service services.IUserService
	Session *sessions.Session
}

func (c *UserController) GetRegister() mvc.View {
	return mvc.View{
		Name: "user/register.html",
	}
}

func (c *UserController) PostRegister() {
	var (
		nickName = c.Ctx.FormValue("nickName")
		userName = c.Ctx.FormValue("userName")
		password = c.Ctx.FormValue("password")
	)
	user := &datamodels.User{
		NickName:     nickName,
		UserName:     userName,
		HashPassword: password,
	}
	_, err := c.Service.AddUser(user)
	if err != nil {
		c.Ctx.Redirect("/user/error")
		return
	}
	c.Ctx.Redirect("/user/login")
	return
}

func (c *UserController) GetLogin() mvc.View {
	return mvc.View{
		Name: "user/login.html",
	}
}

func (c *UserController) PostLogin() mvc.Response {
	// 1.获取用户提交的表单信息
	var (
		userName = c.Ctx.FormValue("userName")
		password = c.Ctx.FormValue("password")
	)
	// 2.验证账号密码正确
	user, isOk := c.Service.IsPwdSuccess(userName, password)
	if !isOk {
		return mvc.Response{
			Path: "/user/login",
		}
	}
	// 3.写入用户ID到cookie中
	tool.GlobalCookie(c.Ctx, "uid", strconv.FormatInt(user.ID, 10)) // 设置session
	c.Session.Set("userID", strconv.FormatInt(user.ID, 10))
	// 跳转页面，登录成功之后跳转产品页面
	return mvc.Response{
		Path: "/product/",
	}
}
