package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"smart-mt-backend/config"
	"smart-mt-backend/middleware"
	"smart-mt-backend/models"

	"github.com/google/uuid"
)

const uploadDir = "./uploads"
var allowedExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

// HandleAssets routes /api/assets and /api/assets/{id}.
func HandleAssets(w http.ResponseWriter, r *http.Request) {
	SetJSONHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var id uint
	var action string
	if len(parts) >= 3 {
		parsed, _ := strconv.ParseUint(parts[2], 10, 64)
		id = uint(parsed)
	}
	if len(parts) >= 4 {
		action = parts[3]
	}

	if id > 0 && (action == "afkir" || action == "retire") {
		retireAsset(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if id > 0 {
			getAssetByID(w, r, id)
		} else {
			getAllAssets(w, r)
		}
	case http.MethodPost:
		if id > 0 {
			updateAsset(w, r, id)
		} else {
			createAsset(w, r)
		}
	case http.MethodPut:
		if id > 0 {
			updateAsset(w, r, id)
		} else {
			WriteError(w, http.StatusBadRequest, "ID aset diperlukan")
		}
	case http.MethodDelete:
		if id > 0 {
			retireAsset(w, r, id)
		} else {
			WriteError(w, http.StatusBadRequest, "ID aset diperlukan")
		}
	default:
		WriteError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
	}
}

// getAllAssets returns all assets with photo count and WO count.
func getAllAssets(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Raw(`
		SELECT a.*,
		       COUNT(DISTINCT p.photo_id)  AS photo_count,
		       COUNT(DISTINCT w.wo_number) + COUNT(DISTINCT pm.pm_id) AS work_order_count,
		       (SELECT filename FROM asset_photo WHERE id_asset = a.id_asset ORDER BY photo_id LIMIT 1) AS first_photo
		FROM master_asset a
		LEFT JOIN asset_photo p  ON p.id_asset  = a.id_asset
		LEFT JOIN work_order  w  ON w.id_asset  = a.id_asset
		LEFT JOIN pm_schedule pm ON pm.id_asset = a.id_asset
		GROUP BY a.id_asset
		ORDER BY a.nama_mesin, a.id_asset
	`).Rows()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Gagal mengambil data aset")
		return
	}
	defer rows.Close()

	assets := make([]map[string]interface{}, 0)
	for rows.Next() {
		var a models.AssetWithMeta
		if err := config.DB.ScanRows(rows, &a); err != nil {
			continue
		}
		assets = append(assets, assetToMap(a))
	}
	WriteJSON(w, http.StatusOK, assets)
}

// getAssetByID returns a single asset with full detail (photos, PM, downtime, KPI).
func getAssetByID(w http.ResponseWriter, r *http.Request, id uint) {
	var a models.AssetWithMeta
	err := config.DB.Raw(`
		SELECT a.*,
		       COUNT(DISTINCT p.photo_id)  AS photo_count,
		       COUNT(DISTINCT w.wo_number) + COUNT(DISTINCT pm.pm_id) AS work_order_count
		FROM master_asset a
		LEFT JOIN asset_photo p  ON p.id_asset  = a.id_asset
		LEFT JOIN work_order  w  ON w.id_asset  = a.id_asset
		LEFT JOIN pm_schedule pm ON pm.id_asset = a.id_asset
		WHERE a.id_asset = ?
		GROUP BY a.id_asset
	`, id).Scan(&a).Error
	if err != nil || a.IDAsset == 0 {
		WriteError(w, http.StatusNotFound, "Aset tidak ditemukan")
		return
	}

	var photos []models.AssetPhoto
	config.DB.Where("id_asset = ?", id).Order("photo_id").Find(&photos)

	var pmList []models.PmSchedule
	config.DB.Where("id_asset = ?", id).Order("tanggal_terjadwal desc").Find(&pmList)

	var failures []models.HistoriKerusakan
	config.DB.Where("id_asset = ? AND wo_number IS NULL", id).Order("tanggal_jam_mulai desc").Find(&failures)

	var downtimes []models.Downtime
	config.DB.Where("id_asset = ?", id).Order("tanggal desc, downtime_id desc").Find(&downtimes)

	var kpis []models.KpiBeforeAfter
	config.DB.Where("id_asset = ?", id).Order("periode desc").Find(&kpis)

	var kondisi []models.KondisiAwal
	config.DB.Where("id_asset = ?", id).Order("tanggal desc, id_kondisi desc").Find(&kondisi)

	var historyWOs []string
	config.DB.Raw("SELECT wo_number FROM work_order WHERE id_asset = ? ORDER BY tanggal_jam_lapor DESC", id).Scan(&historyWOs)
	history := make([]map[string]interface{}, 0, len(historyWOs))
	for _, woNum := range historyWOs {
		row, err := fetchWorkOrderRow(woNum)
		if err == nil {
			history = append(history, row)
		}
	}

	result := map[string]interface{}{
		"asset":             assetToMap(a),
		"photos":            photos,
		"history":           history,
		"pm_history":        pmList,
		"failure_history":   failures,
		"downtime_history":  downtimes,
		"kpi_history":       kpis,
		"condition_history": kondisi,
	}
	WriteJSON(w, http.StatusOK, result)
}

