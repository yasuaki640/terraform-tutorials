# AWS Lambda Go Sample

このプロジェクトは、Go言語で実装されたAWS Lambda関数のサンプルコードと、それをデプロイするためのTerraformコードを含んでいます。

## 前提条件

- Go 1.24以上
- Terraform
- AWS CLI（AWS認証情報の設定が必要）

## セットアップ

1. Goの依存関係をインストールします：
```bash
go mod download
```

2. Terraformでデプロイします：
   `./deploy.sh` スクリプトを実行すると、Lambda関数のビルドとデプロイが自動的に行われます。
   または、個別にTerraformコマンドを実行することも可能です。
```bash
# 推奨: ビルドとデプロイを一度に行う
./deploy.sh

# または、手動で実行する場合
terraform init
terraform plan
terraform apply # main.go や build.sh の変更に応じてビルドも実行されます
```

## 関数の動作

このLambda関数は、呼び出されると以下のようなJSONレスポンスを返します：

```json
{
  "message": "aws.Context{AWSRequestID:\"xxxx-xxxx-xxxx\", InvokedFunctionArn:\"arn:aws:lambda:ap-northeast-1:123456789012:function:go-lambda-function\", ... (省略) ...}"
}
```

## クリーンアップ

リソースを削除するには：

```bash
terraform destroy
```