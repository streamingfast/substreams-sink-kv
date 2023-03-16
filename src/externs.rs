use {
    crate::{
        pb::types::{KvPairs,KvPair, KvKeys},
        memory
    },
    prost::Message,
};
#[link(wasm_import_module = "host")]
extern "C" {
    pub fn get_key(key_ptr: *const u8, key_len: u32, output_ptr: u32) -> u32;
}
pub fn kv_get_key<K: AsRef<str>>(key: K) -> Option<KvPair> {
    let key = key.as_ref();

    unsafe {
        let key_bytes = key.as_bytes();
        let mut output_buf = Vec::with_capacity(8);
        let output_ptr = output_buf.as_mut_ptr();
        let found = get_key(
            key_bytes.as_ptr(),
            key_bytes.len() as u32,
            output_ptr as u32,
        );
        std::mem::forget(output_ptr);

        return if found == 1 {
            Some(KvPair{
                key: key.to_string(),
                value: memory::get_output_data(output_ptr)
            })
        } else {
            None
        };
    }
}

#[link(wasm_import_module = "host")]
extern "C" {
    pub fn get_many_keys(keys_ptr: *const u8, keys_len: u32, output_ptr: u32) -> u32;
}
pub fn kv_get_many_keys(keys: Vec<String>) -> Option<KvPairs> {
    let keys = KvKeys{ keys };
    let keys_bytes = keys.encode_to_vec();

    unsafe {
        let mut output_buf = Vec::with_capacity(8);
        let output_ptr = output_buf.as_mut_ptr();
        let found = get_many_keys(
            keys_bytes.as_ptr(),
            keys_bytes.len() as u32,
            output_ptr as u32,
        );
        std::mem::forget(output_ptr);

        return if found > 0 {
            let v = memory::get_output_data(output_ptr);
            Some(KvPairs::decode(&v[..]).expect("Failed to decode"))
        } else {
            None
        };
    }
}

#[link(wasm_import_module = "host")]
extern "C" {
    pub fn get_by_prefix(prefix_ptr: *const u8, prefix_len: u32, limit: u32, output_ptr: u32) -> u32;
}

pub fn kv_prefix<K: AsRef<str>>(prefix: K, limit_opt: Option<u32>) -> KvPairs {
    let prefix = prefix.as_ref();

    unsafe {
        let prefix_bytes = prefix.as_bytes();
        let mut output_buf = Vec::with_capacity(8);
        let output_ptr = output_buf.as_mut_ptr();
        let mut limit = 0;
        if let Some(l) = limit_opt {
            limit = l;
        }
        let length = get_by_prefix(
            prefix_bytes.as_ptr(),
            prefix_bytes.len() as u32,
                    limit,
            output_ptr as u32,
        );
        std::mem::forget(output_ptr);

        return if length > 0 {
            let v = memory::get_output_data(output_ptr);
            KvPairs::decode(&v[..]).expect("Failed to decode")
        } else {
            KvPairs { pairs: vec![] }
        };
    }
}

#[link(wasm_import_module = "host")]
extern "C" {
    pub fn scan(start_ptr: *const u8, start_len: u32,exclusively_end_ptr: *const u8, exclusively_end_len: u32, limit: u32, output_ptr: u32) -> u32;
}
pub fn kv_scan<K: AsRef<str>>(start: K, exclusively_end: K, limit_opt: Option<u32>) -> KvPairs {
    let start = start.as_ref();
    let exclusively_end = exclusively_end.as_ref();

    unsafe {
        let start_bytes = start.as_bytes();
        let exclusively_end_bytes = exclusively_end.as_bytes();
        let mut output_buf = Vec::with_capacity(8);
        let output_ptr = output_buf.as_mut_ptr();
        let mut limit = 0;
        if let Some(l) = limit_opt {
            limit = l;
        }
        let length = scan(
            start_bytes.as_ptr(),
            start_bytes.len() as u32,
            exclusively_end_bytes.as_ptr(),
            exclusively_end_bytes.len() as u32,
            limit,
            output_ptr as u32,
        );
        std::mem::forget(output_ptr);

        return if length > 0 {
            let v = memory::get_output_data(output_ptr);
            KvPairs::decode(&v[..]).expect("Failed to decode")
        } else {
            KvPairs { pairs: vec![] }
        };
    }
}