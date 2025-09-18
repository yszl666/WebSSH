package main

import (
	"embed"
	"errors"
	"fmt"
	"gossh/app/config"
	"gossh/app/middleware"
	"gossh/app/model"
	"gossh/app/service"
	"gossh/gin"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// 插件路由缓存，用于存储动态添加的插件路由
var pluginRoutes = struct {
	routes map[string]*http.Handler
	mu     sync.RWMutex
}{routes: make(map[string]*http.Handler)}

// 插件文件系统，用于提供插件前端文件访问
// 这个结构可以根据插件配置动态加载不同的插件文件系统


// 使用go 1.16+ 新特性
//
//go:embed webroot
var dir embed.FS

// StaticFile 嵌入普通的静态资源
type StaticFile struct {
	// 静态资源
	embedFS embed.FS

	// 设置embed文件到静态资源的相对路径，也就是embed注释里的路径
	path string
}

// Open 静态资源被访问的核心逻辑
func (w StaticFile) Open(name string) (fs.File, error) {
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return nil, errors.New("http: invalid character in file path")
	}

	fullName := filepath.Join(w.path, filepath.FromSlash(path.Clean("/"+name)))
	fullName = strings.ReplaceAll(fullName, `\`, `/`)
	file, err := w.embedFS.Open(fullName)
	return file, err
}

func init() {
	config.InitConfig()
	model.InitDatabase()
	service.InitSessionClean()
	service.InitSshServer()
	fmt.Printf("WebBaseDir:[%s]\n", config.DefaultConfig.WebBaseDir)
}

func main() {

	gin.SetMode(gin.ReleaseMode)
	var engine = gin.Default()
	engine.Use(middleware.DbCheck(), middleware.NetFilter())
	engine.GET("/web_base_dir", func(c *gin.Context) { c.JSON(200, gin.H{"code": 0, "web_base_dir": config.DefaultConfig.WebBaseDir}) })

	engine.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, config.DefaultConfig.WebBaseDir+"/app")
	})

	// 不需要认证的路由
	var open = engine.Group(config.DefaultConfig.WebBaseDir)
	open.StaticFS("/app", http.FS(StaticFile{embedFS: dir, path: "webroot"}))
	open.POST("/api/login", service.UserLogin)
	open.POST("/api/sys/db_conn_check", service.DbConnCheck)
	open.GET("/api/sys/is_init", service.GetIsInit)
	open.POST("/api/sys/init", service.SysInit)

	// 需要认证的路由
	var auth = engine.Group(config.DefaultConfig.WebBaseDir,
		middleware.SysInit(),
		middleware.JWTAuth(),
		middleware.PremCheck(engine),
	)

	{ // SSH 连接配置
		auth.GET("/api/conn_conf", service.ConfFindAll)
		auth.GET("/api/conn_conf/:id", service.ConfFindByID)
		auth.POST("/api/conn_conf", service.ConfCreate)
		auth.PUT("/api/conn_conf", service.ConfUpdateById)
		auth.DELETE("/api/conn_conf/:id", service.ConfDeleteById)
	}

	{ // 命令收藏
		auth.GET("/api/cmd_note", service.CmdNoteFindAll)
		auth.GET("/api/cmd_note/:id", service.CmdNoteFindByID)
		auth.POST("/api/cmd_note", service.CmdNoteCreate)
		auth.PUT("/api/cmd_note", service.CmdNoteUpdateById)
		auth.DELETE("/api/cmd_note/:id", service.CmdNoteDeleteById)
	}

	{ // 策略配置
		auth.GET("/api/policy_conf", service.PolicyConfFindAll)
		auth.GET("/api/policy_conf/:id", service.PolicyConfFindByID)
		auth.POST("/api/policy_conf", service.PolicyConfCreate)
		auth.PUT("/api/policy_conf", service.PolicyConfUpdateById)
		auth.DELETE("/api/policy_conf/:id", service.PolicyConfDeleteById)
	}

	{ // 访问控制
		auth.GET("/api/net_filter", service.NetFilterFindAll)
		auth.GET("/api/net_filter/:id", service.NetFilterFindByID)
		auth.POST("/api/net_filter", service.NetFilterCreate)
		auth.PUT("/api/net_filter", service.NetFilterUpdateById)
		auth.DELETE("/api/net_filter/:id", service.NetFilterDeleteById)
	}

	{ // Web用户管理
		auth.GET("/api/user", service.UserFindAll)
		auth.GET("/api/user/:id", service.UserFindByID)
		auth.POST("/api/user", service.UserCreate)
		auth.PUT("/api/user", service.UserUpdateById)
		auth.DELETE("/api/user/:id", service.UserDeleteById)
		auth.PATCH("/api/user/check_name_exists", service.CheckUserNameExists)
		auth.PATCH("/api/user/pwd", service.ModifyPasswd)
	}

	{ // SSHD用户管理
		auth.GET("/api/sshd_user", service.SshdUserFindAll)
		auth.GET("/api/sshd_user/:id", service.SshdUserFindByID)
		auth.POST("/api/sshd_user", service.SshdUserCreate)
		auth.PUT("/api/sshd_user", service.SshdUserUpdateById)
		auth.DELETE("/api/sshd_user/:id", service.SshdUserDeleteById)
		auth.PATCH("/api/sshd_user/check_name_exists", service.CheckSshdUserNameExists)
	}

	{ // SSHD证书管理
		auth.GET("/api/sshd_cert", service.SshdCertFindAll)
		auth.GET("/api/sshd_cert_text", service.GetSshdCertAuthorizedKeys)
		auth.GET("/api/sshd_cert/:id", service.SshdCertFindByID)
		auth.POST("/api/sshd_cert", service.SshdCertCreate)
		auth.PUT("/api/sshd_cert", service.SshdCertUpdateById)
		auth.DELETE("/api/sshd_cert/:id", service.SshdCertDeleteById)
		auth.PATCH("/api/sshd_cert/check_name_exists", service.CheckSshdCertNameExists)
	}

	{ // 审计日志
		auth.POST("/api/login_audit", service.LoginAuditSearch)
	}

	{ // SSH链接
		auth.GET("/api/conn_manage/online_client", service.GetOnlineClient)
		auth.PUT("/api/conn_manage/refresh_conn_time", service.RefreshConnTime)
		auth.POST("/api/sftp/create_dir", service.SftpCreateDir)
		auth.POST("/api/sftp/list", service.SftpList)
		auth.GET("/api/sftp/download", service.SftpDownLoad)
		auth.PUT("/api/sftp/upload", service.SftpUpload)
		auth.DELETE("/api/sftp/delete", service.SftpDelete)
		auth.GET("/api/ssh/conn", service.NewSshConn)
		auth.PATCH("/api/ssh/conn", service.ResizeWindow)
		auth.POST("/api/ssh/exec", service.ExecCommand)
		auth.POST("/api/ssh/disconnect", service.Disconnect)
		auth.POST("/api/ssh/create_session", service.CreateSessionId)
	}

	{ // 系统配置
		auth.GET("/api/sys/config", service.GetRunConf)
		auth.POST("/api/sys/config", service.SetRunConf)
	}

	{ // 插件管理（需要管理员权限）
		auth.GET("/api/plugin", service.PluginFindAllAdmin)
		auth.GET("/api/plugin/:id", service.PluginFindByID)
		auth.POST("/api/plugin", service.PluginCreate)
		auth.PUT("/api/plugin", service.PluginUpdateById)
		auth.DELETE("/api/plugin/:id", service.PluginDeleteById)
		auth.PATCH("/api/plugin/:id/status", service.PluginToggleStatus)
	}

	{ // 插件API（不需要认证，通过plugin_id进行验证）
		auth.POST("/api/plugin/exec_ssh", service.PluginExecSSHCommand)
	}

	{ // 插件前端页面访问
		// 这里使用通配符路由，捕获所有/plugin/开头的请求
		auth.GET("/plugin/*filepath", func(c *gin.Context) {
			// 解析插件名称和文件路径
			pathParts := strings.Split(c.Param("filepath"), "/")
			if len(pathParts) < 2 {
				c.JSON(404, gin.H{"code": 404, "msg": "插件路径格式错误"})
				return
			}

			pluginName := pathParts[1]
			var filePath string
			if len(pathParts) > 2 {
				filePath = strings.Join(pathParts[2:], "/")
			}

			// 查询插件信息
			plugin, err := (&model.Plugin{}).FindByName(pluginName)
			if err != nil || plugin.Status != "enabled" {
				c.JSON(404, gin.H{"code": 404, "msg": "插件不存在或已禁用"})
				return
			}

			// 如果没有指定文件路径，默认使用入口文件
			if filePath == "" {
				filePath = plugin.EntryFile
			}

			// 构建完整的文件路径
			fullPath := path.Join(plugin.Path, filePath)

			// 检查文件是否存在
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				c.JSON(404, gin.H{"code": 404, "msg": "文件不存在"})
				return
			}

			// 提供文件服务
			c.File(fullPath)
		})
	}

	address := fmt.Sprintf("%s:%s", config.DefaultConfig.Address, config.DefaultConfig.Port)
	_, certErr := os.Open(config.DefaultConfig.CertFile)
	_, keyErr := os.Open(config.DefaultConfig.KeyFile)

	// 如果证书和私钥文件存在,就使用https协议,否则使用http协议
	if certErr == nil && keyErr == nil {
		slog.Info("https_server_start", "address", address)
		err := engine.RunTLS(address, config.DefaultConfig.CertFile, config.DefaultConfig.KeyFile)
		if err != nil {
			slog.Error("RunServeTLSError:", "msg", err.Error())
			os.Exit(1)
			return
		}
	} else {
		slog.Info("http_server_start", "address", address)
		err := engine.Run(address)
		if err != nil {
			slog.Error("RunServeError:", "msg", err.Error())
			os.Exit(1)
			return
		}
	}
}
