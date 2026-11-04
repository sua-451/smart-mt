package models

import (
	"database/sql"
	"time"
)

// MasterAsset maps to the master_asset table.
type MasterAsset struct {
	IDAsset              uint           `gorm:"primaryKey;column:id_asset;type:int unsigned;autoIncrement" json:"id_asset"`
	KodeAsset            string         `gorm:"column:kode_asset;type:varchar(30);not null;index"          json:"kode_asset"`
	NamaMesin            string         `gorm:"column:nama_mesin;type:varchar(100);not null"               json:"nama_mesin"`
	Merk                 sql.NullString `gorm:"column:merk;type:varchar(100)"                              json:"merk"`
	TipeModel            sql.NullString `gorm:"column:tipe_model;type:varchar(100)"                        json:"tipe_model"`
	SerialNumber         sql.NullString `gorm:"column:serial_number;type:varchar(50);uniqueIndex"         json:"serial_number"`
	TahunPerolehan       sql.NullInt16  `gorm:"column:tahun_perolehan"                                     json:"tahun_perolehan"`
	KapasitasSpesifikasi sql.NullString `gorm:"column:kapasitas_spesifikasi;type:text"                     json:"kapasitas_spesifikasi"`
	Lokasi               sql.NullString `gorm:"column:lokasi;type:varchar(100)"                            json:"lokasi"`
	PenanggungJawab      sql.NullString `gorm:"column:penanggung_jawab;type:varchar(100)"                  json:"penanggung_jawab"`
	Status               string         `gorm:"column:status;type:varchar(30);default:'Operasional'"       json:"status"`
	Kondisi              string         `gorm:"column:kondisi;type:varchar(30);default:'Baik'"              json:"kondisi"`
	SumberData           sql.NullString `gorm:"column:sumber_data;type:varchar(50)"                         json:"sumber_data"`
	TanggalPendataan     sql.NullTime   `gorm:"column:tanggal_pendataan;type:date"                         json:"tanggal_pendataan"`
	Petugas              sql.NullString `gorm:"column:petugas;type:varchar(100)"                           json:"petugas"`
	FotoLink             sql.NullString `gorm:"column:foto_link;type:varchar(255)"                        json:"foto_link"`
}

func (MasterAsset) TableName() string { return "master_asset" }

// AssetPhoto maps to the asset_photo table.
type AssetPhoto struct {
	PhotoID   uint      `gorm:"primaryKey;column:photo_id;type:int unsigned;autoIncrement"   json:"photo_id"`
	IDAsset   uint      `gorm:"column:id_asset;type:int unsigned;not null;index"             json:"id_asset"`
	Filename  string    `gorm:"column:filename;type:varchar(255);not null;uniqueIndex"       json:"filename"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"                             json:"created_at"`
}

func (AssetPhoto) TableName() string { return "asset_photo" }

// AssetWithMeta is a convenience struct for list/detail views.
type AssetWithMeta struct {
	MasterAsset
	PhotoCount     int    `gorm:"column:photo_count"      json:"photo_count"`
	WorkOrderCount int    `gorm:"column:work_order_count" json:"work_order_count"`
	FirstPhoto     string `gorm:"column:first_photo"      json:"first_photo"`
}
