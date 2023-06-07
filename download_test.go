// Do以外はプライベートにしているためtestパッケージではなく、同じパッケージでテストする
package download

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

func Test_downloadInGoroutine(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("This is a test response"))
	}))
	defer server.Close()

	// テーブル駆動テスト
	type structArgs struct {
		url        string
		byteRanges []string
	}
	tests := []struct {
		name    string
		args    structArgs
		want    string
		wantErr error
	}{
		{
			name: "Test1",
			args: structArgs{
				url:        server.URL,
				byteRanges: []string{"bytes=0-7", "bytes=8-15", "bytes=16-23"}, // testではモックなので意味なし
			},
			want:    "This is a test responseThis is a test responseThis is a test response",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numGoroutinesBefore := runtime.NumGoroutine() // テスト前のgoroutineの数を取得

			got, err := downloadInGoroutine(tt.args.url, tt.args.byteRanges)
			if err != tt.wantErr {
				t.Errorf("error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("got = %v, want = %v", got, tt.want)
				return
			}

			numGoroutinesAfter := runtime.NumGoroutine() // テスト後のgoroutineの数を取得
			// テスト前と後のgoroutineの数が一致しているかを確認
			if numGoroutinesBefore != numGoroutinesAfter {
				t.Errorf("Number of goroutines before and after the test do not match: before = %d, after = %d", numGoroutinesBefore, numGoroutinesAfter)
			}
		})
	}
}
