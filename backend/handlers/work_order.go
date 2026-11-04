package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"smart-mt-backend/config"
	"smart-mt-backend/middleware"
	"smart-mt-backend/models"

	"github.com/google/uuid"
)

// HandleWorkOrders routes /api/work-orders and /api/work-orders/{id}/{action}.
func HandleWorkOrders(w http.ResponseWriter, r *http.Request) {
	SetJSONHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// parts: ["api","work-orders"] or ["api","work-orders","{id}"] or ["api","work-orders","{id}","action"]
	var woNum string
	var action string
	if len(parts) >= 3 {
		woNum = parts[2]
	}
	if len(parts) >= 4 {
		action = parts[3]
	}

	switch {
	case woNum != "" && action == "assign":
		handleAssign(w, r, woNum)
	case woNum != "" && action == "review":
		handleReview(w, r, woNum)
	case woNum != "" && action == "report":
		handleReport(w, r, woNum)
	case woNum != "" && r.Method == http.MethodGet:
		getWorkOrderByID(w, r, woNum)
	case woNum == "" && r.Method == http.MethodGet:
		listWorkOrders(w, r)
	case woNum == "" && r.Method == http.MethodPost:
		createWorkOrder(w, r)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
	}
}

// priorityToSeverity maps WO priority to display severity.
var priorityToSeverity = map[string]string{
	"Rendah": "Rendah", "Sedang": "Sedang", "Tinggi": "Tinggi", "Darurat": "Kritis",
}

func listWorkOrders(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetCurrentUser(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "Tidak terautentikasi")
		return
	}

	var query string
	var args []interface{}
	baseQuery := `
		SELECT w.wo_number, w.id_asset, w.tanggal_jam_lapor, w.pelapor,
		       w.jenis, w.prioritas, w.keluhan_gejala, w.biaya, w.status AS wo_status,
		       w.verifikasi, w.teknisi,
		       a.kode_asset, a.nama_mesin, a.merk, a.tipe_model, a.serial_number,
		       a.lokasi, a.kondisi, a.status AS asset_status,
		       app.scheduled_date, app.technician_id, app.workflow_status,
		       app.maintenance_type, app.pm_id,
		       u.full_name AS technician_name,
		       r.report_id, r.work_date, r.before_condition, r.after_condition,
		       r.before_photo, r.after_photo, r.manager_note
		FROM work_order w
		JOIN master_asset a ON a.id_asset = w.id_asset
		%s
		LEFT JOIN app_users u ON u.user_id = app.technician_id
		LEFT JOIN app_report r ON r.wo_number = w.wo_number
		%s
		ORDER BY %s`

	if user.Role == "Manager" {
		query = fmt.Sprintf(baseQuery,
			"LEFT JOIN app_work_order app ON app.wo_number = w.wo_number",
			"",
			"CASE WHEN w.verifikasi = 'Menunggu Approval' THEN 0 ELSE 1 END, COALESCE(app.scheduled_date, DATE(w.tanggal_jam_lapor)) ASC")
	} else {
		query = fmt.Sprintf(baseQuery,
			"JOIN app_work_order app ON app.wo_number = w.wo_number",
			"WHERE app.technician_id = ?",
			"app.scheduled_date ASC")
		args = append(args, user.UserID)
	}

	rows := make([]map[string]interface{}, 0)
	dbRows, err := config.DB.Raw(query, args...).Rows()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Gagal mengambil data work order")
		return
	}
	defer dbRows.Close()

	cols, _ := dbRows.Columns()
	for dbRows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := dbRows.Scan(ptrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range cols {
			v := vals[i]
			if b, ok := v.([]byte); ok {
				v = string(b)
			}
			row[col] = v
		}
		rows = append(rows, enrichWorkOrderRow(row))
	}

	// Also fetch report_details for any orders that have a report_id
	var reportedWOs []string
	for _, r := range rows {
		if r["report_id"] != nil {
			if woNum, ok := r["wo_number"].(string); ok && woNum != "" {
				reportedWOs = append(reportedWOs, woNum)
			}
		}
	}
	if len(reportedWOs) > 0 {
		var allDetails []models.AppReportDetail
		config.DB.Where("wo_number IN (?)", reportedWOs).Order("detail_id").Find(&allDetails)
		detailsByWO := make(map[string][]map[string]interface{})
		for _, d := range allDetails {
			detailsByWO[d.WoNumber] = append(detailsByWO[d.WoNumber], map[string]interface{}{
				"detail_id":          d.DetailID,
				"damage_description": d.DamageDescription,
				"damage_photo":       d.DamagePhoto.String,
				"repair_action":      d.RepairAction,
				"repair_photo":       d.RepairPhoto.String,
			})
		}
		for _, r := range rows {
			if woNum, ok := r["wo_number"].(string); ok {
				if dList, exists := detailsByWO[woNum]; exists {
					r["report_details"] = dList
				}
			}
		}
	}

	WriteJSON(w, http.StatusOK, rows)
}

