# Summary

[backlog](https://backlog.com/ja/) を利用したチケット駆動開発のための git subcommandを作ります。
現在のgitのTopic branchに対して対応するチケット情報を取得(基本タイトルがあれば十分)

まずは要件を見てプランを立てたい
CLIのコマンド体系やこうしたほうが良さそうみたいな提案も歓迎


# 要件

- CLIの実装はcobra使おうかな
- backlogのスクレイピングした結果をcache対応は必須
- 認証はPAT
- 依存関係をあんま増やしたくないのでbacklog sdkは利用せずにraw requestする予定

## sub commands

```
git-backlog
  current   現在 branch の課題タイトル
  list      local branch + 課題タイトル
  open      ブラウザで Backlog 課題を開く
  sync      local branch に出てくる issue key をまとめて cache
  switch    fzf で title を見ながら git switch(unix哲学的にこれをここに実装するかは保留)
```

## git config & environment variables

```
git config --global backlog.baseUrl https://your-space.backlog.jp
git config --global backlog.apiKeyEnv BACKLOG_API_KEY
git config --global backlog.issuePattern '[A-Z][A-Z0-9_]*-[0-9]+'
git config --global backlog.cacheTTL 24h
```

example

```gitconfig
[backlog]
baseUrl = https://your-space.backlog.jp
```


