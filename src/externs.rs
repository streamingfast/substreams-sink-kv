

use {
    crate::pb::types::{KvPairs,KvPair},
    std::{
        convert::TryInto,
        slice,
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
                value: get_output_data(output_ptr)
            })
        } else {
            None
        };
    }
}


#[link(wasm_import_module = "host")]
extern "C" {
    pub fn prefix_scan(prefix_ptr: *const u8, prefix_len: u32, output_ptr: u32) -> u32;
}

pub fn kv_scan_prefix<K: AsRef<str>>(prefix: K) -> KvPairs {
    let prefix = prefix.as_ref();

    unsafe {
        let prefix_bytes = prefix.as_bytes();
        let mut output_buf = Vec::with_capacity(8);
        let output_ptr = output_buf.as_mut_ptr();
        let length = prefix_scan(
            prefix_bytes.as_ptr(),
            prefix_bytes.len() as u32,
            output_ptr as u32,
        );
        std::mem::forget(output_ptr);

        return if length > 0 {
            let v = get_output_data(output_ptr);
            KvPairs::decode(&v[..]).expect("Failed to decode")
        } else {
            KvPairs { pairs: vec![] }
        };
    }
}



pub fn read_u32_from_heap(output_ptr: *mut u8, len: usize) -> u32 {
    unsafe {
        let value_bytes = slice::from_raw_parts(output_ptr, len);
        let value_raw_bytes: [u8; 4] = value_bytes.try_into().expect("error reading raw bytes");
        return u32::from_le_bytes(value_raw_bytes);
    }
}
pub fn get_output_data(output_ptr: *mut u8) -> Vec<u8> {
    unsafe {
        let value_ptr: u32 = read_u32_from_heap(output_ptr, 4);
        let value_len: u32 = read_u32_from_heap(output_ptr.add(4), 4);

        let ret = Vec::from_raw_parts(
            value_ptr as *mut u8,
            value_len as usize,
            value_len as usize,
        );

        ret
    }
}