# gocov - Goカバレッジ集計ツール

[![Test](https://github.com/blck-snwmn/gocov/actions/workflows/test.yml/badge.svg)](https://github.com/blck-snwmn/gocov/actions/workflows/test.yml)

## 概要

`gocov`は、Goプロジェクトのテストカバレッジをディレクトリ単位で集計・表示するコマンドラインツールです。標準の`go test`で生成されたカバレッジプロファイルを解析し、各ディレクトリのカバレッジ率を見やすい表形式で出力します。

## 特徴

- 📁 ディレクトリ単位でのカバレッジ集計
- 🎯 柔軟な階層レベルでの集計が可能
- 🔍 カバレッジ率による結果のフィルタリング機能
- 📊 見やすい表形式での結果表示
- 🚀 シンプルなコマンドラインインターフェース
- 🔧 標準のGoカバレッジプロファイル形式に対応
- ⚙️ 設定ファイルによるプロジェクトごとの設定管理
- ⚡ 大規模プロジェクト向けの並行処理サポート

## インストール

```bash
go install github.com/blck-snwmn/gocov@latest
```

または、ソースコードからビルド：

```bash
git clone https://github.com/blck-snwmn/gocov.git
cd gocov
go build -o gocov
```

## 使い方

### 基本的な使用方法

1. まず、Goプロジェクトでカバレッジプロファイルを生成します：

```bash
go test -coverprofile=coverage.out -coverpkg=./... ./...
```

2. `gocov`を使用してディレクトリ単位のカバレッジを表示します：

```bash
gocov -coverprofile=coverage.out
```

### コマンドラインオプション

- `-coverprofile`: カバレッジプロファイルファイルのパス（必須）
- `-level`: ディレクトリ階層の集計レベル（デフォルト: 0）
  - `0`: 末端ディレクトリごとに集計（デフォルト）
  - `N` (N > 0): パスの最初のN個の要素で集計
  - `-1`: トップレベルで全体を集計
- `-min`: 表示する最小カバレッジ率（0-100、デフォルト: 0）
- `-max`: 表示する最大カバレッジ率（0-100、デフォルト: 100）
- `-format`: 出力フォーマット（table または json、デフォルト: table）
- `-ignore`: 無視するディレクトリのカンマ区切りリスト（ワイルドカード対応）
- `-concurrent`: 大きなカバレッジファイルに対して並行処理を使用（デフォルト: false）
- `-config`: 設定ファイルのパス（デフォルト: カレントディレクトリから上位に向かって`.gocov.yml`を検索）
- `-threshold`: 最小総カバレッジしきい値（0-100、デフォルト: 0）

### 出力例

デフォルト（末端ディレクトリごと）:
```
$ gocov -coverprofile=coverage.out
Directory                                          Statements    Covered Coverage
--------------------------------------------------------------------------------
github.com/example/project/cmd/server                      7          5   71.4%
github.com/example/project/internal/service                7          6   85.7%
github.com/example/project/pkg/util                        7          5   71.4%
--------------------------------------------------------------------------------
TOTAL                                                     21         16   76.2%
```

レベル4で集計（pkg, cmd, internal単位）:
```
$ gocov -coverprofile=coverage.out -level 4
Directory                                          Statements    Covered Coverage
--------------------------------------------------------------------------------
github.com/example/project/cmd                             7          5   71.4%
github.com/example/project/internal                        7          6   85.7%
github.com/example/project/pkg                             7          5   71.4%
--------------------------------------------------------------------------------
TOTAL                                                     21         16   76.2%
```

トップレベルで集計:
```
$ gocov -coverprofile=coverage.out -level -1
Directory                                          Statements    Covered Coverage
--------------------------------------------------------------------------------
.                                                         21         16   76.2%
--------------------------------------------------------------------------------
TOTAL                                                     21         16   76.2%
```

最小カバレッジ率でフィルタリング（75%以上のみ表示）:
```
$ gocov -coverprofile=coverage.out -min 75
Directory                                          Statements    Covered Coverage
--------------------------------------------------------------------------------
github.com/example/project/internal/service                7          6   85.7%
--------------------------------------------------------------------------------
FILTERED TOTAL                                              7          6   85.7%
TOTAL                                                     21         16   76.2%
```

最大カバレッジ率でフィルタリング（75%以下のみ表示）:
```
$ gocov -coverprofile=coverage.out -max 75
Directory                                          Statements    Covered Coverage
--------------------------------------------------------------------------------
github.com/example/project/cmd/server                      7          5   71.4%
github.com/example/project/pkg/util                        7          5   71.4%
--------------------------------------------------------------------------------
FILTERED TOTAL                                             14         10   71.4%
TOTAL                                                     21         16   76.2%
```

JSON形式での出力:
```
$ gocov -coverprofile=coverage.out -format json -min 75
{
  "results": [
    {
      "directory": "github.com/example/project/internal/service",
      "statements": 7,
      "covered": 6,
      "coverage": 85.71428571428571
    }
  ],
  "total": {
    "directory": "TOTAL",
    "statements": 21,
    "covered": 16,
    "coverage": 76.19047619047619
  },
  "filtered_total": {
    "directory": "FILTERED TOTAL",
    "statements": 7,
    "covered": 6,
    "coverage": 85.71428571428571
  }
}
```

特定のディレクトリを無視して集計:
```
$ gocov -coverprofile=coverage.out -ignore "*/internal/*,*/vendor/*"
Directory                                          Statements    Covered Coverage
--------------------------------------------------------------------------------
github.com/example/project/cmd/server                      7          5   71.4%
github.com/example/project/pkg/util                        7          5   71.4%
--------------------------------------------------------------------------------
TOTAL                                                     14         10   71.4%
```

カバレッジしきい値チェック（CIでの使用に最適）:
```
$ gocov -coverprofile=coverage.out -threshold 80
Directory                                          Statements    Covered Coverage
--------------------------------------------------------------------------------
github.com/example/project/cmd/server                      7          5   71.4%
github.com/example/project/internal/service                7          6   85.7%
github.com/example/project/pkg/util                        7          5   71.4%
--------------------------------------------------------------------------------
TOTAL                                                     21         16   76.2%
2025/05/25 15:00:00 coverage 76.2% is below threshold 80.0%
# Exit code: 1
```

## 設定ファイル

`gocov`は`.gocov.yml`という設定ファイルをサポートしています。設定ファイルを使用することで、プロジェクトごとの設定を永続化し、コマンドラインオプションを簡略化できます。

### 設定ファイルの検索順序

1. `-config`フラグで指定されたパス
2. カレントディレクトリから親ディレクトリに向かって`.gocov.yml`を検索

### 設定ファイルの例

```yaml
# gocov configuration file

# ディレクトリ階層の集計レベル
level: 0

# カバレッジ率のフィルタリング
coverage:
  min: 0
  max: 100

# 出力フォーマット
format: table

# 無視するディレクトリパターン
ignore:
  - "*/vendor/*"
  - "*/test/*"
  - "*/mock/*"
  - "*/generated/*"

# 並行処理の有効化
concurrent: true

# カバレッジしきい値（CIでの使用に最適）
threshold: 80
```

### 優先順位

コマンドラインオプションは常に設定ファイルの値を上書きします。例：

```bash
# .gocov.ymlでformat: tableが設定されていても、JSONで出力される
gocov -coverprofile=coverage.out -format json
```

## 動作原理

`gocov`は以下の手順でカバレッジを集計します：

1. 指定されたカバレッジプロファイルファイルを解析
2. 各ファイルのカバレッジ情報をディレクトリ単位で集計
3. ステートメント数とカバーされたステートメント数を計算
4. ディレクトリごとのカバレッジ率を算出
5. 結果を表形式で出力

### 並行処理

`-concurrent`フラグを有効にすると、大規模なカバレッジファイルの処理が高速化されます：

- 10ファイル以下の場合は自動的に通常の処理を使用
- 11ファイル以上の場合はワーカープールパターンで並行処理
- 最大4つのワーカーが同時にファイルを処理
- ファイル数が増えるほど処理時間の短縮効果が大きくなります

## ユースケース

- 大規模なGoプロジェクトでのカバレッジ管理
- CI/CDパイプラインでのカバレッジレポート生成（JSON出力を活用）
- コードレビュー時のカバレッジ確認
- モジュール単位でのテスト品質の可視化
- カバレッジが低いディレクトリの特定（-minフラグを活用）
- テストやvendorディレクトリを除外したカバレッジ分析（-ignoreフラグを活用）

## 制限事項

- カバレッジプロファイルファイルは事前に生成する必要があります
- 出力先は標準出力のみ（リダイレクトで保存可能）

## 必要要件

- Go 1.16以上
- `golang.org/x/tools/cover`パッケージ