// createAsset handles POST /api/assets (multipart form data with optional photo files).
func createAsset(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetCurrentUser(r)
	if !ok || user.Role != "Manager" {
		WriteError(w, http.StatusForbidden, "Hanya Manager yang dapat menambah aset")
		return
	}

	r.ParseMultipartForm(12 << 20) // 12 MB
	asset := models.MasterAsset{
		KodeAsset:  strings.TrimSpace(r.FormValue("kode_asset")),
		NamaMesin:  strings.TrimSpace(r.FormValue("nama_mesin")),
		Merk:       nullStr(r.FormValue("merk")),
		TipeModel:  nullStr(r.FormValue("tipe_model")),
		SerialNumber: nullStr(r.FormValue("serial_number")),
		Lokasi:     nullStr(r.FormValue("lokasi")),
		PenanggungJawab: nullStr(r.FormValue("penanggung_jawab")),
		Status:     coalesce(r.FormValue("status"), "Operasional"),
		Kondisi:    coalesce(r.FormValue("kondisi"), "Baik"),
		SumberData: nullStr(r.FormValue("sumber_data")),
		Petugas:    nullStr(r.FormValue("petugas")),
		KapasitasSpesifikasi: nullStr(r.FormValue("kapasitas_spesifikasi")),
	}
	if y := r.FormValue("tahun_perolehan"); y != "" {
		yr, _ := strconv.ParseInt(y, 10, 16)
		asset.TahunPerolehan = sql.NullInt16{Int16: int16(yr), Valid: true}
	}
	if d := r.FormValue("tanggal_pendataan"); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err == nil {
			asset.TanggalPendataan = sql.NullTime{Time: t, Valid: true}
		}
	}
	if asset.KodeAsset == "" || asset.NamaMesin == "" {
		WriteError(w, http.StatusBadRequest, "Kode aset dan nama mesin wajib diisi.")
		return
	}

	if err := config.DB.Create(&asset).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			WriteError(w, http.StatusConflict, "Nomor seri sudah digunakan.")
		} else {
			WriteError(w, http.StatusInternalServerError, "Gagal menyimpan aset: "+err.Error())
		}
		return
	}

	// Handle uploaded photos
	if r.MultipartForm != nil {
		saveAssetPhotos(r, asset.IDAsset)
	}

	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message":  "Aset berhasil ditambahkan.",
		"id_asset": asset.IDAsset,
	})
}

// updateAsset handles PUT /api/assets/{id}.
func updateAsset(w http.ResponseWriter, r *http.Request, id uint) {
	user, ok := middleware.GetCurrentUser(r)
	if !ok || user.Role != "Manager" {
		WriteError(w, http.StatusForbidden, "Hanya Manager yang dapat mengedit aset")
		return
	}

	var existing models.MasterAsset
	if err := config.DB.First(&existing, id).Error; err != nil {
		WriteError(w, http.StatusNotFound, "Aset tidak ditemukan")
		return
	}

	r.ParseMultipartForm(12 << 20)
	existing.KodeAsset  = strOrDefault(r.FormValue("kode_asset"),  existing.KodeAsset)
	existing.NamaMesin  = strOrDefault(r.FormValue("nama_mesin"),  existing.NamaMesin)
	existing.Merk       = nullStr(r.FormValue("merk"))
	existing.TipeModel  = nullStr(r.FormValue("tipe_model"))
	existing.SerialNumber = nullStr(r.FormValue("serial_number"))
	existing.Lokasi     = nullStr(r.FormValue("lokasi"))
	existing.PenanggungJawab = nullStr(r.FormValue("penanggung_jawab"))
	if s := r.FormValue("status"); s != "" { existing.Status = s }
	if k := r.FormValue("kondisi"); k != "" { existing.Kondisi = k }
	existing.SumberData = nullStr(r.FormValue("sumber_data"))
	existing.Petugas    = nullStr(r.FormValue("petugas"))
	existing.KapasitasSpesifikasi = nullStr(r.FormValue("kapasitas_spesifikasi"))
	if y := r.FormValue("tahun_perolehan"); y != "" {
		yr, _ := strconv.ParseInt(y, 10, 16)
		existing.TahunPerolehan = sql.NullInt16{Int16: int16(yr), Valid: true}
	}
	if d := r.FormValue("tanggal_pendataan"); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err == nil {
			existing.TanggalPendataan = sql.NullTime{Time: t, Valid: true}
		}
	}

	if err := config.DB.Save(&existing).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			WriteError(w, http.StatusConflict, "Nomor seri sudah digunakan.")
		} else {
			WriteError(w, http.StatusInternalServerError, "Gagal memperbarui aset")
		}
		return
	}

	// Handle new photo uploads
	if r.MultipartForm != nil {
		saveAssetPhotos(r, id)
	}

	// Handle photo removals (comma-separated photo IDs in form field "remove_photo_ids")
	if ids := r.FormValue("remove_photo_ids"); ids != "" {
		for _, pid := range strings.Split(ids, ",") {
			pid = strings.TrimSpace(pid)
			if pid != "" {
				config.DB.Where("photo_id = ? AND id_asset = ?", pid, id).Delete(&models.AssetPhoto{})
			}
		}
	}

	// Update foto_link to the first remaining photo
	var firstPhoto models.AssetPhoto
	if config.DB.Where("id_asset = ?", id).Order("photo_id").First(&firstPhoto).Error == nil {
		config.DB.Model(&existing).Update("foto_link", firstPhoto.Filename)
	} else {
		config.DB.Model(&existing).Update("foto_link", nil)
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "Aset berhasil diperbarui."})
}

