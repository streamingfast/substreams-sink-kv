//use crate::change::ToBytes;
use crate::pb::kv::{kv_operation::Type, KvOperation, KvOperations};

impl KvOperations {
    pub fn push_new<K: AsRef<str>, V: AsRef<[u8]>>(&mut self, key: K, value: V, ordinal: u64) {
        self.operations.push(KvOperation {
            key: key.as_ref().to_string(),
            value: value.as_ref().to_vec(),
            ordinal: ordinal,
            r#type: Type::Set.into(),
        })
    }
    pub fn push_delete<V: AsRef<str>>(&mut self, key: V, ordinal: u64) {
        self.operations.push(KvOperation {
            key: key.as_ref().to_string(),
            value: vec![],
            ordinal: ordinal,
            r#type: Type::Delete.into(),
        })
    }
}
