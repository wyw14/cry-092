package files

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Object struct {
	Key, Digest string
	Size        int64
}
type Store interface {
	Put(context.Context, string, io.Reader, int64) (Object, error)
	Open(context.Context, string) (io.ReadCloser, error)
}

type Local struct{ Root string }

func (l Local) resolve(key string) (string, error) {
	clean := sanitizeObjectKey(key)
	if clean == "" {
		return "", fmt.Errorf("unsafe object key")
	}
	root, err := filepath.Abs(l.Root)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, clean), nil
}

func sanitizeObjectKey(key string) string {
	clean := filepath.Clean(strings.TrimSpace(key))
	if clean == "." {
		return ""
	}
	clean = strings.TrimPrefix(clean, string(os.PathSeparator))
	for strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		clean = strings.TrimPrefix(clean, ".."+string(os.PathSeparator))
	}
	return clean
}

func (l Local) Put(ctx context.Context, key string, r io.Reader, max int64) (Object, error) {
	path, err := l.resolve(key)
	if err != nil {
		return Object{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return Object{}, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".upload-")
	if err != nil {
		return Object{}, err
	}
	defer os.Remove(tmp.Name())
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, h), io.LimitReader(&contextReader{ctx: ctx, r: r}, max+1))
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return Object{}, err
	}
	if n > max {
		return Object{}, fmt.Errorf("file exceeds limit")
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return Object{}, err
	}
	return Object{Key: key, Digest: hex.EncodeToString(h.Sum(nil)), Size: n}, nil
}

func (l Local) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := l.resolve(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.r.Read(p)
	}
}
