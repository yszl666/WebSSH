package model

import (
	_ "gossh/gorm"
)

// Plugin 插件模型
// 用于存储插件的基本信息

type Plugin struct {
	ID          uint     `gorm:"column:id;primaryKey,autoIncrement" form:"id" json:"id"`
	Name        string   `gorm:"column:name;not null;size:64;unique" form:"name" binding:"required,min=1,max=63" json:"name"`
	Title       string   `gorm:"column:title;not null;size:128" form:"title" binding:"required,min=1,max=128" json:"title"`
	Description string   `gorm:"column:description;not null;size:512" form:"description" binding:"required,min=1,max=512" json:"description"`
	Path        string   `gorm:"column:path;not null;size:255" form:"path" binding:"required,min=1,max=255" json:"path"`
	EntryFile   string   `gorm:"column:entry_file;not null;size:255;default:'index.html'" form:"entry_file" binding:"min=1,max=255" json:"entry_file"`
	Status      string   `gorm:"column:status;not null;size:32;default:'enabled'" form:"status" binding:"required,min=1,max=32,oneof=enabled disabled" json:"status"`
	OrderNum    int      `gorm:"column:order_num;not null;default:0" form:"order_num" binding:"required,gte=0" json:"order_num"`
	CreatedAt   DateTime `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   DateTime `gorm:"column:updated_at" json:"updated_at"`
}

func (p *Plugin) Create(plugin *Plugin) error {
	return Db.Create(plugin).Error
}

func (p *Plugin) FindByID(id uint) (Plugin, error) {
	var plugin Plugin
	err := Db.First(&plugin, "id = ?", id).Error
	return plugin, err
}

func (p *Plugin) FindAll() ([]Plugin, error) {
	var list []Plugin
	err := Db.Where("status = ?", "enabled").Order("order_num asc, created_at desc").Find(&list).Error
	return list, err
}

func (p *Plugin) FindAllAdmin(offset, limit int) ([]Plugin, error) {
	var list []Plugin
	err := Db.Offset(offset).Limit(limit).Order("order_num asc, created_at desc").Find(&list).Error
	return list, err
}

func (p *Plugin) UpdateById(id uint, plugin *Plugin) error {
	return Db.Model(&plugin).Where("id = ?", id).Select("*").Omit("id").Updates(plugin).Error
}

func (p *Plugin) DeleteByID(id uint) error {
	return Db.Unscoped().Delete(&p, "id = ?", id).Error
}

func (p *Plugin) FindByName(name string) (Plugin, error) {
	var plugin Plugin
	err := Db.First(&plugin, "name = ?", name).Error
	return plugin, err
}
