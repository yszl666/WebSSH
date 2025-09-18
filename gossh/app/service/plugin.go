package service

import (
	_ "errors"
	"fmt"
	"gossh/app/model"
	"gossh/crypto/ssh"
	"gossh/gin"
	"os"
	"path"
	_ "path/filepath"
	_ "strings"
	"time"
)

// PluginCreate 创建插件
func PluginCreate(c *gin.Context) {
	var plugin model.Plugin
	if err := c.ShouldBind(&plugin); err != nil {
		c.JSON(200, gin.H{"code": 2, "msg": err.Error()})
		return
	}

	// 检查插件目录是否存在
	if !isDirExists(plugin.Path) {
		c.JSON(200, gin.H{"code": 3, "msg": "插件目录不存在"})
		return
	}

	// 检查入口文件是否存在
	entryFilePath := path.Join(plugin.Path, plugin.EntryFile)
	if !isFileExists(entryFilePath) {
		c.JSON(200, gin.H{"code": 4, "msg": "入口文件不存在"})
		return
	}

	// 检查插件名称是否已存在
	_, err := (&model.Plugin{}).FindByName(plugin.Name)
	if err == nil {
		c.JSON(200, gin.H{"code": 5, "msg": "插件名称已存在"})
		return
	}

	// 创建插件
	if err := (&model.Plugin{}).Create(&plugin); err != nil {
		c.JSON(200, gin.H{"code": 6, "msg": "创建失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "ok", "data": plugin})
}

// PluginFindAll 查询所有启用的插件
func PluginFindAll(c *gin.Context) {
	list, err := (&model.Plugin{}).FindAll()
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "查询失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "ok", "data": list})
}

// PluginFindAllAdmin 管理员查询所有插件（包括禁用的）
func PluginFindAllAdmin(c *gin.Context) {
	offset := 0
	limit := 100

	if offsetStr := c.Query("offset"); offsetStr != "" {
		fmt.Sscanf(offsetStr, "%d", &offset)
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	list, err := (&model.Plugin{}).FindAllAdmin(offset, limit)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "查询失败: " + err.Error()})
		return
	}

	// 获取总数
	var total int64
	model.Db.Model(&model.Plugin{}).Count(&total)

	c.JSON(200, gin.H{"code": 0, "msg": "ok", "data": list, "total": total})
}

// PluginFindByID 根据ID查询插件
func PluginFindByID(c *gin.Context) {
	var id uint
	fmt.Sscanf(c.Param("id"), "%d", &id)

	plugin, err := (&model.Plugin{}).FindByID(id)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "插件不存在"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "ok", "data": plugin})
}

// PluginUpdateById 更新插件
func PluginUpdateById(c *gin.Context) {
	var plugin model.Plugin
	if err := c.ShouldBind(&plugin); err != nil {
		c.JSON(200, gin.H{"code": 2, "msg": err.Error()})
		return
	}

	// 检查插件是否存在
	exists, err := (&model.Plugin{}).FindByID(plugin.ID)
	if err != nil {
		c.JSON(200, gin.H{"code": 3, "msg": "插件不存在"})
		return
	}

	// 如果修改了目录或入口文件，检查是否存在
	if plugin.Path != exists.Path {
		if !isDirExists(plugin.Path) {
			c.JSON(200, gin.H{"code": 4, "msg": "插件目录不存在"})
			return
		}
	}

	if plugin.EntryFile != exists.EntryFile || plugin.Path != exists.Path {
		entryFilePath := path.Join(plugin.Path, plugin.EntryFile)
		if !isFileExists(entryFilePath) {
			c.JSON(200, gin.H{"code": 5, "msg": "入口文件不存在"})
			return
		}
	}

	// 检查插件名称是否已存在（除了当前插件）
	if plugin.Name != exists.Name {
		_, err := (&model.Plugin{}).FindByName(plugin.Name)
		if err == nil {
			c.JSON(200, gin.H{"code": 6, "msg": "插件名称已存在"})
			return
		}
	}

	// 更新插件
	if err := (&model.Plugin{}).UpdateById(plugin.ID, &plugin); err != nil {
		c.JSON(200, gin.H{"code": 7, "msg": "更新失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "ok", "data": plugin})
}

// PluginDeleteById 删除插件
func PluginDeleteById(c *gin.Context) {
	var id uint
	fmt.Sscanf(c.Param("id"), "%d", &id)

	// 删除插件
	if err := (&model.Plugin{}).DeleteByID(id); err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "删除失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "ok"})
}

// PluginToggleStatus 切换插件状态（启用/禁用）
func PluginToggleStatus(c *gin.Context) {
	var id uint
	fmt.Sscanf(c.Param("id"), "%d", &id)

	// 查找插件
	plugin, err := (&model.Plugin{}).FindByID(id)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "插件不存在"})
		return
	}

	// 切换状态
	if plugin.Status == "enabled" {
		plugin.Status = "disabled"
	} else {
		plugin.Status = "enabled"
	}

	// 更新插件
	if err := (&model.Plugin{}).UpdateById(plugin.ID, &plugin); err != nil {
		c.JSON(200, gin.H{"code": 2, "msg": "更新失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "ok", "data": plugin})
}

// 检查目录是否存在
func isDirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// 检查文件是否存在
func isFileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// PluginExecSSHCommand 插件执行SSH命令的API
// 插件可以通过这个API执行SSH命令并获取结果
func PluginExecSSHCommand(c *gin.Context) {
	type Param struct {
		PluginId  string `form:"plugin_id" binding:"required,min=1" json:"plugin_id"`
		SessionId string `form:"session_id" binding:"required,min=10" json:"session_id"`
		Cmd       string `form:"cmd" binding:"required,min=1" json:"cmd"`
	}
	var param Param
	if err := c.ShouldBind(&param); err != nil {
		c.JSON(200, gin.H{"code": 2, "msg": err.Error()})
		return
	}

	// 检查插件是否存在且已启用
	//plugin, err := (&model.Plugin{}).FindByID(param.PluginId)
	//if err != nil || plugin.Status != "enabled" {
	//	c.JSON(200, gin.H{"code": 3, "msg": "插件不存在或已禁用"})
	//	return
	//}

	// 调用现有的ExecCommand函数执行命令
	// 这里我们需要稍微调整ExecCommand的实现，使其可以被复用
	// 但由于现在无法修改ExecCommand，我们可以复制其逻辑

	cli, ok := OnlineClients.Load(param.SessionId)
	if !ok || cli == nil {
		c.JSON(200, gin.H{"code": 4, "msg": "会话不存在"})
		return
	}

	conn, ok := cli.(*SshConn)
	if !ok || conn == nil {
		c.JSON(200, gin.H{"code": 5, "msg": "连接不存在"})
		return
	}

	// 更新最后活跃时间
	conn.LastActiveTime = time.Now()

	//创建ssh-session
	session, err := conn.sshClient.NewSession()
	if err != nil {
		c.JSON(200, gin.H{"code": 6, "msg": "创建会话失败: " + err.Error()})
		return
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	//执行命令
	out, err := session.CombinedOutput(param.Cmd)
	if err != nil {
		c.JSON(200, gin.H{"code": 7, "msg": "执行命令失败", "data": string(out)})
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": string(out),
	})
}
