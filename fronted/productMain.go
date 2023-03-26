package main

import (
	"github.com/kataras/iris/v12"
)

func main() {
	//1.创建iris 实例
	app := iris.New()
	//4.设置模板
	app.HandleDir("/public", "./fronted/web/public")
	//访问生成好的html静态文件
	app.HandleDir("/html", "./fronted/web/htmlProductShow")

	app.Run(
		iris.Addr("0.0.0.0:8083"),
		//iris.WithoutVersionChecker,
		iris.WithoutServerError(iris.ErrServerClosed), //如果web服务器 出现异常 我们将返回nil
		iris.WithOptimizations,                        //开启优化
	)

}
