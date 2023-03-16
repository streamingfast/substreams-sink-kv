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
    fn get_many(&self, keys: Vec<String>) -> Option<KvPairs>;
    fn prefix<K: AsRef<str>>(&self, prefix: K, limit: Option<u32>) -> KvPairs;
    fn scan<K: AsRef<str>>(&self, start: K, exclusive_end: K, limit: Option<u32>) -> KvPairs;

}

pub struct Store {}

impl StoreNew for Store {
    fn new() -> Self {  Self {} }
}

impl StoreGet for Store {
    fn get<K: AsRef<str>>(&self, key: K) -> Option<KvPair> {
        return externs::kv_get_key(key)
    }
    fn get_many(&self, keys: Vec<String>) -> Option<KvPairs> { return externs::kv_get_many_keys(keys) }
    fn prefix<K: AsRef<str>>(&self, key: K, limit: Option<u32>) -> KvPairs {  return externs::kv_prefix(key, limit) }
    fn scan<K: AsRef<str>>(&self, start: K, exclusive_end: K, limit: Option<u32>) -> KvPairs { return externs::kv_scan(start, exclusive_end, limit) }
}