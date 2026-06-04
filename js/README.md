# fossteams-api JS

JavaScript port of the Go `teams-api` library, using Bun for package management, runtime, and tests.

## Install

```bash
bun install
```

## Test

```bash
bun test
bun test --coverage
```

## Usage

```js
import { New } from './index.js';

const client = await New();
const conversations = await client.getConversations();
```
