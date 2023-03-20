package main

import "github.com/kataras/iris/v12"

func main() {
	// 1.创建iris实例
	app := iris.New()
	// 2.设置错误模式，在mvc模式下提示错误
	app.Logger().SetLevel("debug")
	// 3.注册模板
	tmplate := iris.HTML("./backend/web/views", ".html").Layout("shared/layout.html").Reload(true)
	app.RegisterView(tmplate)
	// 4.设置目标模板
	//app.StaticWeb("/assets", "./backend/web/assets") // 域名+第一个参数路径，会跳转到下面的这个目录
	// StaticWeb在新版本已经废除
	app.HandleDir("/assets", "./backend/web/assets") // 域名+第一个参数路径，会跳转到下面的这个目录
	// 出现异常跳转到指定页面
	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "访问的页面出错"))
		ctx.ViewLayout("")
		ctx.View("shared/error.html")
	})
	// 5.注册控制器
	// 6.启动服务
	app.Run(
		iris.Addr("localhost:8080"),
		//iris.WithoutVersionChecker,   新版本中已经去掉了，可以不设置
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)
}
