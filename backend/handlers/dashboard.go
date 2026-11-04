package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"smart-mt-backend/config"
	"smart-mt-backend/middleware"

	"github.com/google/uuid"
)

// HandleDashboard serves GET /api/dashboard – returns metrics and recent orders.
func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	SetJSONHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	user, ok := middleware.GetCurrentUser(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "Tidak terautentikasi")
		return
	}

	type Metrics struct {
		Assets          int `json:"assets"`
		ActiveOrders    int `json:"active_orders"`
		PendingApproval int `json:"pending_approval"`
		Technicians     int `json:"technicians"`
		Completed       int `json:"completed"`
	}
	var m Metrics

	if user.Role == "Manager" {
		config.DB.Raw("SELECT COUNT(*) FROM master_asset WHERE status != 'Afkir'").Scan(&m.Assets)
		config.DB.Raw("SELECT COUNT(*) FROM work_order WHERE status IN ('Open','Proses')").Scan(&m.ActiveOrders)
		config.DB.Raw("SELECT COUNT(*) FROM work_order WHERE verifikasi = 'Menunggu Approval'").Scan(&m.PendingApproval)
		config.DB.Raw("SELECT COUNT(*) FROM app_users WHERE role = 'Teknisi'").Scan(&m.Technicians)

		ensureMonthlyWorkOrders()

		dbRows, err := config.DB.Raw(`
			SELECT w.wo_number FROM work_order w
			LEFT JOIN app_work_order app ON app.wo_number = w.wo_number
			ORDER BY
			  CASE WHEN w.verifikasi = 'Menunggu Approval' THEN 0 ELSE 1 END,
			  COALESCE(app.scheduled_date, DATE(w.tanggal_jam_lapor)) ASC
			LIMIT 8
		`).Rows()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Gagal mengambil data dashboard")
			return
		}
		defer dbRows.Close()

		orders := make([]map[string]interface{}, 0)
		for dbRows.Next() {
			var woNum string
			dbRows.Scan(&woNum)
			row, err := fetchWorkOrderRow(woNum)
			if err == nil {
				orders = append(orders, row)
			}
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{"metrics": m, "orders": orders})

	} else {
		dbRows, err := config.DB.Raw(`
			SELECT wo_number FROM app_work_order
			WHERE technician_id = ? AND workflow_status IN ('Ditugaskan','Dikerjakan','Perlu Revisi')
			ORDER BY scheduled_date ASC
		`, user.UserID).Rows()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Gagal mengambil data dashboard")
			return
		}
		defer dbRows.Close()

		orders := make([]map[string]interface{}, 0)
		for dbRows.Next() {
			var woNum string
			dbRows.Scan(&woNum)
			row, err := fetchWorkOrderRow(woNum)
			if err == nil {
				orders = append(orders, row)
			}
		}

		m.ActiveOrders = len(orders)
		config.DB.Raw(
			"SELECT COUNT(*) FROM app_work_order WHERE technician_id = ? AND workflow_status = 'Menunggu Approval'",
			user.UserID,
		).Scan(&m.PendingApproval)
		config.DB.Raw(
			"SELECT COUNT(*) FROM app_work_order WHERE technician_id = ? AND workflow_status = 'Selesai'",
			user.UserID,
		).Scan(&m.Completed)

		WriteJSON(w, http.StatusOK, map[string]interface{}{"metrics": m, "orders": orders})
	}
}

// ensureMonthlyWorkOrders creates monthly Preventif WOs for all Operasional assets
// that don't already have one this month.
func ensureMonthlyWorkOrders() {
	type assetRow struct {
		IDAsset         uint
		PenanggungJawab string
	}
	var assets []assetRow
	config.DB.Raw("SELECT id_asset, penanggung_jawab FROM master_asset WHERE status = 'Operasional'").Scan(&assets)

	now := time.Now()
	monthStart := fmt.Sprintf("%d-%02d-01", now.Year(), now.Month())

	for _, a := range assets {
		var count int64
		config.DB.Raw(`
			SELECT COUNT(*) FROM work_order
			WHERE id_asset = ? AND jenis = 'Preventif'
			  AND DATE(tanggal_jam_lapor) >= DATE_FORMAT(CURDATE(),'%Y-%m-01')
		`, a.IDAsset).Scan(&count)
		if count > 0 {
			continue
		}

		suffix   := strings.ToUpper(uuid.New().String()[:8])
		woNumber := fmt.Sprintf("PM-%s-%s", now.Format("0601"), suffix)
		pmID     := "PM-" + suffix

		config.DB.Exec(`
			INSERT IGNORE INTO work_order
			  (wo_number, id_asset, tanggal_jam_lapor, pelapor, jenis, prioritas, keluhan_gejala, status)
			VALUES (?, ?, ?, 'Sistem', 'Preventif', 'Rendah', 'Pemeriksaan rutin bulanan', 'Open')
		`, woNumber, a.IDAsset, monthStart+" 09:00:00")

		config.DB.Exec(`
			INSERT IGNORE INTO pm_schedule
			  (pm_id, id_asset, item_pm, aktivitas, `+"`interval`"+`, tanggal_terjadwal, petugas, status)
			VALUES (?, ?, 'Pemeriksaan rutin bulanan', 'Pemeriksaan dan perawatan mesin', '1 bulan', ?, ?, 'Terjadwal')
		`, pmID, a.IDAsset, monthStart, a.PenanggungJawab)

		config.DB.Exec(`
			INSERT IGNORE INTO app_work_order
			  (wo_number, scheduled_date, workflow_status, maintenance_type, pm_id)
			VALUES (?, ?, 'Menunggu Penugasan', 'Rutin', ?)
		`, woNumber, monthStart, pmID)
	}
}
