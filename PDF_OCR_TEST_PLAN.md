# PDF OCRテスト計画

## 概要

✓ **poppler-utilsを使用したPDF処理**を実装
- Dockerfileに`poppler-utils`を追加
- `pdf_converter.go`でPDFを画像に変換
- Tesseractで`processPDF`メソッドを使用

## 実装状況

### 1. 画像OCR処理
- [x] ocr-serviceが正常に起動
- [ ] 画像ファイルのアップロードテスト

### 2. 画像ファイルOCR処理（documents/名前空間）
- [ ] PNG/JPGを`documents/`名前空間にアップロード
- [ ] OCR処理が自動的にトリガーされるか確認

### 3. PDFファイルOCR処理実装
- [ ] PDFファイルを`documents/`名前空間にアップロード
- [ ] OCR処理がトリガーされるか確認
- [ ] PDFが複数ページの場合、各ページのOCR結果が正しく保存されるか確認
- [ ] OCR結果がデータベースに正しく保存されるか確認

### 4. エラーハンドリング
- [ ] 不正なPDFファイルの処理
- [ ] 破損PDFファイルの処理
- [ ] OCRエラーの適切な処理

## テスト手順

1. PDFファイルをアップロードして正常に処理されるか確認
2. OCR処理が正常に完了するか確認
3. 結果をデータベースから取得して表示
