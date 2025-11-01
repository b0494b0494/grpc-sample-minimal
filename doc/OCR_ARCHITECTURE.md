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

### Phase 2: 独立マイクロサービス（エンジンごとに別コンテナ）

```
┌─────────────────────────────────┐
│  gRPC Server        │
│  (main.go)          │
└─────────────────────────────────┘
           │
           │
┌─────────────────────────────────┐
│ ApplicationService  │
│  - ocrRouter        ────────────────┐
│  (エンジン選択)      │              │
└─────────────────────────────────┘   │
                                     │
                                     │
                     ┌───────────────────────────────┐
                     │  OCRRouter                    │
                     │  (エンジン名でルーティング)     │
                     └───────────────────────────────┘
                                 │
                    ┌────────────┼────────────┐
                    │            │            │
                    │            │            │
    ┌───────────────┘    ┌───────┘    ┌───────┘
    │                    │            │
    │ gRPC               │ gRPC       │ gRPC
    │                    │            │
┌─────────────────┐ ┌──────────┐ ┌────────────┐
│ OCR Service     │ │ OCR Service│ │ OCR Service│
│ Tesseract       │ │ EasyOCR   │ │ PaddleOCR  │
│ (port 50052)    │ │ (port 50053)│ │ (port 50054)│
│ - TesseractEngine│ │ - EasyOCREngine│ │ - PaddleOCREngine│
└─────────────────┘ └──────────┘ └────────────┘
```

**メリット:**
- Dockerイメージサイズの最適化（必要なエンジンのみ）
- エンジンごとの独立スケーリング
- エンジンの追加・削除が容易
- リソース分離（メモリ、CPU）

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

- **環境変数**: 
  - `OCR_CLIENT_MODE=local` または `grpc`
  - `OCR_TESSERACT_ENDPOINT=http://ocr-tesseract:50052`
  - `OCR_EASYOCR_ENDPOINT=http://ocr-easyocr:50053`
  - `OCR_PADDLEOCR_ENDPOINT=http://ocr-paddleocr:50054`
- **ApplicationServiceの変更**: 最小限に抑える
- **拡張性**: エンジンごとに独立したコンテナで実装可能
- **イメージサイズ最適化**: 各エンジンの必要な依存関係のみを含める

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
5. **イメージサイズ最適化**: エンジンごとに別コンテナで必要な依存関係のみ
6. **リソース分離**: エンジンごとに独立したメモリ・CPU制限
7. **スケーリング**: 需要に応じて特定のエンジンのみスケール可能

## 実装フェーズ

### Phase 2A: エンジンごとに別コンテナ設計に移行（現在）
- OCR_ARCHITECTURE.mdの設計を更新
- docker-compose.ymlに複数OCRサービス定義の準備
- OCRRouterの実装計画

### Phase 2B: Tesseractコンテナの分離（次フェーズ）
- 既存のocr-serviceをocr-tesseract-serviceに変更
- 独立したDockerfile.tesseractの作成

### Phase 2C: 新規エンジンの追加（段階的）
- EasyOCRコンテナの追加（docker-compose.yml、Dockerfile.easyocr）
- PaddleOCRコンテナの追加（docker-compose.yml、Dockerfile.paddleocr）
- 各エンジンの実装（domain/easyocr.go、domain/paddleocr.go）
