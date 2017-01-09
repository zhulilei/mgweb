package controllers

import (
	"github.com/astaxie/beego"
	//libcron "github.com/lisijie/cron"
	libcron "mgweb/app/cron"
	"mgweb/app/jobs"
	"mgweb/app/libs"
	"mgweb/app/models"
	"strconv"
	"strings"
	"time"
)

type TaskController struct {
	BaseController
}

type TaskNginx struct {
	Id          int
	UserId      int
	GroupId     int
	TaskName    string
	TaskType    int
	ServerName  string
	Host        string
	Status      int
	Notify      int
	NotifyEmail string
	Timeout     int
}

// 任务列表
func (this *TaskController) List() {
	page, _ := this.GetInt("page")
	if page < 1 {
		page = 1
	}
	groupId, _ := this.GetInt("groupid")
	beego.Critical("groupId is", groupId)
	filters := make([]interface{}, 0)
	if groupId > 0 {
		filters = append(filters, "group_id", groupId)
	}
	result, count := models.TaskGetList(page, this.pageSize, filters...)

	list := make([]map[string]interface{}, len(result))
	for k, v := range result {
		row := make(map[string]interface{})
		row["id"] = v.Id
		row["name"] = v.TaskName
		row["cron_spec"] = v.CronSpec
		row["status"] = v.Status
		row["description"] = v.Description

		e := jobs.GetEntryById(v.Id)
		beego.Critical("nowgetentryis", e)
		if e != nil {
			temp_time := beego.Date(e.Next, "Y-m-d H:i:s")
			//row["next_time"] = beego.Date(e.Next, "Y-m-d H:i:s")
			row["next_time"] = temp_time
			row["prev_time"] = "-"
			if e.Prev.Unix() > 0 {
				row["prev_time"] = beego.Date(e.Prev, "Y-m-d H:i:s")
			} else if v.PrevTime > 0 {
				row["prev_time"] = beego.Date(time.Unix(v.PrevTime, 0), "Y-m-d H:i:s")
			}
			row["running"] = 1
			//beego.Info("next is", e.Next)
			//beego.Info("prev is", e.Prev)
		} else {
			//beego.Critical("next-time is", beego.Date(e.Next, "Y-m-d H:i:s"))
			//beego.Critical("prev-time is", beego.Date(e.Prev, "Y-m-d H:i:s"))
			row["next_time"] = "-"
			if v.PrevTime > 0 {
				row["prev_time"] = beego.Date(time.Unix(v.PrevTime, 0), "Y-m-d H:i:s")
			} else {
				row["prev_time"] = "-"
			}
			row["running"] = 0
			//beego.Info(e.Next)
			//beego.Info(e.Prev)
		}
		row["executetimes"] = v.ExecuteTimes
		list[k] = row
	}

	// 分组列表
	groups, _ := models.TaskGroupGetList(1, 100)

	this.Data["pageTitle"] = "任务列表"
	this.Data["list"] = list
	this.Data["groups"] = groups
	this.Data["groupid"] = groupId
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("TaskController.List", "groupid", groupId), true).ToString()
	this.display()
}

