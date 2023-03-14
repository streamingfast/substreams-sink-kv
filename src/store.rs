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
    fn prefix<K: AsRef<str>>(&self, prefix: K, limit: u32) -> KvPairs;
    fn scan<K: AsRef<str>>(&self, start: K, exclusive_end: K, limit: u32) -> KvPairs;

}

pub struct Store {}

impl StoreNew for Store {
    fn new() -> Self {  Self {} }
}

impl StoreGet for Store {
    fn get<K: AsRef<str>>(&self, key: K) -> Option<KvPair> {
        return externs::kv_get_key(key)
    }
    fn prefix<K: AsRef<str>>(&self, key: K, limit: u32) -> KvPairs {
        return externs::kv_prefix(key, limit)
    }
    fn scan<K: AsRef<str>>(&self, start: K, exclusive_end: K, limit: u32) -> KvPairs {
        return externs::kv_scan(start, exclusive_end, limit)
    }
}