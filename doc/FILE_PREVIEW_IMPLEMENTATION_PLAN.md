# ファイルプレビュー機能実装計画

## 概要

ファイル一覧画面にプレビュー機能を追加します。PDF、画像、テキストファイルをフロントエンドで直接プレビューできるようにします。

## 対象ファイル形式

### 1. 画像ファイル
- **対応形式**: PNG, JPG, JPEG, GIF, WEBP, BMP, SVG, ICO
- **表示方法**: `<img>`タグで直接表示
- **実装の簡単さ**: ⭐⭐⭐⭐⭐

### 2. PDFファイル
- **対応形式**: PDF
- **表示方法**: PDF.js または iframe + blob URL
- **実装の簡単さ**: ⭐⭐⭐⭐

### 3. テキストファイル
- **対応形式**: TXT, MD, CSV, JSON, XML, HTML等
- **表示方法**: `<pre>`または`<code>`タグで表示
- **実装の簡単さ**: ⭐⭐⭐⭐⭐

## 実装方法

### コンポーネント構造

```
FileDownload.tsx (既存)
  ↓
FilePreviewModal.tsx (新規)
  ├── ImagePreview (画像)
  ├── PDFPreview (PDF)
  └── TextPreview (テキスト)
```

### データ取得方法

1. **既存のダウンロードAPIを利用**
   - `/api/download-file` エンドポイントを使用
   - ファイルをBlobとして取得
   - Blob URLを作成して表示

2. **新規プレビューAPI（オプション）**
   - `/api/preview-file` エンドポイントを追加
   - メタデータ（ファイルサイズ、形式等）も返す

### 実装詳細

#### 1. ファイル形式判定

```typescript
const getFileType = (filename: string): 'image' | 'pdf' | 'text' | 'unknown' => {
  const ext = filename.toLowerCase().split('.').pop() || '';
  
  const imageExts = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'svg', 'ico'];
  const textExts = ['txt', 'md', 'csv', 'json', 'xml', 'html', 'css', 'js', 'ts'];
  
  if (imageExts.includes(ext)) return 'image';
  if (ext === 'pdf') return 'pdf';
  if (textExts.includes(ext)) return 'text';
  return 'unknown';
};
```

#### 2. プレビューモーダルコンポーネント

```typescript
// FilePreviewModal.tsx
interface FilePreviewModalProps {
  show: boolean;
  filename: string;
  storageProvider: string;
  onHide: () => void;
}

export const FilePreviewModal: React.FC<FilePreviewModalProps> = ({
  show,
  filename,
  storageProvider,
  onHide
}) => {
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileType = getFileType(filename);

  useEffect(() => {
    if (show && filename) {
      loadPreview();
    }
    return () => {
      // Cleanup blob URL
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [show, filename]);

  const loadPreview = async () => {
    setLoading(true);
    setError(null);
    try {
      // Download file as blob
      const blob = await downloadFileAsBlob(filename, storageProvider);
      const url = URL.createObjectURL(blob);
      setPreviewUrl(url);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal show={show} onHide={onHide} size="lg" centered>
      <Modal.Header closeButton>
        <Modal.Title>Preview: {filename}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        {loading && <Spinner />}
        {error && <Alert variant="danger">{error}</Alert>}
        {previewUrl && (
          <>
            {fileType === 'image' && (
              <img src={previewUrl} alt={filename} className="img-fluid" />
            )}
            {fileType === 'pdf' && (
              <iframe src={previewUrl} width="100%" height="600px" />
            )}
            {fileType === 'text' && (
              <TextPreview url={previewUrl} />
            )}
          </>
        )}
      </Modal.Body>
    </Modal>
  );
};
```

#### 3. テキストプレビュー

```typescript
// TextPreview.tsx
const TextPreview: React.FC<{ url: string }> = ({ url }) => {
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(url)
      .then(res => res.text())
      .then(text => {
        setContent(text);
        setLoading(false);
      });
  }, [url]);

  if (loading) return <Spinner />;

  return (
    <pre className="bg-light p-3 rounded" style={{ maxHeight: '500px', overflow: 'auto' }}>
      <code>{content}</code>
    </pre>
  );
};
```

#### 4. ファイルダウンロード（Blob取得）

