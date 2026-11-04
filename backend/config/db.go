package config

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := "root:@tcp(127.0.0.1:3306)/maintenance_pnc?charset=utf8mb4&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("❌ Gagal koneksi ke database MySQL:", err)
		return
	}

	DB = database
	fmt.Println("✅ Berhasil terhubung ke database MySQL!")
}
