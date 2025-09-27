package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sublink/api" // 〔中文注释〕: 新增导入 api 包
	"sublink/middlewares"
	"sublink/models"
	"sublink/routers"
	"sublink/settings"
	"sublink/utils"
	"time" // 〔中文注释〕: 新增导入 time 包 
    "github.com/gin-contrib/cors" // 〔中文注释〕: 1. 新增导入 CORS 中间件

	"github.com/gin-gonic/gin"
)

//go:embed static/js/*
//go:embed static/css/*
//go:embed static/img/*
//go:embed static/*
var embeddedFiles embed.FS

//go:embed template
var Template embed.FS

// 版本号
var version string

func Templateinit() {
	// 设置template路径
	// 检查目录是否创建
	subFS, err := fs.Sub(Template, "template")
	if err != nil {
		log.Println(err)
		return // 如果出错，直接返回
	}
	entries, err := fs.ReadDir(subFS, ".")
	if err != nil {
		log.Println(err)
		return // 如果出错，直接返回
	}
	// 创建template目录
	_, err = os.Stat("./template")
	if os.IsNotExist(err) {
		err = os.Mkdir("./template", 0666)
		if err != nil {
			log.Println(err)
			return
		}
	}
	// 写入默认模板
	for _, entry := range entries {
		_, err := os.Stat("./template/" + entry.Name())
		//如果文件不存在则写入默认模板
		if os.IsNotExist(err) {
			data, err := fs.ReadFile(subFS, entry.Name())
			if err != nil {
				log.Println(err)
				continue
			}
			err = os.WriteFile("./template/"+entry.Name(), data, 0666)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func main() {
	// 初始化配置
	models.ConfigInit()
	config := models.ReadConfig() // 读取配置文件
	var port = config.Port        // 读取端口号
	// 获取版本号
	var Isversion bool
	version = "v2.2"
	flag.BoolVar(&Isversion, "version", false, "显示版本号")
	flag.Parse()
	if Isversion {
		fmt.Println(version)
		return
	}
	// 初始化数据库
	models.InitSqlite()
	// 获取命令行参数
	args := os.Args
	// 如果长度小于2则没有接收到任何参数
	if len(args) < 2 {
		Run(port)
		return
	}
	// 命令行参数选择
	settingCmd := flag.NewFlagSet("setting", flag.ExitOnError)
	var username, password string
	settingCmd.StringVar(&username, "username", "", "设置账号")
	settingCmd.StringVar(&password, "password", "", "设置密码")
	settingCmd.IntVar(&port, "port", 8000, "修改端口")
	switch args[1] {
	// 解析setting命令标志
	case "setting":
		settingCmd.Parse(args[2:])
		fmt.Println(username, password)
		settings.ResetUser(username, password)
		return
	case "run":
		settingCmd.Parse(args[2:])
		models.SetConfig(models.Config{
			Port: port,
		}) // 设置端口
		Run(port)
	default:
		return

	}
}

func Run(port int) {
	// 初始化gin框架
	r := gin.Default()
	// 〔中文注释〕: 1. CORS 跨域配置保持不变
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 〔中文注释〕: 为了方便调试，暂时用 "*"，生产环境建议替换为您的 X-Panel 域名
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// 初始化日志配置
	utils.Loginit()
	// 初始化模板
	Templateinit()

	// 〔中文注释〕: 2. 将所有【公共路由】（无需登录）直接注册在 r 上
	// ----------------------------------------------------
	// 设置静态资源和首页
	staticFiles, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Println(err)
	}
	r.StaticFS("/static", http.FS(staticFiles))
    // 设置模板路径
	r.GET("/", func(c *gin.Context) {
		data, err := fs.ReadFile(staticFiles, "index.html")
		if err != nil {
			c.Error(err)
			return
		}
		c.Data(200, "text/html", data)
	})

	// 〔中文注释〕: 3. 创建一个新的【私有路由组】，并将需要 Token 验证的路由全部放入其中
	// ----------------------------------------------------
	privateGroup := r.Group("/api/v1")
	// 〔中文注释〕: 4. 只对这个私有路由组应用 Token 验证中间件
	privateGroup.Use(middlewares.AuthorToken)
	{
		// 〔中文注释〕: 将所有需要登录才能访问的路由注册到这个 `privateGroup`
		routers.User(privateGroup)
		routers.Mentus(privateGroup)
		routers.Nodes(privateGroup)
		routers.Total(privateGroup)
		routers.Templates(privateGroup)
	}

		// X-Panel 通信的公开接口
	r.POST("/api/short", api.GenerateShortLink)
	r.POST("/api/convert", api.ConvertSubscription)

	// 其他公共接口
	routers.Subcription(r) // 订阅链接 /c/ 是公开的
	routers.Version(r, version) // 版本号接口是公开的
	routers.Clients(r)     // 客户端订阅 /c/ 必须是公开的

	// 启动服务
	r.Run(fmt.Sprintf("0.0.0.0:%d", port))
}