```typescript
// grpcService.ts に追加
export const downloadFileAsBlob = async (
  filename: string,
  storageProvider: string
): Promise<Blob> => {
  const response = await fetch(
    `${API_BASE_URL}/api/download-file?filename=${encodeURIComponent(filename)}&storageProvider=${storageProvider}`,
    {
      method: 'GET',
      headers: {
        'Authorization': AUTH_TOKEN,
      },
    }
  );

  if (!response.ok) {
    throw new Error(`Failed to download file: ${response.statusText}`);
  }

  return await response.blob();
};
```

#### 5. FileDownload.tsx の更新

```typescript
// FileDownload.tsx に追加
const [previewFilename, setPreviewFilename] = useState<string | null>(null);

// テーブルにプレビューボタンを追加
<Button
  variant="outline-info"
  size="sm"
  onClick={() => setPreviewFilename(file.filename)}
  disabled={getFileType(file.filename) === 'unknown'}
>
  Preview
</Button>

// プレビューモーダルを追加
<FilePreviewModal
  show={!!previewFilename}
  filename={previewFilename || ''}
  storageProvider={storageProvider}
  onHide={() => setPreviewFilename(null)}
/>
```

## 技術的な考慮事項

### PDF.js の使用（オプション）

PDF.jsを使用すると、より高機能なPDFビューアーを実装できます。

```typescript
// PDF.js を使用する場合
import * as pdfjsLib from 'pdfjs-dist';

const PDFPreview: React.FC<{ url: string }> = ({ url }) => {
  const [pdfDoc, setPdfDoc] = useState<any>(null);
  const [pageNum, setPageNum] = useState(1);

  useEffect(() => {
    pdfjsLib.getDocument(url).promise.then(pdf => {
      setPdfDoc(pdf);
    });
  }, [url]);

  // PDFレンダリング処理...
};
```

**メリット**:
- ページング
- ズーム機能
- 検索機能

**デメリット**:
- バンドルサイズが大きくなる
- 実装が複雑

**推奨**: 最初は iframe で実装し、必要に応じて PDF.js を追加

### パフォーマンス最適化

1. **遅延読み込み**: プレビューを表示する時だけファイルをダウンロード
2. **Blob URLのクリーンアップ**: コンポーネントのアンマウント時に `URL.revokeObjectURL()` を呼び出す
3. **キャッシュ**: 同じファイルのプレビューはキャッシュする

### エラーハンドリング

- ファイルが大きすぎる場合（例: > 10MB）はプレビューを無効化
- ファイル形式がサポートされていない場合のメッセージ
- ダウンロードエラーの適切な表示

## 実装タスク

### Phase 1: 基本的なプレビュー機能
- [ ] `FilePreviewModal.tsx` コンポーネントの作成
- [ ] `downloadFileAsBlob` 関数の実装
- [ ] `getFileType` 関数の実装
- [ ] `FileDownload.tsx` にプレビューボタンを追加
- [ ] 画像プレビューの実装
- [ ] テキストプレビューの実装

### Phase 2: PDFプレビュー
- [ ] PDFプレビューの実装（iframe）
- [ ] PDF.js の評価（オプション）

### Phase 3: 改善・最適化
- [ ] エラーハンドリングの強化
- [ ] パフォーマンス最適化
- [ ] 大容量ファイルの制限
- [ ] ローディング状態の改善

## UI/UXデザイン

### プレビューボタン
- ファイル一覧テーブルの「Action」列に追加
- サポートされているファイル形式のみ有効化
- アイコン: 👁️ または "Preview"

### モーダルデザイン
- **サイズ**: Large (`size="lg"`)
- **画像**: レスポンシブ表示（`img-fluid`）
- **PDF**: iframe で表示（高さ600px）
- **テキスト**: スクロール可能な領域（最大高さ500px）
- **背景**: モーダルボディは白背景、テキストは`bg-light`

## ファイルサイズ制限

- **画像**: 10MB以下を推奨
- **PDF**: 20MB以下を推奨
- **テキスト**: 1MB以下を推奨

超過する場合は、プレビューを無効化または警告を表示

## 次のステップ

1. `FilePreviewModal.tsx` コンポーネントを作成
2. `grpcService.ts` に `downloadFileAsBlob` を追加
3. `FileDownload.tsx` を更新してプレビューボタンを追加
4. 画像・テキストプレビューを実装
5. PDFプレビューを実装
6. テストと改善
