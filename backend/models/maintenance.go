package models

import "database/sql"

// HistoriKerusakan maps to the histori_kerusakan table.
type HistoriKerusakan struct {
	FailureID         string          `gorm:"primaryKey;column:failure_id;type:varchar(30)"    json:"failure_id"`
	IDAsset           uint            `gorm:"column:id_asset;type:int unsigned;not null;index" json:"id_asset"`
	WoNumber          sql.NullString  `gorm:"column:wo_number;type:varchar(30);index"          json:"wo_number"`
	TanggalJamMulai   string          `gorm:"column:tanggal_jam_mulai;not null"                json:"tanggal_jam_mulai"`
	TanggalJamSelesai sql.NullString  `gorm:"column:tanggal_jam_selesai"                       json:"tanggal_jam_selesai"`
	Komponen          sql.NullString  `gorm:"column:komponen;type:varchar(150)"                json:"komponen"`
	Gejala            sql.NullString  `gorm:"column:gejala;type:text"                          json:"gejala"`
	Penyebab          sql.NullString  `gorm:"column:penyebab;type:text"                        json:"penyebab"`
	TindakanPerbaikan sql.NullString  `gorm:"column:tindakan_perbaikan;type:text"              json:"tindakan_perbaikan"`
	Teknisi           sql.NullString  `gorm:"column:teknisi;type:varchar(100)"                 json:"teknisi"`
	RepairTimeJam     sql.NullFloat64 `gorm:"column:repair_time_jam;type:decimal(10,2)"        json:"repair_time_jam"`
	DowntimeJam       sql.NullFloat64 `gorm:"column:downtime_jam;type:decimal(10,2)"           json:"downtime_jam"`
	Biaya             float64         `gorm:"column:biaya;default:0;type:decimal(15,2)"         json:"biaya"`
	SumberData        sql.NullString  `gorm:"column:sumber_data;type:varchar(50)"              json:"sumber_data"`
	BuktiLink         sql.NullString  `gorm:"column:bukti_link;type:varchar(255)"              json:"bukti_link"`
}

func (HistoriKerusakan) TableName() string { return "histori_kerusakan" }

// Downtime maps to the downtime table.
type Downtime struct {
	DowntimeID      string          `gorm:"primaryKey;column:downtime_id;type:varchar(30)"   json:"downtime_id"`
	IDAsset         uint            `gorm:"column:id_asset;type:int unsigned;not null;index" json:"id_asset"`
	Tanggal         string          `gorm:"column:tanggal;not null;type:date"                json:"tanggal"`
	MulaiDowntime   sql.NullString  `gorm:"column:mulai_downtime"                            json:"mulai_downtime"`
	SelesaiDowntime sql.NullString  `gorm:"column:selesai_downtime"                          json:"selesai_downtime"`
	DurasiJam       sql.NullFloat64 `gorm:"column:durasi_jam;type:decimal(10,2)"            json:"durasi_jam"`
	Penyebab        sql.NullString  `gorm:"column:penyebab;type:text"                        json:"penyebab"`
	Jenis           sql.NullString  `gorm:"column:jenis;type:varchar(20)"                    json:"jenis"`
	WoNumber        sql.NullString  `gorm:"column:wo_number;type:varchar(30);index"          json:"wo_number"`
}

func (Downtime) TableName() string { return "downtime" }

// KpiBeforeAfter maps to the kpi_before_after table.
type KpiBeforeAfter struct {
	IDKpi              uint     `gorm:"primaryKey;column:id_kpi;type:int unsigned;autoIncrement" json:"id_kpi"`
	IDAsset            uint     `gorm:"column:id_asset;type:int unsigned;not null;index"         json:"id_asset"`
	Periode            string   `gorm:"column:periode;type:varchar(30);not null"               json:"periode"`
	OperatingTimeJam   float64  `gorm:"column:operating_time_jam;type:decimal(10,2)"           json:"operating_time_jam"`
	Failure            int      `gorm:"column:failure"                                         json:"failure"`
	RepairTimeJam      float64  `gorm:"column:repair_time_jam;type:decimal(10,2)"              json:"repair_time_jam"`
	DowntimeJam        float64  `gorm:"column:downtime_jam;type:decimal(10,2)"                 json:"downtime_jam"`
	PmTerjadwal        int      `gorm:"column:pm_terjadwal"                                    json:"pm_terjadwal"`
	PmTerlaksana       int      `gorm:"column:pm_terlaksana"                                   json:"pm_terlaksana"`
	MtbfJam            *float64 `gorm:"column:mtbf_jam;->;<-:false"                           json:"mtbf_jam"`
	MttrJam            *float64 `gorm:"column:mttr_jam;->;<-:false"                           json:"mttr_jam"`
	AvailabilityPersen *float64 `gorm:"column:availability_persen;->;<-:false"                json:"availability_persen"`
	PmCompliancePersen *float64 `gorm:"column:pm_compliance_persen;->;<-:false"                json:"pm_compliance_persen"`
}

func (KpiBeforeAfter) TableName() string { return "kpi_before_after" }

// KondisiAwal maps to the kondisi_awal table.
type KondisiAwal struct {
	IDKondisi      uint           `gorm:"primaryKey;column:id_kondisi;type:int unsigned;autoIncrement" json:"id_kondisi"`
	Tanggal        string         `gorm:"column:tanggal;not null;type:date"          json:"tanggal"`
	IDAsset        uint           `gorm:"column:id_asset;type:int unsigned;not null;index" json:"id_asset"`
	KomponenItem   string         `gorm:"column:komponen_item;type:varchar(150);not null" json:"komponen_item"`
	StandarKondisi sql.NullString `gorm:"column:standar_kondisi;type:varchar(255)"  json:"standar_kondisi"`
	KondisiAktual  sql.NullString `gorm:"column:kondisi_aktual;type:varchar(255)"   json:"kondisi_aktual"`
	OkNg           sql.NullString `gorm:"column:ok_ng;type:varchar(10)"              json:"ok_ng"`
	Temuan         sql.NullString `gorm:"column:temuan;type:text"                    json:"temuan"`
	Tindakan       sql.NullString `gorm:"column:tindakan;type:text"                  json:"tindakan"`
	Petugas        sql.NullString `gorm:"column:petugas;type:varchar(100)"           json:"petugas"`
	BuktiLink      sql.NullString `gorm:"column:bukti_link;type:varchar(255)"        json:"bukti_link"`
}

func (KondisiAwal) TableName() string { return "kondisi_awal" }