// 添加任务
func (this *TaskController) Add() {
	groupid, _ := this.GetInt("groupid")
	beego.Debug("Add group id is", groupid)
	group, err := models.TaskGroupGetById(groupid)
	if err != nil {
		this.showMsg(err.Error())
	}
	if group.GroupName == "CMD" {
		beego.Critical("Group is", group.GroupName)
		this.AddCMD(groupid)
	}
	if group.GroupName == "NGINX" {
		beego.Critical("Group is", group.GroupName)
		this.AddNGINX(groupid)
	}
	/*
		if this.isPost() {
			task := new(models.Task)
			task.UserId = this.userId
			task.GroupId, _ = this.GetInt("group_id")
			task.TaskName = strings.TrimSpace(this.GetString("task_name"))
			task.Description = strings.TrimSpace(this.GetString("description"))
			task.Concurrent, _ = this.GetInt("concurrent")
			task.CronSpec = strings.TrimSpace(this.GetString("cron_spec"))
			task.Command = strings.TrimSpace(this.GetString("command"))
			task.Notify, _ = this.GetInt("notify")
			task.Timeout, _ = this.GetInt("timeout")

			notifyEmail := strings.TrimSpace(this.GetString("notify_email"))
			if notifyEmail != "" {
				emailList := make([]string, 0)
				tmp := strings.Split(notifyEmail, "\n")
				for _, v := range tmp {
					v = strings.TrimSpace(v)
					if !libs.IsEmail([]byte(v)) {
						this.ajaxMsg("无效的Email地址："+v, MSG_ERR)
					} else {
						emailList = append(emailList, v)
					}
				}
				task.NotifyEmail = strings.Join(emailList, "\n")
			}

			if task.TaskName == "" || task.CronSpec == "" || task.Command == "" {
				this.ajaxMsg("请填写完整信息", MSG_ERR)
			}
			if _, err := libcron.Parse(task.CronSpec); err != nil {
				this.ajaxMsg("cron表达式无效", MSG_ERR)
			}
			if _, err := models.TaskAdd(task); err != nil {
				this.ajaxMsg(err.Error(), MSG_ERR)
			}

			this.ajaxMsg("", MSG_OK)
		}

		// 分组列表
		groups, _ := models.TaskGroupGetList(1, 100)
		this.Data["groups"] = groups
		this.Data["pageTitle"] = "添加任务"
		this.display()
	*/
}

//添加AddCMD
func (this *TaskController) AddCMD(groupid int) {
	beego.Critical("hahahahaa groupid")
	this.actionName = "addcmd"
	if this.isPost() {
		task := new(models.Task)
		task.UserId = this.userId
		//task.GroupId, _ = this.GetInt("group_id")
		task.GroupId = groupid
		task.TaskName = strings.TrimSpace(this.GetString("task_name"))
		task.Description = strings.TrimSpace(this.GetString("description"))
		task.Concurrent, _ = this.GetInt("concurrent")
		task.CronSpec = strings.TrimSpace(this.GetString("cron_spec"))
		task.Command = strings.TrimSpace(this.GetString("command"))
		task.Notify, _ = this.GetInt("notify")
		task.Timeout, _ = this.GetInt("timeout")

		notifyEmail := strings.TrimSpace(this.GetString("notify_email"))
		if notifyEmail != "" {
			emailList := make([]string, 0)
			tmp := strings.Split(notifyEmail, "\n")
			for _, v := range tmp {
				v = strings.TrimSpace(v)
				if !libs.IsEmail([]byte(v)) {
					this.ajaxMsg("无效的Email地址："+v, MSG_ERR)
				} else {
					emailList = append(emailList, v)
				}
			}
			task.NotifyEmail = strings.Join(emailList, "\n")
		}

		if task.TaskName == "" || task.CronSpec == "" || task.Command == "" {
			this.ajaxMsg("请填写完整信息", MSG_ERR)
		}
		if _, err := libcron.Parse(task.CronSpec); err != nil {
			this.ajaxMsg("cron表达式无效", MSG_ERR)
		}
		if _, err := models.TaskAdd(task); err != nil {
			this.ajaxMsg(err.Error(), MSG_ERR)
		}

		this.ajaxMsg("", MSG_OK)
	}

	// 分组列表
	beego.Critical("hahahahaa AddCMD")
	groups, _ := models.TaskGroupGetList(1, 100)
	this.Data["groups"] = groups
	this.Data["pageTitle"] = "添加任务"
	this.display()
}

