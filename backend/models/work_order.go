package models

import "database/sql"

// WorkOrder maps to the work_order table.
type WorkOrder struct {
	WoNumber         string         `gorm:"primaryKey;column:wo_number;type:varchar(30)" json:"wo_number"`
	IDAsset          uint           `gorm:"column:id_asset;type:int unsigned;not null;index" json:"id_asset"`
	TanggalJamLapor  string         `gorm:"column:tanggal_jam_lapor;not null"            json:"tanggal_jam_lapor"`
	Pelapor          sql.NullString `gorm:"column:pelapor;type:varchar(100)"             json:"pelapor"`
	Jenis            sql.NullString `gorm:"column:jenis;type:varchar(50)"                json:"jenis"`
	Prioritas        sql.NullString `gorm:"column:prioritas;type:varchar(50)"            json:"prioritas"`
	KeluhanGejala    sql.NullString `gorm:"column:keluhan_gejala;type:text"              json:"keluhan_gejala"`
	Teknisi          sql.NullString `gorm:"column:teknisi;type:varchar(100)"             json:"teknisi"`
	MulaiPekerjaan   sql.NullString `gorm:"column:mulai_pekerjaan"                       json:"mulai_pekerjaan"`
	SelesaiPekerjaan sql.NullString `gorm:"column:selesai_pekerjaan"                     json:"selesai_pekerjaan"`
	Tindakan         sql.NullString `gorm:"column:tindakan;type:text"                    json:"tindakan"`
	SparePart        sql.NullString `gorm:"column:spare_part;type:varchar(255)"          json:"spare_part"`
	Biaya            float64        `gorm:"column:biaya;default:0;type:decimal(15,2)"    json:"biaya"`
	Status           string         `gorm:"column:status;type:varchar(30);default:'Open'" json:"status"`
	Verifikasi       sql.NullString `gorm:"column:verifikasi;type:varchar(100)"          json:"verifikasi"`
	BuktiLink        sql.NullString `gorm:"column:bukti_link;type:varchar(255)"          json:"bukti_link"`
}

func (WorkOrder) TableName() string { return "work_order" }

// AppWorkOrder holds the extended scheduling/workflow state for a work order.
type AppWorkOrder struct {
	WoNumber        string         `gorm:"primaryKey;column:wo_number;type:varchar(30)" json:"wo_number"`
	ScheduledDate   string         `gorm:"column:scheduled_date;type:varchar(20);not null" json:"scheduled_date"`
	TechnicianID    sql.NullInt64  `gorm:"column:technician_id;index"                    json:"technician_id"`
	WorkflowStatus  string         `gorm:"column:workflow_status;type:varchar(50);not null" json:"workflow_status"`
	MaintenanceType string         `gorm:"column:maintenance_type;type:varchar(50);not null" json:"maintenance_type"`
	PmID            sql.NullString `gorm:"column:pm_id;type:varchar(30)"                 json:"pm_id"`
}

func (AppWorkOrder) TableName() string { return "app_work_order" }

// AppReport maps to the app_report table (technician submission).
type AppReport struct {
	ReportID        uint           `gorm:"primaryKey;column:report_id;autoIncrement"              json:"report_id"`
	WoNumber        string         `gorm:"column:wo_number;type:varchar(30);not null;uniqueIndex" json:"wo_number"`
	WorkDate        string         `gorm:"column:work_date;type:varchar(20);not null"             json:"work_date"`
	BeforeCondition string         `gorm:"column:before_condition;not null;type:text"             json:"before_condition"`
	BeforePhoto     sql.NullString `gorm:"column:before_photo;type:varchar(255)"                  json:"before_photo"`
	AfterCondition  string         `gorm:"column:after_condition;not null;type:text"              json:"after_condition"`
	AfterPhoto      sql.NullString `gorm:"column:after_photo;type:varchar(255)"                   json:"after_photo"`
	ManagerNote     string         `gorm:"column:manager_note;type:text"                         json:"manager_note"`
	SubmittedAt     sql.NullString `gorm:"column:submitted_at"                                    json:"submitted_at"`
}

func (AppReport) TableName() string { return "app_report" }

// AppReportDetail maps to the app_report_detail table (per-damage-point records).
type AppReportDetail struct {
	DetailID          uint           `gorm:"primaryKey;column:detail_id;autoIncrement"        json:"detail_id"`
	WoNumber          string         `gorm:"column:wo_number;type:varchar(30);not null;index" json:"wo_number"`
	FailureID         sql.NullString `gorm:"column:failure_id;type:varchar(30)"               json:"failure_id"`
	DamageDescription string         `gorm:"column:damage_description;not null;type:text"     json:"damage_description"`
	DamagePhoto       sql.NullString `gorm:"column:damage_photo;type:varchar(255)"            json:"damage_photo"`
	RepairAction      string         `gorm:"column:repair_action;not null;type:text"          json:"repair_action"`
	RepairPhoto       sql.NullString `gorm:"column:repair_photo;type:varchar(255)"            json:"repair_photo"`
}

func (AppReportDetail) TableName() string { return "app_report_detail" }

// WorkOrderRow is a flattened read-model used by list and dashboard endpoints.
type WorkOrderRow struct {
	WoNumber        string `json:"wo_number"`
	TicketID        string `json:"ticket_id"`
	IDAsset         uint   `json:"id_asset"`
	AssetID         uint   `json:"asset_id"`
	MachineName     string `json:"machine_name"`
	KodeAsset       string `json:"kode_asset"`
	Location        string `json:"location"`
	Category        string `json:"category"`
	Merk            string `json:"merk"`
	TipeModel       string `json:"tipe_model"`
	Active          bool   `json:"active"`
	AssetKondisi    string `json:"asset_kondisi"`
	SerialNumber    string `json:"serial_number"`

	TanggalJamLapor string  `json:"tanggal_jam_lapor"`
	Jenis           string  `json:"jenis"`
	Prioritas       string  `json:"prioritas"`
	KeluhanGejala   string  `json:"keluhan_gejala"`
	Biaya           float64 `json:"biaya"`
	WoStatus        string  `json:"wo_status"`
	Verifikasi      string  `json:"verifikasi"`

	ScheduledDate   string `json:"scheduled_date"`
	TechnicianID    *int64 `json:"technician_id"`
	TechnicianName  string `json:"technician_name"`
	WorkflowStatus  string `json:"workflow_status"`
	Status          string `json:"status"`
	MaintenanceType string `json:"maintenance_type"`
	Severity        string `json:"severity"`
	PmID            string `json:"pm_id"`

	ReportID        *uint  `json:"report_id"`
	WorkDate        string `json:"work_date"`
	BeforeCondition string `json:"before_condition"`
	BeforePhoto     string `json:"before_photo"`
	AfterCondition  string `json:"after_condition"`
	AfterPhoto      string `json:"after_photo"`
	ManagerNote     string `json:"manager_note"`
}
