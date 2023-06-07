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

// downloadInGoroutineは指定されたURLからデータを並行処理でダウンロードします。
// ダウンロードしたデータは最終的にまとめられ、文字列として返されます。
func downloadInGoroutine(url string, byteRanges []string) (string, error) {
	var segmentedData []string = make([]string, len(byteRanges))
	eg, ctx := errgroup.WithContext(context.Background())
	client := new(http.Client) // 先にクライアントを作成しておきfor文でgoroutine毎にコネクションを再利用して効率を上げる
	// ↓テストでgoroutineのリークが発生しないようにする場合、ただし効率は下がる
	// client := &http.Client{
	// 	Transport: &http.Transport{
	// 		DisableKeepAlives: true, // HTTPコネクションを再利用しない
	// 	},
	// }
	for i, byteRange := range byteRanges {
		i := i
		byteRange := byteRange
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return nil
			default:
				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					return err
				}
				req.Header.Set("Range", byteRange)
				resp, err := client.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				byteData, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				segmentedData[i] = fmt.Sprint(string(byteData))
				return nil
			}
		})
	}
	if err := eg.Wait(); err != nil {
		return "", err
	}
	var allData string
	for _, v := range segmentedData {
		allData += v
	}
	return allData, nil
}

// canSegmentedDownloadは分割ダウンロードが可能かどうかを判定します。
func canSegmentedDownload(resp *http.Response) bool {
	// Accept-Rangesヘッダーが存在しない場合は一括ダウンロード
	acceptRanges := resp.Header.Get("Accept-Ranges")
	if acceptRanges == "" || acceptRanges != "bytes" {
		return false
	}
	// Content-Lengthヘッダーが存在しない場合は一括ダウンロード
	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return false
	}
	return true
}

// getContentLengthは指定されたURLのコンテンツの長さを取得します。
func getContentLength(resp *http.Response) (int, error) {
	contentLength := resp.Header.Get("Content-Length") // canSegmentedDownloadでチェック済み
	intCtntLen, err := strconv.Atoi(contentLength)
	if err != nil {
		return 0, err
	}
	return intCtntLen, nil
}

// makeRangesは各々の並行ダウンロード範囲を計算します。
// 分割数と全体の長さを引数にとり、各範囲を表す文字列のスライスを返します。
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

// createFileは指定されたURL名から、ダウンロードしたデータをファイルに保存します。
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

// segmentedDownloadは指定されたURLからデータを分割ダウンロードします。
func segmentedDownload(resp *http.Response, url string, divNum int) error {
	contentLength, err := getContentLength(resp)
	if err != nil {
		return err
	}
	byteRanges := makeRanges(divNum, contentLength)
	allData, err := downloadInGoroutine(url, byteRanges)
	if err != nil {
		return err
	}
	err = createFile(url, allData)
	if err != nil {
		return err
	}
	return nil
}

// batchDownloadは指定されたURLからデータを一括ダウンロードします。
func batchDownload(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	byteData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = createFile(url, string(byteData))
	if err != nil {
		return err
	}
	return nil
}

// Doは指定されたURLからデータをダウンロードします。
// バイト範囲リクエストが可能な場合は分割ダウンロードを行い、それ以外の場合は一括ダウンロードを行います。
func Do(url string, divNum int) error {
	resp, err := http.Head(url)
	if err != nil {
		return err
	}
	flag := canSegmentedDownload(resp)
	if flag {
		return segmentedDownload(resp, url, divNum)
	} else {
		return batchDownload(url)
	}
}
