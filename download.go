package download

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/sync/errgroup"
)

func downloadInGoroutine(url string, arrRange []string, divNum int) (string, error) {
	var splitData []string = make([]string, divNum)
	eg, ctx := errgroup.WithContext(context.Background())
	for i, ctxRange := range arrRange {
		i := i
		ctxRange := ctxRange
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return nil
			default:
				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					return err
				}
				req.Header.Set("Range", ctxRange)
				client := new(http.Client)
				resp, err := client.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				byteArray, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				splitData[i] = fmt.Sprint(string(byteArray))
				return nil
			}
		})
	}
	if err := eg.Wait(); err != nil {
		return "", err
	}
	var allData string
	for _, v := range splitData {
		allData += v
	}
	return allData, nil
}

func hasAcceptRangesBytes(url string) (bool, error) {
	res, err := http.Head(url)
	if err != nil {
		return false, err
	}
	acceptRanges := res.Header.Get("Accept-Ranges")
	if acceptRanges == "bytes" {
		return true, nil
	} else {
		return false, nil
	}
}

func getContentLength(url string) (int, error) {
	res, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	contentLength := res.Header.Get("Content-Length")
	intCtntLen, err := strconv.Atoi(contentLength)
	if err != nil {
		return 0, err
	}
	return intCtntLen, nil
}

func makeRanges(num int, length int) []string {
	var result []string
	div := length / num
	start := 0
	end := div
	for length > 0 {
		str := fmt.Sprintf("bytes=%d-%d", start, end)
		start = end + 1
		length -= div
		if length < 0 {
			break
		} else {
			end = start + div
		}
		result = append(result, str)
	}
	return result
}

func createFile(url string, content string) (err error) {
	basename := filepath.Base(url)
	f, err := os.Create(basename)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := f.Close()
		// 他にエラーがなければ戻り値のerrに代入
		if err == nil {
			err = closeErr
		}
	}()
	_, err = fmt.Fprint(f, content)
	if err != nil {
		return err
	}
	return nil
}

func segmentedDownload(url string, divNum int) error {
	contentLength, err := getContentLength(url)
	if err != nil {
		return err
	}
	arrRange := makeRanges(divNum, contentLength)
	allData, err := downloadInGoroutine(url, arrRange, divNum)
	if err != nil {
		return err
	}
	err = createFile(url, allData)
	if err != nil {
		return err
	}
	return nil
}

func batchDownload(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = createFile(url, string(byteArray))
	if err != nil {
		return err
	}
	return nil
}

func Do(url string, divNum int) error {
	byteFlag, err := hasAcceptRangesBytes(url)
	if err != nil {
		return err
	}
	if byteFlag {
		return segmentedDownload(url, divNum)
	} else {
		return batchDownload(url)
	}
}
