package filter

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/go-getter"
)

type listSource struct {
	Path    string
	IsBlock bool
}

func (s *listSource) Open() (src io.ReadCloser, err error) {
	if strings.HasPrefix(s.Path, "s3::") {
		return getObjectFromS3(s.Path)
	}
	return openFile(s.Path)
}

func openFile(path string) (src io.ReadCloser, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func getObjectFromS3(uri string) (src io.ReadCloser, err error) {
	newURI, err := buildURLWithArgs(uri)
	if err != nil {
		return nil, err
	}

	dst := computeCacheKey(uri)
	err = getter.GetFile(dst, newURI)
	if err != nil {
		perr := err
		if _, err := os.Stat(dst); errors.Is(err, os.ErrNotExist) {
			return nil, perr
		}
	}
	return openFile(dst)
}

func buildURLWithArgs(uri string) (string, error) {
	uri = strings.TrimPrefix(uri, "s3::")

	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("aws_access_key_id", os.Getenv("AWS_ACCESS_KEY_ID"))
	q.Add("aws_access_key_secret", os.Getenv("AWS_SECRET_ACCESS_KEY"))
	q.Add("region", os.Getenv("AWS_REGION"))
	u.RawQuery = q.Encode()

	return "s3::" + u.String(), nil
}

func computeCacheKey(uri string) string {
	h := sha256.New()
	h.Write([]byte(uri))
	return hex.EncodeToString(h.Sum(nil))
}
