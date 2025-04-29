package controller

import (
	"strconv"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/repo/file"
	"github.com/gin-gonic/gin"
)

type FileController struct {
	FileRepo file.FileRepo
}

func NewFileController(repo file.FileRepo) *FileController {
	return &FileController{FileRepo: repo}
}

func (bc *FileController) GetFile(ctx *gin.Context) {
	id := ctx.Param("id")
	download := ctx.DefaultQuery("download", "")

	blob, err := bc.FileRepo.GetByID(ctx.Request.Context(), id)
	if err != nil || blob == nil {
		handler.HandleResponse(ctx, err, "file not found")
		return
	}

	ctx.Header("Content-Type", blob.MimeType)
	ctx.Header("Content-Length", strconv.FormatInt(blob.Size, 10))
	if download != "" {
		ctx.Header("Content-Disposition", "attachment; filename=\""+download+"\"")
	}

	ctx.Data(200, blob.MimeType, blob.Content)
}
