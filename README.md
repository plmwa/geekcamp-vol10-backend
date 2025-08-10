# 📱 API仕様書: 「Grass Chain」

このプロジェクトは、GitHubのコントリビューション数を利用してモンスターを封印するモバイルアプリのバックエンドAPIです。ユーザーは日々の開発活動（コントリビューション）を通じてモンスターを育て、最終的に封印することを目指します。

---

## 🌟 機能概要

* **GitHubアカウントでの認証**: Firebase Authenticationを利用したログイン機能。
* **コントリビューションの取得**: GitHub GraphQL API v4と連携し、ユーザーの活動を取得。
* **モンスター育成・封印システム**: コントリビューション数に応じてモンスターのHPが減少し、0になると封印が完了します。
* **ユーザーデータの永続化**: ユーザー情報やモンスターの育成状況をFirestoreに保存。

---

## 🛠️ 技術スタック

| カテゴリ | 技術 |
| :--- | :--- |
| **バックエンド** | Go (Gin Framework) |
| **フロントエンド** | React Native (Expo) |
| **データベース** | Firestore |
| **認証** | Firebase Authentication (GitHub OAuth) |
| **外部API** | GitHub GraphQL API v4 |

---

## 📂 プロジェクト構成
あとあと載せます

---

## 🚀 セットアップ手順

1.  **リポジトリをクローン**
    ```bash
    git clone git@github.com:plmwa/geekcamp-vol10-backend.git
    cd geekcamp-vol10-backend
    ```

2.  **依存パッケージをインストール**
    Dockerを使用していないため、Go v.1.24.6を入れてもらうことが必要です。
    ```bash
    go mod download
    ```

3.  **Linterのセットアップ (任意)**
    ローカルでLinterを動かす人は入れてください
    ```bash
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    golangci-lint run
    ```

4.  **ローカルサーバーの起動**
    ```bash
    go run ./cmd/server/main.go
    ```
    サーバーは `http://localhost:8081` で起動します。

---

## 📦 データベース設計 (Firestore)

データベースはNoSQLのCloud Firestoreを使用します。データはコレクションとドキュメントの階層で管理されます。

### ルート階層
アプリケーションのルートには、`monsters` と `users` の2つの主要なコレクションが存在します。

* `monsters` **(コレクション)**
    <br>モンスターのマスターデータを格納します。
    * `{monster_id}` **(ドキュメント)**
      <br>モンスターIDは"001","002"...のstringを使用
        ```json
        // Path: /monsters/001
        // 一例
        {
          "name": "スライム",
          "description": "最も基本的なモンスター。まずはこいつを倒すことから始まる。",
          "imageURL": "https://example.com/images/slime.png",
          "requiredContributions": 30,
        }
        ```