// retireAsset marks the asset as Afkir (retired) via DELETE /api/assets/{id}.
func retireAsset(w http.ResponseWriter, r *http.Request, id uint) {
	user, ok := middleware.GetCurrentUser(r)
	if !ok || user.Role != "Manager" {
		WriteError(w, http.StatusForbidden, "Hanya Manager yang dapat mengafkirkan aset")
		return
	}
	res := config.DB.Exec("UPDATE master_asset SET status = 'Afkir' WHERE id_asset = ? AND status != 'Afkir'", id)
	if res.Error != nil {
		WriteError(w, http.StatusInternalServerError, "Gagal mengafkirkan aset")
		return
	}
	if res.RowsAffected == 0 {
		WriteError(w, http.StatusNotFound, "Aset tidak ditemukan atau sudah afkir.")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Aset berhasil ditandai afkir."})
}

// HandleUpload serves uploaded files: GET /api/uploads/{filename}
func HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		SetJSONHeaders(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	filename := filepath.Base(r.URL.Path)
	if filename == "." || filename == "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(uploadDir, filename))
}

// ---- helpers ----------------------------------------------------------------

func saveAssetPhotos(r *http.Request, assetID uint) {
	os.MkdirAll(uploadDir, 0755)
	for _, fh := range r.MultipartForm.File["photos"] {
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if !allowedExts[ext] {
			continue
		}
		src, err := fh.Open()
		if err != nil {
			continue
		}
		name := fmt.Sprintf("%s%s", uuid.NewString(), ext)
		dst, err := os.Create(filepath.Join(uploadDir, name))
		if err != nil {
			src.Close()
			continue
		}
		buf := make([]byte, 32*1024)
		for {
			n, err := src.Read(buf)
			if n > 0 { dst.Write(buf[:n]) }
			if err != nil { break }
		}
		src.Close()
		dst.Close()

		photo := models.AssetPhoto{IDAsset: assetID, Filename: name}
		if config.DB.Create(&photo).Error == nil {
			// Set foto_link to first photo
			config.DB.Exec("UPDATE master_asset SET foto_link = COALESCE(foto_link, ?) WHERE id_asset = ?", name, assetID)
		}
	}
}

func nullStr(s string) sql.NullString {
	s = strings.TrimSpace(s)
	return sql.NullString{String: s, Valid: s != ""}
}

func coalesce(s, def string) string {
	if s == "" { return def }
	return s
}

func strOrDefault(s, def string) string {
	if strings.TrimSpace(s) == "" { return def }
	return strings.TrimSpace(s)
}

func assetToMap(a models.AssetWithMeta) map[string]interface{} {
	return map[string]interface{}{
		"id_asset":               a.IDAsset,
		"kode_asset":             a.KodeAsset,
		"nama_mesin":             a.NamaMesin,
		"merk":                   a.Merk.String,
		"tipe_model":             a.TipeModel.String,
		"serial_number":          a.SerialNumber.String,
		"tahun_perolehan":        a.TahunPerolehan.Int16,
		"kapasitas_spesifikasi":  a.KapasitasSpesifikasi.String,
		"lokasi":                 a.Lokasi.String,
		"penanggung_jawab":       a.PenanggungJawab.String,
		"status":                 a.Status,
		"kondisi":                a.Kondisi,
		"sumber_data":            a.SumberData.String,
		"tanggal_pendataan":      nullTimeStr(a.TanggalPendataan),
		"petugas":                a.Petugas.String,
		"foto_link":              a.FotoLink.String,
		"photo_count":            a.PhotoCount,
		"work_order_count":       a.WorkOrderCount,
		"first_photo":            a.FirstPhoto,
	}
}

func nullTimeStr(t sql.NullTime) string {
	if !t.Valid { return "" }
	return t.Time.Format("2006-01-02")
}