//添加AddNGINX
func (this *TaskController) AddNGINX(groupid int) {
	beego.ReadFromRequest(&this.Controller)
	beego.Critical("hahahahaa groupid")
	this.actionName = "addnginx"

	if this.isPost() {
		tasknginx := &TaskNginx{}
		randomid, _ := strconv.Atoi(string(libs.Krand(5, libs.KC_RAND_KIND_NUM)))
		tasknginx.Id = randomid
		//Jd.rid = randomid
		//this.Data["taskid"] = Jd.rid
		tasknginx.UserId = this.userId
		tasknginx.GroupId = groupid
		tasknginx.TaskName = strings.TrimSpace(this.GetString("task_name"))
		tasknginx.ServerName = strings.TrimSpace(this.GetString("server_name"))
		tasknginx.Host = strings.TrimSpace(this.GetString("host"))
		tasknginx.Notify, _ = this.GetInt("notify")
		tasknginx.Timeout, _ = this.GetInt("timeout")

		beego.Critical("nginxtask is", tasknginx)
		/*
			task := new(models.Task)
			task.UserId = this.userId
			//task.GroupId, _ = this.GetInt("group_id")
			task.GroupId = groupid
			task.TaskName = strings.TrimSpace(this.GetString("task_name"))
			task.Description = strings.TrimSpace(this.GetString("description"))
			task.Concurrent, _ = this.GetInt("concurrent")
			task.CronSpec = strings.TrimSpace(this.GetString("cron_spec"))
			task.Command = strings.TrimSpace(this.GetString("command"))
			task.Notify, _ = this.GetInt("notify")
			task.Timeout, _ = this.GetInt("timeout")
		*/
		notifyEmail := strings.TrimSpace(this.GetString("notify_email"))
		if notifyEmail != "" {
			emailList := make([]string, 0)
			tmp := strings.Split(notifyEmail, "\n")
			for _, v := range tmp {
				v = strings.TrimSpace(v)
				if !libs.IsEmail([]byte(v)) {
					this.ajaxMsg("无效的Email地址："+v, MSG_ERR)
				} else {
					emailList = append(emailList, v)
				}
			}
			tasknginx.NotifyEmail = strings.Join(emailList, "\n")
		}

		if tasknginx.TaskName == "" || tasknginx.ServerName == "" {
			this.ajaxMsg("请填写完整信息", MSG_ERR)
		}
		//加入addcmd
		taskcmd := new(models.Task)
		taskcmd.UserId = tasknginx.UserId
		taskcmd.GroupId = tasknginx.GroupId
		taskcmd.TaskName = tasknginx.TaskName
		taskcmd.Description = tasknginx.Host
		taskcmd.CronSpec = "1s"
		taskcmd.Command = "nginx_find_server -P '/etc/nginx/sites-enabled' -S " + tasknginx.ServerName
		taskcmd.Notify = tasknginx.Notify
		taskcmd.Timeout = 10
		taskcmd.NotifyEmail = tasknginx.NotifyEmail
		taskcmd.Id = tasknginx.Id
		beego.Info("taskid is", taskcmd.Id)

		if _, err := models.TaskAdd(taskcmd); err != nil {
			this.ajaxMsg(err.Error(), MSG_ERR)
		}
		/*
			if _, err := libcron.Parse(task.CronSpec); err != nil {
				this.ajaxMsg("cron表达式无效", MSG_ERR)
			}
			if _, err := models.TaskAdd(task); err != nil {
				this.ajaxMsg(err.Error(), MSG_ERR)
			}
		*/
		beego.Critical("nginxtask is", tasknginx)
		this.StartNginx(randomid)
		this.redirect(beego.URLFor("TaskController.NginxSingleList", "id", randomid))
		//refer := beego.URLFor("TaskController.NginxSingleList")
		//this.redirect(refer)
		this.ajaxMsg("", MSG_OK)

	}

	// 分组列表
	beego.Critical("hahahahaa AddNGINX")
	groups, _ := models.TaskGroupGetList(1, 100)
	//this.Data["taskid"] = Jd.rid
	this.Data["groups"] = groups
	this.Data["pageTitle"] = "添加任务"
	this.display()
}

