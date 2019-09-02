■ ブログ

https://yhidetoshi.hatenablog.com/entry/2019/08/31/145031


■ デプロイ
```
- Goコンパイル
  - $ make build

- ServerlessFrameworkでデプロイ
  - $ sls deploy --aws-profile <PROFILE> --slackurl <SLACK_WEBHOOK_URL>
```


`$ make help`
```
build:             Build binaries
build-deps:        Setup build
deps:              Install dependencies
devel-deps:        Setup development
help:              Show help
lint:              Lint
```
