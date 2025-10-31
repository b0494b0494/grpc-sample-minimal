# テスト計画

## テスト項目

### 1. サービス起動確認 ✓
- [x] server: ポート50051
- [x] ocr-service: ポート50052
- [x] webapp: ポート8080

### 2. 画像ファイルOCR処理（images/ namespace）

#### テストケースA: PNGファイルアップロード
1. ブラウザで http://localhost:8080 を開く
2. Filesページに移動してストレージプロバイダーを選択（Azure/S3/GCS）
3. PNGファイルをアップロード
4. アップロード後、`images/` namespaceに分類されることを確認
5. `/ocr`ページでOCR処理を手動でトリガー
6. 結果にOCR処理結果が表示されることを確認

#### テストケースB: JPGファイルアップロード
- 同様の手順でJPGファイルをテスト

### 3. PDFファイルOCR処理（documents/ namespace）

#### テストケースC: PDFファイルアップロード（推奨）
1. PDFファイルをアップロード
2. アップロード後、`documents/` namespaceに分類されることを確認
3. `/ocr`ページでOCR処理を手動でトリガー
4. PDFが複数ページの場合、各ページのOCR結果が正しく表示されることを確認
5. 結果にPDF全体のOCR結果が表示されることを確認

### 4. エラーハンドリング
- エラーログの確認
- エラー発生時の動作確認
- OCR処理失敗時の適切な処理確認

## 確認項目

### Namespace分類
- PNG/JPG → `images/`
- PDF → `documents/`
- 動画/音声 → `media/`（OCR処理対象外）

### OCR処理
- `images/`と`documents/`の両方でOCR処理を実行

### ログ確認
```bash
docker-compose logs ocr-service | grep -E "OCR|PDF|images"
docker-compose logs server | grep -E "OCR|namespace"
```