// nginx-single-list
func (this *TaskController) NginxSingleList() {
	this.actionName = "nginxsinglelist"
	id, _ := this.GetInt("id")
	beego.Info("i am in nginxsinglelist!!!!!!!!")
	beego.Info("id is", id)

	task, err := models.TaskGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	taskLog, err := models.TaskLogGetByTaskId(id)
	data := make(map[string]interface{})
	if err != nil {
		//this.showMsg(err.Error())
		data["status"] = "正在搜索"
		this.Data["data"] = data
	} else {
		data["id"] = taskLog.Id
		data["output"] = taskLog.Output
		data["error"] = taskLog.Error
		data["start_time"] = beego.Date(time.Unix(taskLog.CreateTime, 0), "Y-m-d H:i:s")
		data["process_time"] = float64(taskLog.ProcessTime) / 1000
		data["ouput_size"] = libs.SizeFormat(float64(len(taskLog.Output)))
		data["status"] = "查找好~(≧▽≦)/啦啦啦"

		this.Data["data"] = data
	}

	// 分组列表
	groups, _ := models.TaskGroupGetList(1, 100)
	this.Data["groups"] = groups
	this.Data["task"] = task
	this.Data["pageTitle"] = "编辑任务"
	this.display()
}

// 编辑任务
func (this *TaskController) Edit() {
	id, _ := this.GetInt("id")

	task, err := models.TaskGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	if this.isPost() {
		task.TaskName = strings.TrimSpace(this.GetString("task_name"))
		task.Description = strings.TrimSpace(this.GetString("description"))
		task.GroupId, _ = this.GetInt("group_id")
		task.Concurrent, _ = this.GetInt("concurrent")
		task.CronSpec = strings.TrimSpace(this.GetString("cron_spec"))
		task.Command = strings.TrimSpace(this.GetString("command"))
		task.Notify, _ = this.GetInt("notify")
		task.Timeout, _ = this.GetInt("timeout")

		notifyEmail := strings.TrimSpace(this.GetString("notify_email"))
		if notifyEmail != "" {
			tmp := strings.Split(notifyEmail, "\n")
			emailList := make([]string, 0, len(tmp))
			for _, v := range tmp {
				v = strings.TrimSpace(v)
				if !libs.IsEmail([]byte(v)) {
					this.ajaxMsg("无效的Email地址："+v, MSG_ERR)
				} else {
					emailList = append(emailList, v)
				}
			}
			task.NotifyEmail = strings.Join(emailList, "\n")
		}

		if task.TaskName == "" || task.CronSpec == "" || task.Command == "" {
			this.ajaxMsg("请填写完整信息", MSG_ERR)
		}
		if _, err := libcron.Parse(task.CronSpec); err != nil {
			this.ajaxMsg("cron表达式无效", MSG_ERR)
		}
		if err := task.Update(); err != nil {
			this.ajaxMsg(err.Error(), MSG_ERR)
		}

		this.ajaxMsg("", MSG_OK)
	}

	// 分组列表
	groups, _ := models.TaskGroupGetList(1, 100)
	this.Data["groups"] = groups
	this.Data["task"] = task
	this.Data["pageTitle"] = "编辑任务"
	this.display()
}

// 任务执行日志列表
func (this *TaskController) Logs() {
	taskId, _ := this.GetInt("id")
	page, _ := this.GetInt("page")
	if page < 1 {
		page = 1
	}

	task, err := models.TaskGetById(taskId)
	if err != nil {
		this.showMsg(err.Error())
	}

	result, count := models.TaskLogGetList(page, this.pageSize, "task_id", task.Id)

	list := make([]map[string]interface{}, len(result))
	for k, v := range result {
		row := make(map[string]interface{})
		row["id"] = v.Id
		row["start_time"] = beego.Date(time.Unix(v.CreateTime, 0), "Y-m-d H:i:s")
		row["process_time"] = float64(v.ProcessTime) / 1000
		row["ouput_size"] = libs.SizeFormat(float64(len(v.Output)))
		row["status"] = v.Status
		list[k] = row
	}

	this.Data["pageTitle"] = "任务执行日志"
	this.Data["list"] = list
	this.Data["task"] = task
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("TaskController.Logs", "id", taskId), true).ToString()
	this.display()
}

