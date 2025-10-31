# Namespace更新サマリー

## 変更内容

### 1. 新しく`images/` namespaceを追加

**変更前**:
- `documents/`: ドキュメントファイル（PDF, DOCなど）
- `media/`: 画像・動画ファイルすべて
- `others/`: その他

**変更後**:
- `documents/`: ドキュメントファイル（PDF, DOCなど）
- `images/`: 画像ファイル（PNG, JPG, GIF, WEBP, BMP, SVG, ICOなど）
- `media/`: 動画・音声ファイル（MP4, AVI, MOV, MP3, WAVなど）
- `others/`: その他

### 2. OCR処理条件の更新

**変更前**:
```go
if namespace == "documents" && s.ocrClient != nil {
    // OCR処理
}
```

**変更後**:
```go
if (namespace == "documents" || namespace == "images") && s.ocrClient != nil {
    // OCR処理
}
```

### 3. 更新されたファイル

- `server/domain/storage.go`: GetFileNamespaceとBuildStoragePath
- `server/application/service.go`: OCR処理条件
- `server/domain/azure_storage.go`: ListFilesの更新
- `server/domain/gcs_storage.go`: ListFilesの更新
- `server/domain/s3_storage.go`: ListFilesの更新

## 効果

- PNG/JPG画像 → `images/` namespace → **OCR処理可能**
- PDFファイル → `documents/` namespace → **OCR処理可能**
- 動画/音声 → `media/` namespace → OCR処理対象外（適切）

## 実装手順

1. PNG/JPG画像をアップロード → `images/`名前空間に分類
2. OCR処理を自動トリガー
3. OCR結果をデータベースに保存