func getWorkOrderByID(w http.ResponseWriter, r *http.Request, woNum string) {
	row, err := fetchWorkOrderRow(woNum)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Work order tidak ditemukan")
		return
	}

	// In legacy, when technician opens a Ditugaskan ticket, it moves to Dikerjakan
	if user, ok := middleware.GetCurrentUser(r); ok && user.Role == "Teknisi" {
		if status, _ := row["status"].(string); status == "Ditugaskan" {
			config.DB.Exec("UPDATE app_work_order SET workflow_status = 'Dikerjakan' WHERE wo_number = ?", woNum)
			config.DB.Exec("UPDATE work_order SET status = 'Proses' WHERE wo_number = ?", woNum)
			row["status"] = "Dikerjakan"
			row["workflow_status"] = "Dikerjakan"
		}
	}

	var details []models.AppReportDetail
	config.DB.Where("wo_number = ?", woNum).Order("detail_id").Find(&details)
	row["report_details"] = details
	WriteJSON(w, http.StatusOK, row)
}

func createWorkOrder(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetCurrentUser(r)
	if !ok || user.Role != "Manager" {
		WriteError(w, http.StatusForbidden, "Hanya Manager yang dapat membuat work order")
		return
	}

	var req struct {
		AssetID       uint   `json:"asset_id"`
		ScheduledDate string `json:"scheduled_date"`
		Jenis         string `json:"jenis"`
		Prioritas     string `json:"prioritas"`
		KeluhanGejala string `json:"keluhan_gejala"`
		ItemPm        string `json:"item_pm"`
		Aktivitas     string `json:"aktivitas"`
		Interval      string `json:"interval"`
		TechnicianID  *uint  `json:"technician_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Format permintaan tidak valid")
		return
	}
	if req.AssetID == 0 || req.ScheduledDate == "" || req.Jenis == "" {
		WriteError(w, http.StatusBadRequest, "Aset, tanggal jadwal, dan jenis wajib diisi.")
		return
	}

	now := time.Now()
	suffix := strings.ToUpper(uuid.NewString()[:8])
	woNumber := fmt.Sprintf("WO-%s-%s", now.Format("0601"), suffix)

	wo := models.WorkOrder{
		WoNumber:        woNumber,
		IDAsset:         req.AssetID,
		TanggalJamLapor: now.Format("2006-01-02 15:04:05"),
		Pelapor:         sql.NullString{String: user.FullName, Valid: true},
		Jenis:           sql.NullString{String: req.Jenis, Valid: req.Jenis != ""},
		Prioritas:       sql.NullString{String: req.Prioritas, Valid: req.Prioritas != ""},
		KeluhanGejala:   sql.NullString{String: req.KeluhanGejala, Valid: req.KeluhanGejala != ""},
		Status:          "Open",
	}
	if err := config.DB.Create(&wo).Error; err != nil {
		WriteError(w, http.StatusInternalServerError, "Gagal membuat work order: "+err.Error())
		return
	}

	workflowStatus := "Menunggu Penugasan"
	var techID sql.NullInt64
	if req.TechnicianID != nil && *req.TechnicianID != 0 {
		if technicianHasScheduleConflict(*req.TechnicianID, req.ScheduledDate, "") {
			WriteError(w, http.StatusBadRequest, "Teknisi sudah memiliki work order aktif pada tanggal tersebut.")
			return
		}
		techID = sql.NullInt64{Int64: int64(*req.TechnicianID), Valid: true}
		workflowStatus = "Ditugaskan"
		// denormalise technician name
		var tech models.AppUser
		if config.DB.First(&tech, *req.TechnicianID).Error == nil {
			config.DB.Exec("UPDATE work_order SET teknisi = ? WHERE wo_number = ?", tech.FullName, woNumber)
		}
	}

	maintenanceType := req.Jenis
	if req.Jenis == "Preventif" && req.ItemPm != "" {
		maintenanceType = req.ItemPm
	}

	appWO := models.AppWorkOrder{
		WoNumber:        woNumber,
		ScheduledDate:   req.ScheduledDate,
		TechnicianID:    techID,
		WorkflowStatus:  workflowStatus,
		MaintenanceType: maintenanceType,
	}

	if req.Jenis == "Preventif" && req.ItemPm != "" {
		pmID := fmt.Sprintf("PM-%s", suffix)
		pm := models.PmSchedule{
			PmID:             pmID,
			IDAsset:          req.AssetID,
			ItemPm:           req.ItemPm,
			Aktivitas:        sql.NullString{String: req.Aktivitas, Valid: req.Aktivitas != ""},
			Interval:         sql.NullString{String: req.Interval, Valid: req.Interval != ""},
			TanggalTerjadwal: req.ScheduledDate,
			Status:           "Terjadwal",
		}
		if techID.Valid {
			var tech models.AppUser
			if config.DB.First(&tech, techID.Int64).Error == nil {
				pm.Petugas = sql.NullString{String: tech.FullName, Valid: true}
			}
		}
		config.DB.Create(&pm)
		appWO.PmID = sql.NullString{String: pmID, Valid: true}
	}

	config.DB.Create(&appWO)

	WriteJSON(w, http.StatusCreated, map[string]string{
		"message":   "Work order berhasil dibuat.",
		"wo_number": woNumber,
	})
}

func handleAssign(w http.ResponseWriter, r *http.Request, woNum string) {
	SetJSONHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	user, ok := middleware.GetCurrentUser(r)
	if !ok || user.Role != "Manager" {
		WriteError(w, http.StatusForbidden, "Hanya Manager yang dapat menugaskan teknisi")
		return
	}
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
		return
	}

	var req struct {
		TechnicianID uint `json:"technician_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.TechnicianID == 0 {
		WriteError(w, http.StatusBadRequest, "ID teknisi diperlukan")
		return
	}

	var appWO models.AppWorkOrder
	if err := config.DB.Where("wo_number = ?", woNum).First(&appWO).Error; err != nil {
		WriteError(w, http.StatusNotFound, "Work order tidak ditemukan")
		return
	}
	if technicianHasScheduleConflict(req.TechnicianID, appWO.ScheduledDate, woNum) {
		WriteError(w, http.StatusBadRequest, "Teknisi sudah memiliki work order aktif pada tanggal tersebut.")
		return
	}

	res := config.DB.Exec(`
		UPDATE app_work_order
		SET technician_id = ?, workflow_status = 'Ditugaskan'
		WHERE wo_number = ? AND workflow_status IN ('Menunggu Penugasan', 'Ditugaskan')
	`, req.TechnicianID, woNum)
	if res.Error != nil || res.RowsAffected == 0 {
		WriteError(w, http.StatusBadRequest, "Gagal menugaskan teknisi atau work order tidak dapat ditugaskan ulang")
		return
	}

	var tech models.AppUser
	if config.DB.First(&tech, req.TechnicianID).Error == nil {
		config.DB.Exec("UPDATE work_order SET teknisi = ? WHERE wo_number = ?", tech.FullName, woNum)
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "Teknisi berhasil ditugaskan."})
}

