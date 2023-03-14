//! Store Implementation for Substreams.

use {
    crate::{
        pb::types::{KvPairs,KvPair},
        externs,
    }
};

pub trait StoreNew {
    fn new() -> Self;
}

pub trait StoreGet {
    fn get<K: AsRef<str>>(&self, key: K) -> Option<KvPair>;
    fn prefix<K: AsRef<str>>(&self, prefix: K) -> KvPairs;
}

pub struct Store {}

impl StoreNew for Store {
    fn new() -> Self {  Self {} }
}

impl StoreGet for Store {
    fn get<K: AsRef<str>>(&self, key: K) -> Option<KvPair> {
        return externs::kv_get_key(key)
    }
    fn prefix<K: AsRef<str>>(&self, key: K) -> KvPairs {
        return externs::kv_scan_prefix(key)
    }
}