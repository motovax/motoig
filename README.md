# motoig

Unofficial Instagram private API client for Go.

**motoig is a Go port of [instagrapi](https://github.com/subzeroid/instagrapi).** When implementing or debugging behavior, treat instagrapi as the reference implementation — endpoint shapes, request params, realtime event formats, and session handling should match upstream unless Go-specific constraints require a documented deviation.

## Reference implementation

| | |
|---|---|
| Upstream | [subzeroid/instagrapi](https://github.com/subzeroid/instagrapi) |
| Tracked branch | `master` |
| Symbol map (code) | [`references/instagrapi/refs.go`](references/instagrapi/refs.go) |
| Realtime docs | [instagrapi realtime guide](https://github.com/subzeroid/instagrapi/blob/master/docs/usage-guide/realtime.md) |
| Direct/DM docs | [instagrapi direct guide](https://github.com/subzeroid/instagrapi/blob/master/docs/usage-guide/direct.html) |

### API mapping

| motoig | instagrapi |
|--------|------------|
| `Client.Login` | `LoginMixin.login` — [`mixins/auth.py`](https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/auth.py) |
| `Client.SetSessionID` / `LoginBySessionID` | session restore via `sessionid` — [`mixins/auth.py`](https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/auth.py) |
| `Client.DirectThreads` | `DirectMixin.direct_threads_chunk` — [`mixins/direct.py`](https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/direct.py) |
| `Client.DirectMessages` | `DirectMixin.direct_messages` — [`mixins/direct.py`](https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/direct.py) |
| `Client.DirectSend` | `DirectMixin.direct_send` — [`mixins/direct.py`](https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/direct.py) |
| `Client.RealtimeConnect` | `RealtimeMixin.realtime_connect` — [`mixins/realtime.py`](https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/realtime.py) |
| `Client.RealtimeDirectSubscribe` | `RealtimeClient.direct_subscribe` — [`realtime/client.py`](https://github.com/subzeroid/instagrapi/blob/master/instagrapi/realtime/client.py) |
| `realtime.RealtimeClient` | MQTToT transport — [`realtime/`](https://github.com/subzeroid/instagrapi/tree/master/instagrapi/realtime) |
| `state.State.PrivateRequest` | `PrivateRequestMixin.private_request` — [`mixins/private.py`](https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/private.py) |
| `extractors.ExtractDirect*` | [`extractors.py`](https://github.com/subzeroid/instagrapi/blob/master/instagrapi/extractors.py) |

Lookup URLs from code:

```go
import igref "github.com/motovax/motoig/references/instagrapi"

igref.Ref("Client.DirectThreads")
// → https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/direct.py
```

Realtime `message` events follow instagrapi's nested shape (`{"message": {...}}`). Unwrap before reading patch fields:

```go
flat := igref.UnwrapMessagePayload(payload)
// flat["path"], flat["op"], flat["thread_id"], flat["text"], ...
```

## Porting checklist

When adding a feature from instagrapi:

1. Locate the Python method in `instagrapi/mixins/` or `instagrapi/realtime/`.
2. Port params, payload fields, and response parsing to Go.
3. Add the symbol to `references/instagrapi/refs.go`.
4. Add a `// Reference: https://github.com/subzeroid/instagrapi/blob/master/...` comment on the Go method.
5. Add or extend tests using fixture JSON from upstream behavior.

## Install

```bash
go get github.com/motovax/motoig
```

## Quick start

```go
package main

import (
    "context"
    "fmt"

    "github.com/motovax/motoig"
)

func main() {
    ctx := context.Background()
    client := motoig.New()

    if err := client.SetSessionID(ctx, "<sessionid-from-browser-cookies>"); err != nil {
        panic(err)
    }

    threads, err := client.DirectThreads(ctx, 20)
    if err != nil {
        panic(err)
    }
    fmt.Printf("inbox threads: %d\n", len(threads))
}
```

Multi-account session storage uses `Manager` with SQLite or JSON backends — same role as instagrapi's `dump_settings` / `load_settings`.

## Status

Experimental. Instagram's private API and MQTToT transport can change without notice. Validate against instagrapi when something breaks upstream.