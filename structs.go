package main

import "gorm.io/gorm"

func migrateDB(db *gorm.DB) {

	_ = db.AutoMigrate(&Guild{})

}

type Guild struct {
	GuildID   string `gorm:"primaryKey"`
	ChannelID string
	RoleID    string
}
