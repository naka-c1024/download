# 分割ダウンローダ

このパッケージは、Go言語で書かれた分割ダウンロード用のコードです。指定されたURLからデータをダウンロードし、バイト範囲リクエストが可能な場合は分割ダウンロードを行います。それ以外の場合は一括ダウンロードを行います。

## 使用方法

### cloneする場合

- clone

```
git clone https://github.com/naka-c1024/download.git
```

- main関数がある階層まで移動

```
cd download/cmd/download
```

- ビルド

```
go build .
```

- 実行

```
./download https://example.com/path/to/file
```

### ローカルでmain関数を作成する場合

以下のように呼び出して使用します。

```go
package main

import "github.com/naka-c1024/download"

func main() {
	err := download.Do("https://example.com/path/to/file", 5)
	if err != nil {
		// handle error
	}
}
```

この例では、指定したURLからファイルを5つに分割してダウンロードします。

## 注意事項

このパッケージは並行ダウンロードをサポートしていますが、指定されたURLがバイト範囲リクエストをサポートしていない場合は、一括ダウンロードを行います。
また、並行処理はgoroutineを起動する分の時間が必要なので、軽いデータでは`wget`コマンドより遅くなります。

### バイト範囲リクエストの確認方法

```
curl -I https://example.com/path/to/file
```

ヘッダの中に`Accept-Ranges: bytes`が存在していたらバイト範囲リクエストが可能です。
