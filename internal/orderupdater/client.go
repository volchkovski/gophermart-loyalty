package orderupdater

import (
	"github.com/go-resty/resty/v2"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"net/http"
	"strconv"
	"time"
)

func newRestyClient() *resty.Client {
	return resty.New().
		SetRetryCount(3).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				return r.StatusCode() == http.StatusTooManyRequests
			},
		).
		SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			retryAfter := r.Header().Get("Retry-After")
			if retryAfter == "" {
				return 0, nil
			}
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				return time.Duration(seconds) * time.Second, nil
			}
			return 0, nil
		}).
		OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
			if r.StatusCode() == http.StatusTooManyRequests {
				logger.Log.Infof("Received 429, retry after: %s\n", r.Header().Get("Retry-After"))
			}
			return nil
		})
}