func handleReview(w http.ResponseWriter, r *http.Request, woNum string) {
	SetJSONHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	user, ok := middleware.GetCurrentUser(r)
	if !ok || user.Role != "Manager" {
		WriteError(w, http.StatusForbidden, "Hanya Manager yang dapat meninjau laporan")
		return
	}
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
		return
	}

	var req struct {
		Action      string `json:"action"`
		ManagerNote string `json:"manager_note"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	switch req.Action {
	case "approve":
		config.DB.Exec("UPDATE work_order SET status = 'Selesai', verifikasi = ? WHERE wo_number = ?", user.FullName, woNum)
		config.DB.Exec("UPDATE app_work_order SET workflow_status = 'Selesai' WHERE wo_number = ?", woNum)
		config.DB.Exec("UPDATE app_report SET manager_note = '' WHERE wo_number = ?", woNum)
		// Update pm_schedule if linked
		config.DB.Exec(`
			UPDATE pm_schedule pm
			JOIN app_work_order a ON a.pm_id = pm.pm_id
			SET pm.status = 'Terlaksana', pm.tanggal_realisasi = CURDATE()
			WHERE a.wo_number = ?`, woNum)
	case "revise":
		if req.ManagerNote == "" {
			WriteError(w, http.StatusBadRequest, "Catatan revisi wajib diisi")
			return
		}
		config.DB.Exec("UPDATE work_order SET verifikasi = ? WHERE wo_number = ?",
			"Revisi:"+req.ManagerNote, woNum)
		config.DB.Exec("UPDATE app_work_order SET workflow_status = 'Perlu Revisi' WHERE wo_number = ?", woNum)
		config.DB.Exec("UPDATE app_report SET manager_note = ? WHERE wo_number = ?", req.ManagerNote, woNum)
	default:
		WriteError(w, http.StatusBadRequest, "Aksi tidak dikenal (gunakan 'approve' atau 'revise')")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Laporan diperbarui."})
}

func handleReport(w http.ResponseWriter, r *http.Request, woNum string) {
	SetJSONHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
		return
	}
	user, ok := middleware.GetCurrentUser(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "Tidak terautentikasi")
		return
	}

	var appWO models.AppWorkOrder
	if err := config.DB.Where("wo_number = ?", woNum).First(&appWO).Error; err != nil {
		WriteError(w, http.StatusNotFound, "Work order tidak ditemukan")
		return
	}
	if user.Role != "Manager" {
		if !appWO.TechnicianID.Valid || uint(appWO.TechnicianID.Int64) != user.UserID {
			WriteError(w, http.StatusForbidden, "Anda tidak ditugaskan untuk work order ini")
			return
		}
	}

	if err := r.ParseMultipartForm(12 << 20); err != nil {
		r.ParseForm()
	}

	workDate    := r.FormValue("work_date")
	beforeCond  := r.FormValue("before_condition")
	afterCond   := r.FormValue("after_condition")
	if workDate == "" || beforeCond == "" || afterCond == "" {
		WriteError(w, http.StatusBadRequest, "Tanggal, kondisi sebelum, dan kondisi sesudah wajib diisi.")
		return
	}

	var beforePhotoName, afterPhotoName string
	if r.MultipartForm != nil {
		if f := r.MultipartForm.File["before_photo"]; len(f) > 0 {
			beforePhotoName = saveMultipartPhoto(f[0])
		}
		if f := r.MultipartForm.File["after_photo"]; len(f) > 0 {
			afterPhotoName = saveMultipartPhoto(f[0])
		}
	}

	var report models.AppReport
	isNew := config.DB.Where("wo_number = ?", woNum).First(&report).Error != nil
	report.WoNumber        = woNum
	report.WorkDate        = workDate
	report.BeforeCondition = beforeCond
	report.AfterCondition  = afterCond
	if beforePhotoName != "" {
		report.BeforePhoto = sql.NullString{String: beforePhotoName, Valid: true}
	}
	if afterPhotoName != "" {
		report.AfterPhoto = sql.NullString{String: afterPhotoName, Valid: true}
	}
	report.ManagerNote = ""
	report.SubmittedAt = sql.NullString{String: time.Now().Format("2006-01-02 15:04:05"), Valid: true}

	if isNew {
		config.DB.Create(&report)
	} else {
		config.DB.Save(&report)
		config.DB.Where("wo_number = ?", woNum).Delete(&models.AppReportDetail{})
		config.DB.Where("wo_number = ?", woNum).Delete(&models.HistoriKerusakan{})
	}

	// Save detail cards
	var damageDescs, repairActions []string
	var damagePhotos, repairPhotos []*multipart.FileHeader
	if r.MultipartForm != nil {
		damageDescs  = r.MultipartForm.Value["damage_description[]"]
		repairActions = r.MultipartForm.Value["repair_action[]"]
		damagePhotos = r.MultipartForm.File["damage_photo[]"]
		repairPhotos = r.MultipartForm.File["repair_photo[]"]
	}

	var wo models.WorkOrder
	config.DB.First(&wo, "wo_number = ?", woNum)

	// Record initial condition into kondisi_awal
	if wo.IDAsset != 0 {
		okNg := "NG"
		if strings.EqualFold(beforeCond, "baik") || strings.EqualFold(beforeCond, "normal") {
			okNg = "OK"
		}
		kondisi := models.KondisiAwal{
			Tanggal:        workDate,
			IDAsset:        wo.IDAsset,
			KomponenItem:   fmt.Sprintf("Kondisi sebelum maintenance %s", woNum),
			KondisiAktual:  sql.NullString{String: beforeCond, Valid: true},
			OkNg:           sql.NullString{String: okNg, Valid: true},
			Temuan:         sql.NullString{String: beforeCond, Valid: true},
			Petugas:        sql.NullString{String: user.FullName, Valid: true},
		}
		if beforePhotoName != "" {
			kondisi.BuktiLink = sql.NullString{String: beforePhotoName, Valid: true}
		}
		config.DB.Create(&kondisi)
	}

	for i, desc := range damageDescs {
		desc = strings.TrimSpace(desc)
		if desc == "" {
			continue
		}
		repair := ""
		if i < len(repairActions) {
			repair = strings.TrimSpace(repairActions[i])
		}

		detail := models.AppReportDetail{
			WoNumber:          woNum,
			DamageDescription: desc,
			RepairAction:      repair,
		}

		if f := r.MultipartForm.File[fmt.Sprintf("damage_photo_%d", i)]; len(f) > 0 {
			if name := saveMultipartPhoto(f[0]); name != "" {
				detail.DamagePhoto = sql.NullString{String: name, Valid: true}
			}
		} else if existing := r.FormValue(fmt.Sprintf("existing_damage_photo_%d", i)); existing != "" {
			detail.DamagePhoto = sql.NullString{String: existing, Valid: true}
		} else if i < len(damagePhotos) {
			if name := saveMultipartPhoto(damagePhotos[i]); name != "" {
				detail.DamagePhoto = sql.NullString{String: name, Valid: true}
			}
		}

		if f := r.MultipartForm.File[fmt.Sprintf("repair_photo_%d", i)]; len(f) > 0 {
			if name := saveMultipartPhoto(f[0]); name != "" {
				detail.RepairPhoto = sql.NullString{String: name, Valid: true}
			}
		} else if existing := r.FormValue(fmt.Sprintf("existing_repair_photo_%d", i)); existing != "" {
			detail.RepairPhoto = sql.NullString{String: existing, Valid: true}
		} else if i < len(repairPhotos) {
			if name := saveMultipartPhoto(repairPhotos[i]); name != "" {
				detail.RepairPhoto = sql.NullString{String: name, Valid: true}
			}
		}

		// Create histori_kerusakan entry
		failureID := fmt.Sprintf("F-%s", strings.ToUpper(uuid.NewString()[:12]))
		if len(failureID) > 30 {
			failureID = failureID[:30]
		}
		detail.FailureID = sql.NullString{String: failureID, Valid: true}
		config.DB.Create(&detail)

		if wo.IDAsset != 0 {
			hk := models.HistoriKerusakan{
				FailureID:         failureID,
				IDAsset:           wo.IDAsset,
				WoNumber:          sql.NullString{String: woNum, Valid: true},
				TanggalJamMulai:   wo.TanggalJamLapor,
				TanggalJamSelesai: sql.NullString{String: workDate + " 17:00:00", Valid: true},
				Komponen:          sql.NullString{String: desc, Valid: true},
				Gejala:            sql.NullString{String: desc, Valid: true},
				TindakanPerbaikan: sql.NullString{String: repair, Valid: true},
				Teknisi:           sql.NullString{String: user.FullName, Valid: true},
			}
			if detail.DamagePhoto.Valid {
				hk.BuktiLink = detail.DamagePhoto
			} else if detail.RepairPhoto.Valid {
				hk.BuktiLink = detail.RepairPhoto
			}
			config.DB.Create(&hk)
		}
	}

	config.DB.Exec(`
		UPDATE work_order
		SET status = 'Proses',
		    mulai_pekerjaan = COALESCE(mulai_pekerjaan, ?),
		    selesai_pekerjaan = ?,
		    tindakan = ?,
		    bukti_link = COALESCE(?, bukti_link),
		    verifikasi = 'Menunggu Approval'
		WHERE wo_number = ?
	`, workDate+" 09:00:00", workDate+" 17:00:00", afterCond, afterPhotoName, woNum)
	config.DB.Exec("UPDATE app_work_order SET workflow_status = 'Menunggu Approval' WHERE wo_number = ?", woNum)

	// If linked PM, update realization
	config.DB.Exec(`
		UPDATE pm_schedule pm
		JOIN app_work_order a ON a.pm_id = pm.pm_id
		SET pm.tanggal_realisasi = ?, pm.hasil = ?, pm.tindakan = ?
		WHERE a.wo_number = ?
	`, workDate, afterCond, afterCond, woNum)

	WriteJSON(w, http.StatusOK, map[string]string{"message": "Laporan berhasil dikirim."})
}

// technicianHasScheduleConflict checks if a technician is already booked on scheduledDate.
func technicianHasScheduleConflict(techID uint, scheduledDate string, excludeWo string) bool {
	var count int64
	q := config.DB.Model(&models.AppWorkOrder{}).
		Where("technician_id = ? AND scheduled_date = ? AND workflow_status IN ('Ditugaskan','Dikerjakan','Perlu Revisi')", techID, scheduledDate)
	if excludeWo != "" {
		q = q.Where("wo_number != ?", excludeWo)
	}
	q.Count(&count)
	return count > 0
}

// fetchWorkOrderRow returns a single work order enriched as a map.
func fetchWorkOrderRow(woNum string) (map[string]interface{}, error) {
	dbRows, err := config.DB.Raw(`
		SELECT w.wo_number, w.id_asset, w.tanggal_jam_lapor, w.pelapor,
		       w.jenis, w.prioritas, w.keluhan_gejala, w.biaya, w.status AS wo_status,
		       w.verifikasi, w.teknisi,
		       a.kode_asset, a.nama_mesin, a.merk, a.tipe_model, a.serial_number,
		       a.lokasi, a.kondisi, a.status AS asset_status,
		       app.scheduled_date, app.technician_id, app.workflow_status,
		       app.maintenance_type, app.pm_id,
		       u.full_name AS technician_name,
		       r.report_id, r.work_date, r.before_condition, r.after_condition,
		       r.before_photo, r.after_photo, r.manager_note
		FROM work_order w
		JOIN master_asset a ON a.id_asset = w.id_asset
		LEFT JOIN app_work_order app ON app.wo_number = w.wo_number
		LEFT JOIN app_users u ON u.user_id = app.technician_id
		LEFT JOIN app_report r ON r.wo_number = w.wo_number
		WHERE w.wo_number = ?
	`, woNum).Rows()
	if err != nil {
		return nil, err
	}
	defer dbRows.Close()

	cols, _ := dbRows.Columns()
	if !dbRows.Next() {
		return nil, fmt.Errorf("not found")
	}
	vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	dbRows.Scan(ptrs...)
	row := make(map[string]interface{})
	for i, col := range cols {
		v := vals[i]
		if b, ok := v.([]byte); ok {
			v = string(b)
		}
		row[col] = v
	}
	return enrichWorkOrderRow(row), nil
}

// enrichWorkOrderRow adds computed display fields to a raw map row.
func enrichWorkOrderRow(row map[string]interface{}) map[string]interface{} {
	verifikasi, _ := row["verifikasi"].(string)
	woStatus, _   := row["wo_status"].(string)
	techID        := row["technician_id"]
	workflowStatus, _ := row["workflow_status"].(string)

	effectiveStatus := workflowStatus
	if effectiveStatus == "" {
		switch {
		case verifikasi == "Menunggu Approval":
			effectiveStatus = "Menunggu Approval"
		case strings.HasPrefix(verifikasi, "Revisi:"):
			effectiveStatus = "Perlu Revisi"
		case woStatus == "Open":
			if isSet(techID) {
				effectiveStatus = "Ditugaskan"
			} else {
				effectiveStatus = "Menunggu Penugasan"
			}
		case woStatus == "Proses":
			effectiveStatus = "Dikerjakan"
		default:
			effectiveStatus = woStatus
		}
	}

	jenis, _      := row["jenis"].(string)
	prioritas, _  := row["prioritas"].(string)
	mainType, _   := row["maintenance_type"].(string)
	if mainType == "" {
		mainType = jenis
	}
	severity := priorityToSeverity[prioritas]
	if severity == "" {
		severity = prioritas
	}

	merk, _      := row["merk"].(string)
	tipe, _      := row["tipe_model"].(string)
	cat := strings.Join(filterEmpty([]string{merk, tipe}), " · ")
	if cat == "" {
		cat = "Belum diisi"
	}

	assetStatus, _ := row["asset_status"].(string)

	row["ticket_id"]        = row["wo_number"]
	row["asset_id"]         = row["id_asset"]
	row["machine_name"]     = row["nama_mesin"]
	row["location"]         = strOrBlank(row["lokasi"])
	row["category"]         = cat
	row["active"]           = assetStatus != "Afkir"
	row["status"]           = effectiveStatus
	row["workflow_status"]  = effectiveStatus
	row["maintenance_type"] = mainType
	row["severity"]         = severity
	return row
}

// ---- helpers ----------------------------------------------------------------

func isSet(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case int64:
		return val != 0
	case string:
		return val != "" && val != "0"
	case []byte:
		s := string(val)
		return s != "" && s != "0"
	}
	return true
}

func strOrBlank(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return ""
}

func filterEmpty(ss []string) []string {
	var out []string
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

func saveMultipartPhoto(fh *multipart.FileHeader) string {
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedExts[ext] {
		return ""
	}
	src, err := fh.Open()
	if err != nil {
		return ""
	}
	defer src.Close()
	os.MkdirAll(uploadDir, 0755)
	name := fmt.Sprintf("%s%s", uuid.NewString(), ext)
	dst, err := os.Create(filepath.Join(uploadDir, name))
	if err != nil {
		return ""
	}
	defer dst.Close()
	buf := make([]byte, 32*1024)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
		}
		if readErr != nil {
			break
		}
	}
	return name
}
