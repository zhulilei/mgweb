package main

import (
	"github.com/astaxie/beego"
	//"github.com/astaxie/beego/logs"
	"html/template"
	"mgweb/app/controllers"
	"mgweb/app/jobs"
	_ "mgweb/app/mail"
	"mgweb/app/models"
	"net/http"
)

const VERSION = "1.0.0"

func main() {
	models.Init()
	jobs.InitJobs()

	// 设置默认404页面
	beego.ErrorHandler("404", func(rw http.ResponseWriter, r *http.Request) {
		t, _ := template.New("404.html").ParseFiles(beego.BConfig.WebConfig.ViewsPath + "/error/404.html")
		data := make(map[string]interface{})
		data["content"] = "page not found"
		t.Execute(rw, data)
	})

	// 生产环境不输出debug日志
	/*
		if beego.AppConfig.String("runmode") == "pro" {
			beego.SetLevel(beego.LevelInformational)
		} else {
			beego.SetLevel(beego.LevelInformational)
		}
	*/

	beego.AppConfig.Set("version", VERSION)

	/*
		beego.SetLogger("file", `{"filename":"./test.log"}`)
		beego.SetLevel(beego.LevelDebug)

		beego.Info("zhulilei")
		beego.Critical("zhulileib")
		log.Println()
	*/

	// 路由设置
	beego.Router("/", &controllers.MainController{}, "*:Index")
	beego.Router("/login", &controllers.MainController{}, "*:Login")
	beego.Router("/logout", &controllers.MainController{}, "*:Logout")
	beego.Router("/profile", &controllers.MainController{}, "*:Profile")
	beego.Router("/gettime", &controllers.MainController{}, "*:GetTime")
	beego.Router("/help", &controllers.HelpController{}, "*:Index")
	beego.AutoRouter(&controllers.TaskController{})
	beego.AutoRouter(&controllers.GroupController{})

	//beego.SetLevel(beego.LevelInformational)

	beego.BConfig.WebConfig.Session.SessionOn = true

	beego.Run()

}
