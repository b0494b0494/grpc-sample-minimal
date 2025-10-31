# OCR実装状況

作成日: 2024年11月

## 概要

現在のOCR機能の実装状況をまとめます。

## 対応ストレージプロバイダー

### ✓ 対応済みストレージプロバイダー

1. **Azure Blob Storage (Azurite)**
   - 対応済み
   - エミュレーター: Azurite
   - OCR対応: 完了

2. **AWS S3 (Localstack)**
   - 対応済み
   - エミュレーター: Localstack
   - OCR対応: 完了

3. **Google Cloud Storage (fake-gcs)**
   - 対応済み
   - エミュレーター: fake-gcs-server
   - OCR対応: 完了

## 実装済みOCRエンジン

### ✓ Tesseract OCR (現在使用中)

- **言語**: 日本語と英語
- **言語設定**: `jpn+eng` (日本語 + 英語)
- **実装ファイル**: 
  - `server/domain/tesseract_ocr.go` - Tesseract OCRエンジン実装
  - `server/ocr/main.go` - OCR独立サービス
- **Dockerfile**: `Dockerfile.ocr`
  - Tesseract OCRライブラリと日本語トレーニングデータを含む
- **設定**: 
  ```go
  // server/ocr/main.go:253-254
  tesseractEngine := domain.NewTesseractEngine("jpn+eng")
  ocrService.RegisterEngine(tesseractEngine)
  ```

## 未実装OCRエンジン（Phase 2予定）

### ☐ EasyOCR
- **状態**: 未実装
- **時期**: Phase 2で実装予定

### ☐ PaddleOCR
- **状態**: 未実装
- **時期**: Phase 2で実装予定

## 自動OCR処理

### ✓ 実装済み

`documents/` または `images/` 名前空間にファイルがアップロードされると、自動的にOCR処理をキューに登録します。

**実装場所**: `server/application/service.go:154-170`

```go
// ファイルがdocuments/またはimages/名前空間の場合はOCRをキューに追加
if (namespace == "documents" || namespace == "images") && s.ocrClient != nil {
    // バックグラウンドでOCRタスクをキューに追加
    queueService, err := domain.NewQueueService(provider)
    if err != nil {
        log.Printf("Warning: Failed to create queue service: %v", err)
    } else {
        // 非同期でOCRタスクをエンキュー
        go func() {
            if err := queueService.EnqueueOCRTask(context.Background(), filename, provider); err != nil {
                log.Printf("Warning: Failed to enqueue OCR task: %v", err)
            } else {
                log.Printf("OCR task queued for file: %s (provider: %s)", filename, provider)
            }
        }()
    }
}
```

## OCRワーカー

### ✓ 実装済み

**実装場所**: `server/ocr/main.go:306-308`

```go
// すべてのストレージプロバイダーに対してOCRワーカーを起動
ctx := context.Background()
go startOCRWorker(ctx, "azure", ocrService, ocrResultRepo, getStorageService)
go startOCRWorker(ctx, "s3", ocrService, ocrResultRepo, getStorageService)
go startOCRWorker(ctx, "gcs", ocrService, ocrResultRepo, getStorageService)
log.Printf("OCR workers started for all storage providers")
```

## エラーハンドリング

### ✓ 実装済み

1. **エラーログテーブル (`ocr_error_logs`)**
   - エラー発生時の詳細情報を記録
   - エラータイプ: `storage_error`, `ocr_error`, `db_error`, `panic`

2. **panic回復**
   - `defer + recover`でpanicをキャッチしてログに記録

3. **リトライ機構**
   - 最大3回までリトライ
   - `ocr_results`テーブルにステータスを保存し、`ocr_error_logs`に記録

## データベース

### ✓ 実装済み

1. **`ocr_results`**
   - OCR処理結果を保存
   - `status`: `processing`, `completed`, `failed`
   - `filename`, `storage_provider`, `engine_name`の組み合わせで一意

2. **`ocr_pages`**
   - 複数ページPDFの場合のOCR結果を保存

3. **`ocr_error_logs`** (実装予定)
   - エラーログを記録
   - デバッグや監視に使用

## Webインターフェース

### ✓ 実装済み

- **OCR結果ページ**: `/ocr`
- **OCR処理開始**: 手動でトリガー可能
- **結果表示**: 抽出されたテキストと信頼度スコアを表示

## Docker Compose設定

### ✓ ocr-service

```yaml
ocr-service:
  build:
    context: .
    dockerfile: Dockerfile.ocr
  ports:
    - "50052:50052"
  depends_on:
    - server
    - localstack
    - fake-gcs
    - azurite
```

## まとめ

| 項目 | 状態 |
|------|------|
| ストレージプロバイダー対応（Azure/S3/GCS） | ✓ 実装済み |
| OCRエンジン実装（Tesseract） | ✓ 実装済み |
| OCRエンジン実装（EasyOCR） | ☐ 未実装 |
| OCRエンジン実装（PaddleOCR） | ☐ 未実装 |
| 画像OCR処理 | ✓ 実装済み |
| PDF OCR処理 | ✓ 実装済み（poppler使用） |
| エラーハンドリング | ✓ 実装済み |
| データベース保存 | ✓ 実装済み |

## 将来の拡張（Phase 2）

1. EasyOCR実装
2. PaddleOCR実装
3. 複数エンジン結果比較機能
4. メッセージキュー実装（SQS, Azure Queue, Pub/Sub）
