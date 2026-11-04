package models

import "time"

type MasterAsset struct {
	IDAsset              uint       `gorm:"primaryKey;column:id_asset;autoIncrement" json:"id_asset"`
	KodeAsset            string     `gorm:"column:kode_asset;not null" json:"kode_asset"`
	NamaMesin            string     `gorm:"column:nama_mesin;not null" json:"nama_mesin"`
	Merk                 string     `gorm:"column:merk" json:"merk"`
	TipeModel            string     `gorm:"column:tipe_model" json:"tipe_model"`
	SerialNumber         string     `gorm:"column:serial_number" json:"serial_number"`
	TahunPerolehan       int        `gorm:"column:tahun_perolehan" json:"tahun_perolehan"`
	KapasitasSpesifikasi string     `gorm:"column:kapasitas_spesifikasi" json:"kapasitas_spesifikasi"`
	Lokasi               string     `gorm:"column:lokasi" json:"lokasi"`
	PenanggungJawab      string     `gorm:"column:penanggung_jawab" json:"penanggung_jawab"`
	Status               string     `gorm:"column:status;default:'Operasional'" json:"status"`
	Kondisi              string     `gorm:"column:kondisi;default:'Baik'" json:"kondisi"`
	TanggalPendataan     *time.Time `gorm:"column:tanggal_pendataan" json:"tanggal_pendataan"`
	FotoLink             string     `gorm:"column:foto_link" json:"foto_link"`
}

func (MasterAsset) TableName() string {
	return "master_asset"
}
