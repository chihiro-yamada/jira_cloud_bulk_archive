# JIRA Cloud Bulk Archive

JIRA Cloudで任意のラベルが付与されている課題を一斉アーカイブするGoツールです。

## 機能

- 指定されたプロジェクトとラベルに基づいて課題を検索
- JIRA Cloud Bulk Archive APIを使用した効率的な一括アーカイブ処理
- 大量の課題を自動的にバッチ分割（デフォルト1000件/バッチ）
- 環境変数による設定管理（godotenv対応）
- 詳細なログ出力とサマリーレポート

## 必要要件

- Go 1.23.5以上
- JIRA Cloud APIトークン

## セットアップ

1. リポジトリをクローン:
```bash
git clone https://github.com/c_yamada/jira_cloud_bulk_archive.git
cd jira_cloud_bulk_archive
```

2. 環境変数を設定:
```bash
cp .env.example .env
# .envファイルを編集して設定値を入力
```

3. 必要な環境変数:
- `JIRA_BASE_URL`: JIRAインスタンスのURL (例: https://your-domain.atlassian.net)
- `JIRA_EMAIL`: JIRAアカウントのメールアドレス
- `JIRA_API_TOKEN`: JIRA APIトークン
- `JIRA_PROJECT_KEY`: 対象プロジェクトのキー
- `ARCHIVE_LABEL`: アーカイブ対象のラベル名 (デフォルト: archive)
- `MAX_WORKERS`: 互換性のため残していますが、現在は使用されていません

## JIRA APIトークンの取得方法

1. https://id.atlassian.com/manage-profile/security/api-tokens にアクセス
2. 「APIトークンを作成」をクリック
3. トークン名を入力して作成
4. 生成されたトークンをコピーして`JIRA_API_TOKEN`に設定

## 実行

.envファイルを使用する場合（推奨）:

```bash
# .envファイルを作成・編集した後
go run ./cmd/archive
```

環境変数を直接指定する場合:

```bash
JIRA_BASE_URL=https://your-domain.atlassian.net \
JIRA_EMAIL=your-email@example.com \
JIRA_API_TOKEN=your-api-token \
JIRA_PROJECT_KEY=YOUR_PROJECT \
ARCHIVE_LABEL=archive \
MAX_WORKERS=5 \
go run ./cmd/archive
```

**注**: godotenvを使用しているため、.envファイルがあれば自動的に読み込まれます。.envファイルが無い場合はシステムの環境変数が使用されます。

## プロジェクト構造

```
.
├── cmd/
│   └── archive/          # メインアプリケーション
├── internal/
│   ├── config/           # 設定管理
│   └── jira/             # JIRA APIクライアント
├── pkg/
│   └── worker/           # 並列処理ワーカー
├── .env.example          # 環境変数のサンプル
└── go.mod               # Go モジュール定義
```

## 注意事項

- アーカイブはPJの管理者のみ可能です。
- APIレート制限に注意してください
- 大量の課題をアーカイブする場合は`MAX_WORKERS`を適切に調整してください
