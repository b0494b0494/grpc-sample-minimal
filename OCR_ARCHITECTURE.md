# OCRアーキテクチャ概要

## 概要

OCR機能は**マイクロサービスアーキテクチャ**として実装されており、将来的にgRPCで独立したサービスとして展開可能です。

## アーキテクチャ

### Phase 1: ローカル統合アーキテクチャ

```
┌─────────────────────────────────┐
│  gRPC Server        │
│  (main.go)          │
└─────────────────────────────────┘
           │
           │
┌─────────────────────────────────┐
│ ApplicationService  │
│  - ocrClient        ────────────────┐
│  - ocrResultRepo    │              │
└─────────────────────────────────┘   │
                                     │
                                     │
                     ┌───────────────────────────────┐
                     │  OCRClient Interface          │
                     │  (domain/ocr_client.go)       │
                     └───────────────────────────────┘
                                 │
                                 │
                     ┌───────────────────────────────┐
                     │  ocrClientAdapter             │
                     │  (gRPCクライアント)            │
                     └───────────────────────────────┘
                                 │
                                 │
                     ┌───────────────────────────────┐
                     │  OCR Service                  │
                     │  (domain/ocr_service.go)       │
                     └───────────────────────────────┘
                                 │
                                 │
                     ┌───────────────────────────────┐
                     │  TesseractEngine              │
                     │  (domain/tesseract_ocr.go)    │
                     └───────────────────────────────┘
```

### Phase 2: 独立マイクロサービス

```
┌─────────────────────────────────┐
│  gRPC Server        │
│  (main.go)          │
└─────────────────────────────────┘
           │
           │
┌─────────────────────────────────┐
│ ApplicationService  │
│  - ocrClient        ────────────────┐
│  (gRPC)             │              │
└─────────────────────────────────┘   │
                                     │
                                     │
                     ┌───────────────────────────────┐
                     │  OCRClient Interface          │
                     │  (gRPC)                       │
                     └───────────────────────────────┘
                                 │
                                 │
                     ┌───────────────────────────────┐
                     │  ocrClientAdapter             │
                     │  (gRPCクライアント)            │
                     └───────────────────────────────┘
                                 │ gRPC
                                 │
                     ┌───────────────────────────────┐
                     │  OCR Service                  │
                     │  (独立サービス)                │
                     │  - TesseractEngine             │
                     │  - EasyOCREngine                │
                     │  - PaddleOCREngine              │
                     └───────────────────────────────┘
```

## 実装詳細

### 1. OCRClientインターフェースとアダプター

`domain/ocr_client.go`で`OCRClient`インターフェースを定義し、実装を抽象化します。
- **Phase 1**: ローカルで`ocrClientAdapter`が`OCRService`を直接呼び出し
- **Phase 2**: gRPCで`ocrClientAdapter`がgRPCサーバー経由で呼び出し
- **ApplicationService**: 変更なしで利用可能

### 2. 初期化コード

```go
// server/main.go
ocrService := domain.NewOCRService()
tesseractEngine := domain.NewTesseractEngine("jpn+eng")
ocrService.RegisterEngine(tesseractEngine)

ocrClient, _ := domain.NewOCRClient("local", ocrService) // Phase 2: "grpc"
ocrResultRepo, _ := domain.NewOCRResultRepository(ctx)

appService := application.NewApplicationService(
    domainService, 
    storageService, 
    fileRepo,
    ocrClient,      // 追加
    ocrResultRepo,  // 追加
)
```

### 3. 設定オプション

- **環境変数**: `OCR_CLIENT_MODE=local` または `grpc`
- **ApplicationServiceの変更**: 最小限に抑える
- **拡張性**: 将来的に複数のOCRエンジンに対応可能

## ファイル構成

```
server/
├── domain/
│   ├── ocr_service.go        # OCRエンジン管理サービス
│   ├── ocr_client.go         # OCRClientインターフェースとアダプター
│   ├── tesseract_ocr.go      # Tesseract実装
│   ├── database.go           # OCR結果保存
│   └── ...
├── application/
│   └── service.go            # ApplicationServiceがOCRClientを使用
├── ocr/
│   └── main.go               # OCR独立サービス（Phase 2）
└── main.go                   # 初期化
```

## メリット

1. **拡張性**: OCR機能をApplicationServiceから分離
2. **独立性**: OCRサービスを独立して開発・デプロイ可能
3. **テスト容易性**: モックやスタブで簡単にテスト可能
4. **柔軟性**: 将来的に複数エンジン対応も容易
