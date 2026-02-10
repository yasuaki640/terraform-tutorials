# S3 + CloudFront CV配信 Terraform チュートリアル実装計画

## 概要

S3にPDF（CV/履歴書）を配置し、CloudFront経由でのみアクセス可能にする構成をTerraformで構築します。Origin Access Control (OAC) を使用したセキュアな配信を学びます。

## なぜこの構成が必要か

### CloudFrontを使う理由
1. **セキュリティ**: S3への直接パブリックアクセスを完全ブロック
2. **パフォーマンス**: エッジロケーションによるキャッシュで高速配信
3. **HTTPS対応**: CloudFrontで簡単にHTTPS有効化
4. **コスト最適化**: S3直接アクセスよりデータ転送コストが低い

### OAC vs OAI
| 項目 | OAI（非推奨） | OAC（推奨） |
|------|---------------|-------------|
| サポート状況 | レガシー | 2022年〜標準 |
| SSE-KMS暗号化 | 非対応 | 対応 |
| AWS署名v4 | 非対応 | 対応 |

---

## アーキテクチャ

```
User → CloudFront (HTTPS) → OAC → S3 Bucket (Private)
                                      └── cv.pdf
```

---

## ディレクトリ構成

```
aws-s3-cloudfront-sample/
├── README.md
├── PLAN.md                  ← この計画ファイル
├── modules/
│   ├── s3/
│   │   ├── main.tf          # S3バケット、パブリックアクセスブロック、暗号化
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── cloudfront/
│       ├── main.tf          # OAC、Distribution
│       ├── variables.tf
│       └── outputs.tf
└── environments/
    └── dev/
        ├── main.tf          # モジュール呼び出し + バケットポリシー
        ├── variables.tf
        ├── outputs.tf
        ├── versions.tf
        └── terraform.tfvars.example
```

---

## 実装するAWSリソース

| リソース | Terraformリソース | 目的 |
|----------|-------------------|------|
| S3バケット | `aws_s3_bucket` | PDF保存先 |
| パブリックアクセスブロック | `aws_s3_bucket_public_access_block` | 直接アクセス禁止 |
| バケット暗号化 | `aws_s3_bucket_server_side_encryption_configuration` | SSE-S3暗号化 |
| CloudFront OAC | `aws_cloudfront_origin_access_control` | セキュアなオリジンアクセス |
| CloudFront Distribution | `aws_cloudfront_distribution` | CDN配信 |
| S3バケットポリシー | `aws_s3_bucket_policy` | OACからのみアクセス許可 |

---

## TODO(human) - 学習者が実装する箇所

教育効果を高めるため、以下の重要な設計判断箇所は学習者に実装させます：

| 箇所 | ファイル | 学習ポイント | 難易度 |
|------|----------|-------------|--------|
| S3 Public Access Block の4設定 | `modules/s3/main.tf` | S3セキュリティの基本 | 易 |
| CloudFront `allowed_methods` | `modules/cloudfront/main.tf` | HTTPメソッドの選択 | 易 |
| CloudFront `viewer_protocol_policy` | `modules/cloudfront/main.tf` | HTTPS強制の重要性 | 易 |
| S3 Bucket Policy の `AWS:SourceArn` 条件 | `environments/dev/main.tf` | IAMポリシーとOACの連携 | 中 |

---

## 実装手順

### ✅ Step 1: ディレクトリ構造の作成（完了）
```bash
mkdir -p aws-s3-cloudfront-sample/modules/{s3,cloudfront}
mkdir -p aws-s3-cloudfront-sample/environments/dev
```

### ⬜ Step 2: S3モジュールの実装
- [ ] `modules/s3/variables.tf` - 入力変数定義
- [ ] `modules/s3/main.tf` - バケット、パブリックアクセスブロック、暗号化
- [ ] `modules/s3/outputs.tf` - バケットARN、ドメイン名の出力

### ⬜ Step 3: CloudFrontモジュールの実装
- [ ] `modules/cloudfront/variables.tf` - 入力変数定義
- [ ] `modules/cloudfront/main.tf` - OAC、Distribution（TODO(human)含む）
- [ ] `modules/cloudfront/outputs.tf` - Distribution URL出力

### ⬜ Step 4: 環境設定の実装
- [ ] `environments/dev/versions.tf` - Provider設定
- [ ] `environments/dev/variables.tf` - 環境変数
- [ ] `environments/dev/main.tf` - モジュール呼び出し + バケットポリシー（TODO(human)含む）
- [ ] `environments/dev/outputs.tf` - CloudFront URLの出力
- [ ] `environments/dev/terraform.tfvars.example` - サンプル値

### ⬜ Step 5: README.mdの作成
- [ ] 学習目標、アーキテクチャ図、使い方の説明

---

## 作成対象ファイル一覧

すべて新規作成：

1. `aws-s3-cloudfront-sample/modules/s3/variables.tf`
2. `aws-s3-cloudfront-sample/modules/s3/main.tf`
3. `aws-s3-cloudfront-sample/modules/s3/outputs.tf`
4. `aws-s3-cloudfront-sample/modules/cloudfront/variables.tf`
5. `aws-s3-cloudfront-sample/modules/cloudfront/main.tf`
6. `aws-s3-cloudfront-sample/modules/cloudfront/outputs.tf`
7. `aws-s3-cloudfront-sample/environments/dev/versions.tf`
8. `aws-s3-cloudfront-sample/environments/dev/variables.tf`
9. `aws-s3-cloudfront-sample/environments/dev/main.tf`
10. `aws-s3-cloudfront-sample/environments/dev/outputs.tf`
11. `aws-s3-cloudfront-sample/environments/dev/terraform.tfvars.example`
12. `aws-s3-cloudfront-sample/README.md`

---

## 検証方法

### デプロイ後の動作確認

```bash
# 1. Terraform適用
cd environments/dev
terraform init
terraform plan
terraform apply

# 2. CloudFront URLでアクセス確認（成功するはず）
curl -I https://d1234xxxx.cloudfront.net/cv.pdf

# 3. S3直接URLでアクセス確認（403エラーになるはず）
curl -I https://bucket-name.s3.ap-northeast-1.amazonaws.com/cv.pdf

# 4. PDFアップロード
aws s3 cp cv.pdf s3://your-bucket-name/ --content-type "application/pdf"
```

### クリーンアップ

```bash
terraform destroy
```

---

## コスト情報（注意事項）

| リソース | 課金体系 | 目安 |
|----------|----------|------|
| S3 | 保存量 + リクエスト | 数十円/月（小規模） |
| CloudFront | 転送量 + リクエスト | 無料枠: 1TB/月 |

⚠️ 学習後は `terraform destroy` を忘れずに実行してください。

---

## 作業再開時のコマンド

```bash
cd /Users/yasuaki640/ghq/github.com/yasuaki640/terraform-tutorials/aws-s3-cloudfront-sample
cat PLAN.md  # この計画を確認
```