// 查看日志详情
func (this *TaskController) ViewLog() {
	id, _ := this.GetInt("id")

	taskLog, err := models.TaskLogGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	task, err := models.TaskGetById(taskLog.TaskId)
	if err != nil {
		this.showMsg(err.Error())
	}

	data := make(map[string]interface{})
	data["id"] = taskLog.Id
	data["output"] = taskLog.Output
	data["error"] = taskLog.Error
	data["start_time"] = beego.Date(time.Unix(taskLog.CreateTime, 0), "Y-m-d H:i:s")
	data["process_time"] = float64(taskLog.ProcessTime) / 1000
	data["ouput_size"] = libs.SizeFormat(float64(len(taskLog.Output)))
	data["status"] = taskLog.Status

	this.Data["task"] = task
	this.Data["data"] = data
	this.Data["pageTitle"] = "查看日志"
	this.display()
	/*
		for i := range [100]int{1} {
			time.Sleep(1 * time.Second)
			this.Ctx.WriteString(strconv.Itoa(i))
		}
	*/

}

// 批量操作日志
func (this *TaskController) LogBatch() {
	action := this.GetString("action")
	ids := this.GetStrings("ids")
	if len(ids) < 1 {
		this.ajaxMsg("请选择要操作的项目", MSG_ERR)
	}
	for _, v := range ids {
		id, _ := strconv.Atoi(v)
		if id < 1 {
			continue
		}
		switch action {
		case "delete":
			models.TaskLogDelById(id)
		}
	}

	this.ajaxMsg("", MSG_OK)
}

// 批量操作
func (this *TaskController) Batch() {
	action := this.GetString("action")
	ids := this.GetStrings("ids")
	if len(ids) < 1 {
		this.ajaxMsg("请选择要操作的项目", MSG_ERR)
	}

	for _, v := range ids {
		id, _ := strconv.Atoi(v)
		if id < 1 {
			continue
		}
		switch action {
		case "active":
			if task, err := models.TaskGetById(id); err == nil {
				job, err := jobs.NewJobFromTask(task)
				if err == nil {
					jobs.AddJob(task.CronSpec, job)
					task.Status = 1
					task.Update()
				}
			}
		case "pause":
			jobs.RemoveJob(id)
			if task, err := models.TaskGetById(id); err == nil {
				task.Status = 0
				task.Update()
			}
		case "delete":
			models.TaskDel(id)
			models.TaskLogDelByTaskId(id)
			jobs.RemoveJob(id)
		}
	}

	this.ajaxMsg("", MSG_OK)
}

// 启动任务
func (this *TaskController) Start() {
	id, _ := this.GetInt("id")

	task, err := models.TaskGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	job, err := jobs.NewJobFromTask(task)
	if err != nil {
		this.showMsg(err.Error())
	}

	if jobs.AddJob(task.CronSpec, job) {
		task.Status = 1
		task.Update()
	}

	refer := this.Ctx.Request.Referer()
	if refer == "" {
		refer = beego.URLFor("TaskController.List")
	}
	this.redirect(refer)
}

// 启动nginx任务
func (this *TaskController) StartNginx(id int) {
	//id, _ := this.GetInt("id")

	task, err := models.TaskGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	job, err := jobs.NewJobFromTask(task)
	if err != nil {
		this.showMsg(err.Error())
	}

	if jobs.AddJob(task.CronSpec, job) {
		task.Status = 1
		task.Update()
	}

	/*
		refer := this.Ctx.Request.Referer()
		if refer == "" {
			refer = beego.URLFor("TaskController.List")
		}
		this.redirect(refer)
	*/
}

// 暂停任务
func (this *TaskController) Pause() {
	id, _ := this.GetInt("id")

	task, err := models.TaskGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	if task.ExecuteTimes == 0 {
		jobs.RemoveJob(id)

		task.Status = 0
		task.Update()

		refer := this.Ctx.Request.Referer()
		if refer == "" {
			refer = beego.URLFor("TaskController.List")
		}
		this.redirect(refer)
	} else {
		refer := this.Ctx.Request.Referer()
		if refer == "" {
			refer = beego.URLFor("TaskController.List")
		}
		this.redirect(refer)
	}

}

// 立即执行
func (this *TaskController) Run() {
	id, _ := this.GetInt("id")

	task, err := models.TaskGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}
	beego.Debug("run task is", task)

	job, err := jobs.NewJobFromTask(task)
	if err != nil {
		this.showMsg(err.Error())
	}
	job.Run()

	this.redirect(beego.URLFor("TaskController.ViewLog", "id", job.GetLogId()))
}
