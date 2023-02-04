# substreams-sink-kv

This crate is a simple wrapper around formatting substreams output 
to a kv store.

## Create a kv_out module in your substreams

```rust
// lib.rs

use substreams_sink_kv::pb::kv::KvOperations;

...

pub fn kv_out(
    ... some stores ...
) -> Result<KvOperations, Error> {

    let mut kv_ops: KvOperations = Default::default();

    // process your data, push to your KV
    kv_ops.push_new(someKey, someValue, ordinal);
    kv_ops.push_delete(anotherKey, anotherOrdinal);

    Ok(kv_ops)
}
```
