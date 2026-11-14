package controller

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin"

	"niulai/internal/config"
	"niulai/pkg/response"
)

const maxCoverSize = 2 << 20 // 2MB

// 读取接口只允许这种由服务端生成的文件名，防止路径穿越
var uploadFilenamePattern = regexp.MustCompile(`^[a-f0-9]{32}\.(jpg|png|webp|gif)$`)

// UploadController 商品封面上传
type UploadController struct{}

func NewUploadController() *UploadController {
	return &UploadController{}
}

// UploadCover 上传封面 POST /api/v1/admin/upload/cover
func (ctl *UploadController) UploadCover(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "请选择要上传的图片")
		return
	}
	if file.Size > maxCoverSize {
		response.Fail(c, response.CodeInvalidParam, "图片不能超过 2MB")
		return
	}

	src, err := file.Open()
	if err != nil {
		c.Error(err)
		response.Fail(c, response.CodeError, "读取图片失败")
		return
	}
	defer src.Close()

	// 只读前 512 字节嗅探真实类型，不信任扩展名与 Content-Type
	buf := make([]byte, 512)
	n, err := io.ReadFull(src, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		c.Error(err)
		response.Fail(c, response.CodeError, "读取图片失败")
		return
	}
	ext, ok := detectImageExt(buf[:n])
	if !ok {
		response.Fail(c, response.CodeInvalidParam, "仅支持 jpg / png / webp / gif 图片")
		return
	}

	name, err := randomFilename(ext)
	if err != nil {
		c.Error(err)
		response.Fail(c, response.CodeError, "生成文件名失败")
		return
	}
	dir := config.GlobalConfig.Upload.Dir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.Error(err)
		response.Fail(c, response.CodeError, "创建上传目录失败")
		return
	}
	if err := c.SaveUploadedFile(file, filepath.Join(dir, name)); err != nil {
		c.Error(err)
		response.Fail(c, response.CodeError, "保存图片失败，请检查上传目录权限")
		return
	}
	response.Success(c, gin.H{"url": "/api/v1/uploads/" + name})
}

// ServeUpload 读取已上传图片 GET /api/v1/uploads/:filename。
// 封面需由 <img> 直接加载，无法携带 Authorization，因此该接口免鉴权。
func (ctl *UploadController) ServeUpload(c *gin.Context) {
	name := c.Param("filename")
	if !uploadFilenamePattern.MatchString(name) {
		response.Fail(c, response.CodeNotFound, "图片不存在")
		return
	}
	path := filepath.Join(config.GlobalConfig.Upload.Dir, name)
	if _, err := os.Stat(path); err != nil {
		response.Fail(c, response.CodeNotFound, "图片不存在")
		return
	}
	c.File(path)
}

// detectImageExt 依据文件内容判断图片类型，返回安全的扩展名。
// webp 单独判断，避免依赖 http.DetectContentType 的版本差异。
func detectImageExt(buf []byte) (string, bool) {
	if len(buf) >= 12 && string(buf[0:4]) == "RIFF" && string(buf[8:12]) == "WEBP" {
		return ".webp", true
	}
	switch http.DetectContentType(buf) {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	default:
		return "", false
	}
}

func randomFilename(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成随机文件名失败: %w", err)
	}
	return hex.EncodeToString(b) + ext, nil
}
