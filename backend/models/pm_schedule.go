package models

import "database/sql"

// PmSchedule maps to the pm_schedule table.
type PmSchedule struct {
	PmID             string         `gorm:"primaryKey;column:pm_id;type:varchar(30)"      json:"pm_id"`
	IDAsset          uint           `gorm:"column:id_asset;type:int unsigned;not null;index" json:"id_asset"`
	ItemPm           string         `gorm:"column:item_pm;type:varchar(150);not null"     json:"item_pm"`
	Aktivitas        sql.NullString `gorm:"column:aktivitas;type:text"                    json:"aktivitas"`
	Interval         sql.NullString `gorm:"column:interval;type:varchar(50)"              json:"interval"`
	TanggalTerjadwal string         `gorm:"column:tanggal_terjadwal;not null;type:date"   json:"tanggal_terjadwal"`
	TanggalRealisasi sql.NullString `gorm:"column:tanggal_realisasi;type:date"            json:"tanggal_realisasi"`
	Hasil            sql.NullString `gorm:"column:hasil;type:text"                        json:"hasil"`
	Tindakan         sql.NullString `gorm:"column:tindakan;type:text"                     json:"tindakan"`
	Petugas          sql.NullString `gorm:"column:petugas;type:varchar(100)"              json:"petugas"`
	Status           string         `gorm:"column:status;type:varchar(30);default:'Terjadwal'" json:"status"`
	BuktiLink        sql.NullString `gorm:"column:bukti_link;type:varchar(255)"          json:"bukti_link"`
}

func (PmSchedule) TableName() string { return "pm_schedule" }
