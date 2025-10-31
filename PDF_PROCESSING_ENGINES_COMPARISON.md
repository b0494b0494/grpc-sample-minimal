# 各OCRエンジンのPDF処理方法

## 現在の実装

### Tesseract OCR（現在使用中）
- **PDF変換**: poppler-utilsを使用
- **問題**: Alpine Linuxでmusl libcの互換性問題
- **代替案**: poppler-utilsでPDF処理

## 将来の実装予定PDF処理方法

### EasyOCR（Phase 2予定）
- **実装**: Python API + Dockerコンテナ
- **PDF変換方法**:
  - **pdf2image** (Pillow + poppler)を使用
  - または **PyMuPDF** (fitz)を使用
  - Python経由で実装

### PaddleOCR（Phase 2予定）
- **実装**: PaddlePaddle Docker + Python API
- **PDF変換方法**:
  - **pdf2image**を使用
  - または **PyMuPDF**を使用
  - PaddleOCRで画像を処理

## 各PDF変換方法の比較

### 方法1: pdf2image (poppler)を使用
- **メリット**: 高品質な変換結果、Alpine Linuxで動作可能
- **デメリット**: Dockerfileに`poppler-utils`が必要

### 方法2: PyMuPDF (fitz)を使用（Pythonのみ）
- **メリット**: Pythonライブラリとして簡単に利用可能
- **デメリット**: GoでTesseractを使う場合には適用不可

## 推奨実装方法

### 現在のPDF変換実装

**方針**:
1. すべてのOCRエンジンで共通のPDF変換処理を使用
2. Alpine Linuxでの互換性を考慮
3. 将来の拡張性を考慮

**実装**: `server/domain/pdf_converter.go`で実装済み

## まとめ

### 現在の実装
- **Tesseract**: poppler-utilsでPDF変換済み
- **EasyOCR/PaddleOCR**: 将来の実装予定

### 推奨実装
- **poppler-utils (pdftoppm)**: すべてのエンジンで共通使用
- 将来的にPDF変換処理を統合
- 各OCRエンジンで処理を実行
