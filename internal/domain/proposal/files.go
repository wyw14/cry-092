package proposal

import (
	"fmt"
	"path/filepath"
	"strings"
)

type FilePolicy struct {
	MaxSize    int64
	Extensions map[string][]string
}

func DefaultFilePolicy() FilePolicy {
	return FilePolicy{MaxSize: 20 << 20, Extensions: map[string][]string{".pdf": {"application/pdf"}, ".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document"}, ".jpg": {"image/jpeg"}, ".png": {"image/png"}}}
}

func (p FilePolicy) Validate(name, mime string, size int64, digest, storeKey string) error {
	ext := strings.ToLower(filepath.Ext(name))
	allowed, ok := p.Extensions[ext]
	if !ok || size <= 0 || size > p.MaxSize || len(digest) != 64 || strings.Contains(storeKey, "..") || filepath.IsAbs(storeKey) {
		return fmt.Errorf("file metadata rejected")
	}
	for _, value := range allowed {
		if value == mime {
			return nil
		}
	}
	return fmt.Errorf("mime does not match extension")
}
