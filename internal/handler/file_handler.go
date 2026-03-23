package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"gin-rocket/internal/service"
	"gin-rocket/pkg/response"
	"gin-rocket/pkg/storage"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService service.FileService
}

type deleteFileRequest struct {
	ObjectKey string `json:"object_key" binding:"required" example:"avatar/2026/03/23/uuid.png"`
}

func NewFileHandler(fileService service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

// Upload godoc
// @Summary Upload single file
// @Description Upload one file to MinIO under a business category such as avatar or docs.
// @Tags files
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param category formData string false "business category"
// @Param file formData file true "file"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Failure 503 {object} response.Body
// @Router /api/v1/files/upload [post]
func (h *FileHandler) Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "文件参数无效")
		return
	}

	info, err := h.fileService.Upload(c.Request.Context(), strings.TrimSpace(c.PostForm("category")), fileHeader)
	if err != nil {
		h.writeFileError(c, err, "上传文件失败")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"file": info})
}

// UploadMultiple godoc
// @Summary Upload multiple files
// @Description Upload multiple files to MinIO under the same business category.
// @Tags files
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param category formData string false "business category"
// @Param files formData file true "files"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Failure 503 {object} response.Body
// @Router /api/v1/files/upload-multiple [post]
func (h *FileHandler) UploadMultiple(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "多文件参数无效")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.Fail(c, http.StatusBadRequest, "至少上传一个文件")
		return
	}

	uploadedFiles, err := h.fileService.UploadMultiple(c.Request.Context(), strings.TrimSpace(c.PostForm("category")), files)
	if err != nil {
		h.writeFileError(c, err, "批量上传文件失败")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"files": uploadedFiles})
}

// Delete godoc
// @Summary Delete file
// @Description Delete a file from MinIO by object key.
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body deleteFileRequest true "delete request"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Failure 503 {object} response.Body
// @Router /api/v1/files [delete]
func (h *FileHandler) Delete(c *gin.Context) {
	var req deleteFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "删除请求参数无效")
		return
	}

	if err := h.fileService.Delete(c.Request.Context(), req.ObjectKey); err != nil {
		h.writeFileError(c, err, "删除文件失败")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"object_key": req.ObjectKey})
}

// Presign godoc
// @Summary Generate presigned URL
// @Description Generate a temporary file access URL by object key.
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param object_key query string true "object key"
// @Param expires query string false "expiry duration, e.g. 15m"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Failure 503 {object} response.Body
// @Router /api/v1/files/presign [get]
func (h *FileHandler) Presign(c *gin.Context) {
	objectKey := strings.TrimSpace(c.Query("object_key"))
	if objectKey == "" {
		response.Fail(c, http.StatusBadRequest, "object_key 不能为空")
		return
	}

	var expiry time.Duration
	if expiryText := strings.TrimSpace(c.Query("expires")); expiryText != "" {
		parsedExpiry, err := time.ParseDuration(expiryText)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "expires 参数无效")
			return
		}
		expiry = parsedExpiry
	}

	presignedURL, err := h.fileService.GetPresignedURL(c.Request.Context(), objectKey, expiry)
	if err != nil {
		h.writeFileError(c, err, "生成预签名 URL 失败")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"object_key": objectKey, "url": presignedURL})
}

func (h *FileHandler) writeFileError(c *gin.Context, err error, defaultMessage string) {
	switch {
	case errors.Is(err, service.ErrFileStorageDisabled):
		response.Fail(c, http.StatusServiceUnavailable, "MinIO 未启用")
	case errors.Is(err, storage.ErrInvalidObjectKey):
		response.Fail(c, http.StatusBadRequest, "object_key 无效")
	default:
		response.Fail(c, http.StatusInternalServerError, defaultMessage)
	}
}