* `users` **(コレクション)**
    <br>各ユーザーのデータを格納します。
    * `{firebase_uid}` **(ドキュメント)**
        <br>認証時に払い出されるFirebase UIDをドキュメントIDとして使用します。
        ```json
        // Path: /users/{firebase_uid}
        {
          "githubUserName": "plmwa",
          "photoURL": "https://avatars.githubusercontent.com/u/12345678?v=4",
          "createdAt": "2025-06-01T10:00:00Z",
          "continuousSealRecord": 0,
          "maxSealRecord": 0,
        }
        ```
        * `sealedMonsters` **(サブコレクション)**
            <br>そのユーザーが過去に封印したモンスターの履歴を格納します。
            * `{seal_id}` **(ドキュメント)**
                <br>自動生成されたIDを持つドキュメントです。
                ```json
                // Path: /users/{firebase_uid}/sealedMonsters/{seal_id}
                {
                  "monsterId": "001",
                  "monsterName": "スライム",
                  "sealedAt": "2025-07-31T23:50:00Z"
                }
                ```
                ```
        * `currentMonster` **(サブコレクション)**
            <br>そのユーザーが現在封印中のモンスターを表示
            * `{なんでもいいです、自動生成で}` **(ドキュメント)**
                <br>自動生成されたIDを持つドキュメントです。
                ```json
                // Path: /users/{firebase_uid}/currentMonster/{自動生成ID}
                {
                  "monsterId": "002",
                  "progressContributions": 25,
                  "requiredContributions": 30,
                  "lastContributionReflectedAt": "2025-08-08T22:15:00Z",
                  "assignedAt": "2025-08-01T18:00:00Z"
                }
                ```


## 🔌 APIエンドポイント仕様

**認証**: `/health` を除くすべてのエンドポイントで、リクエストヘッダーにFirebase Authenticationによって発行されたIDトークン (`Authorization: Bearer <ID_TOKEN>`) が必要です。

### ヘルスチェック

#### `GET /health`
サーバーの稼働状態を確認します。
* **レスポンス (200 OK)**:
    ```json
    {
      "status": "ok"
    }
    ```

### ユーザー関連

#### `POST /users`
新しいユーザーを登録します。初回ログイン時に使用されます。
* **リクエストボディ**:
    ```json
    {
      "firebase_id":"abcdefg12345",
      "githubUserName": "plmwa",
      "photoURL": "https://avatars.githubusercontent.com/u/12345678?v=4"
    }
    ```
* **レスポンス (201 Created)**: 登録されたユーザー情報。
    ```json
    {
      "githubUserName": "plmwa",
      "photoURL": "https://avatars.githubusercontent.com/u/12345678?v=4",
      "createdAt": "2025-06-01T10:00:00Z",
      "continuousSealRecord": 0,
      "maxSealRecord": 0,
    }
    ```

#### `GET /users/:id`
指定したIDのユーザー情報を取得します。`:id`にはユーザーのFirebase UIDを指定します。
* **レスポンス (200 OK)**:
    ```json
    {
      "githubUserName": "dev-hero-taro",
      "photoURL": "https://avatars.githubusercontent.com/u/12345678?v=4",
      "createdAt": "2025-06-01T10:00:00Z",
      "continuousSealRecord": 3,
      "maxSealRecord": 8,
      "currentMonster": {
        "monsterId": "002",
        "progressContributions": 25,
        "requiredContributions": 30,
        "lastContributionReflectedAt": "2025-08-08T22:15:00Z",
        "assignedAt": "2025-08-01T18:00:00Z"
      }
      "sealedMonsters": [
        {"monsterId":"001","monsterName":"スライム","sealedAt":"2025-08-09T15:54:50.45Z"},
        {"monsterId":"002","monsterName":"デカスライム","sealedAt":"2025-08-10T03:09:04+09:00"}
      ]
    }
    ```

### コントリビューション関連

#### `GET /contributions/:id`
GitHubから最新のコントリビューションを取得し、モンスターの育成状況を更新します。`:id`にはユーザーのFirebase UIDを指定します。

* **レスポンス (200 OK)**: 更新後のモンスターの育成状況。
    ```json
    {
      "monsterId": "002",
      "progressContributions": 25,
      "requiredContributions": 30,
      "lastContributionReflectedAt": "2025-08-09T22:50:00Z", // 更新日時
      "assignedAt": "2025-08-01T18:00:00Z"
    }
    ```

#### `GET /contributions/:id`
現在育成中(currentMonster)のモンスターの進捗状況を取得します。`:id`にはユーザーのFirebase UIDを指定します。
* **レスポンス (200 OK)**:
    ```json
    {
      "monsterId": "002",
      "progressContributions": 25,
      "requiredContributions": 30,
      "lastContributionReflectedAt": "2025-08-08T22:15:00Z",
      "assignedAt": "2025-08-01T18:00:00Z"
    }
    ```

## エンドポイントテスト
#### `POST /users/`
```
curl -X POST http://localhost:8081/users -H "Content-Type: application/json" -d '{"firebaseId":"test-1","githubUserName": "plmwa","photoURL": "https://avatars.githubusercontent.com/u/12345678?v=4"}'
```

#### `GET /users/:id`
```
curl -X GET http://localhost:8081/users/Hce2hzzylPvC2LQ7BATjDwAegcbl
```
#### `GET /contributions/:id`
```
curl -X GET http://localhost:8081/contributions/Hce2hzzylPvC2LQ7BATjDwAegcbl
```