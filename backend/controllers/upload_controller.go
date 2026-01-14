package controllers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	cloudinary "github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// UploadProductImage handles POST /api/uploads/cloudinary
// Accepts multipart/form-data with field name "file" and uploads to Cloudinary.
func UploadProductImage(w http.ResponseWriter, r *http.Request) {
	// Limit upload size to ~10MB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	if err := r.ParseMultipartForm(12 << 20); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid multipart form", err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing file field 'file'", err)
		return
	}
	defer file.Close()

	if err := validateImageHeader(header); err != nil {
		respondWithError(w, http.StatusBadRequest, "Unsupported image type", err)
		return
	}

	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		respondWithError(w, http.StatusInternalServerError, "Cloudinary env vars not configured", fmt.Errorf("missing CLOUDINARY_* env"))
		return
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cloudinary init failed", err)
		return
	}

	ctx := r.Context()
	// Folder structure helps organization in Cloudinary
	folder := r.FormValue("folder")
	if folder == "" {
		folder = "bakeflow/products"
	}

	// UseTimestamp to help uniqueness and cache busting
	uploadParams := uploader.UploadParams{
		Folder:       folder,
		ResourceType: "image",
	}

	res, err := cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Image upload failed", err)
		return
	}

	// Return secure URL and public ID
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"url":        res.SecureURL,
		"public_id":  res.PublicID,
		"asset_id":   res.AssetID,
		"version":    res.Version,
		"created_at": res.CreatedAt,
	})
}

func validateImageHeader(h *multipart.FileHeader) error {
	ct := h.Header.Get("Content-Type")
	ct = strings.ToLower(ct)
	switch ct {
	case "image/jpeg", "image/jpg", "image/png", "image/webp", "image/gif":
		return nil
	default:
		return fmt.Errorf("unsupported content-type: %s", ct)
	}
}
