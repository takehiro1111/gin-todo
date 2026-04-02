---
name: new-component
description: React コンポーネントの雛形を生成する。引数にコンポーネント名を指定する。例: /new-component TaskCard
---

以下の手順で React コンポーネントの雛形を作成してください。

## 引数

`$ARGUMENTS` にコンポーネント名が入ります（例: `TaskCard`）。
引数がない場合はコンポーネント名をユーザーに確認してください。

## 作成するファイル

### 1. `frontend/src/components/<Name>/<Name>.tsx`

```tsx
export function <Name>() {
  return <div></div>
}
```

- named export を使う (`export default` 禁止)
- Props は `interface Props` ではなく関数の引数に直接型を埋め込む
  ```tsx
  // Good
  export function TaskCard({ title, done }: { title: string; done: boolean }) { ... }

  // Bad
  interface Props { title: string; done: boolean }
  export function TaskCard({ title, done }: Props) { ... }
  ```
- Props がない場合は引数なしでよい
- .claude/rules/frontend-components.md の規約に従う

### 2. `frontend/src/components/<Name>/index.ts`

```ts
export { <Name> } from './<Name>'
```

## 注意

- ファイルはまだ存在しないことを確認してから作成する
- 作成後、どのような Props が必要か簡単にユーザーに確認する
