mod pb;

use std::slice;
use std::str;
use std::convert::TryInto;
use std::path::Prefix;
#[allow(unused_imports)]
use wasmedge_bindgen::*;
use wasmedge_bindgen_macro::*;
use crate::pb::reader::{Request, Response};
use prost::Message;

#[wasmedge_bindgen]
pub fn sf_reader_v1_eth_get(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = Request::decode(&v[..]).expect("Failed to decode");

    // Integrate calls to `kv_get()` or something
    // that hits the HOST VM (in Go)

    let optValue = kv_get_key(req.key);

    if optValue.is_none() {
        return Ok(Response{value: String::from("not found")}.encode_to_vec())
    }

    let value = optValue.unwrap();
    let output = str::from_utf8(&*value).unwrap();

    return Ok(Response{
        value: output.to_string(),
    }.encode_to_vec());
}

#[wasmedge_bindgen]
pub fn sf_reader_v1_eth_scan(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = Request::decode(&v[..]).expect("Failed to decode");

    let optValues = kv_scan_prefix(req.key);

    if optValues.is_none() {
        return Ok(Response{value: String::from("nothing found")}.encode_to_vec())
    }

    // Get output
    // let mut outputs = <Vec<u8>>
    // for value in optValues {
    //     let value = value.unwrap();
    //     let output = str::from_utf
    // }
    //
    // let value = optValue.unwrap();
    // let output = str::from_utf8(&*value).unwrap();
    //
    // return Ok(Responses{
    //     value: output.to_string(),
    // }.encode_to_vec());
}

#[link(wasm_import_module = "host")]
extern "C" {
    pub fn prefix_scan(prefix: *const String, output_buf: u32) -> u32;
}
pub fn kv_scan_prefix<K: AsRef<str>>(prefix: String) -> Option<Vec<u8>> {
    let key = key.as_ref();

    unsafe {
        let mut output_buf = Vec::with_capacity(20);
        let output_ptr = output_buf.as_mut_ptr();
        let found = prefix_scan(
            prefix.into_string(),
            output_ptr as u32,
        );
        std::mem::forget(output_ptr);
    }
}

#[link(wasm_import_module = "host")]
extern "C" {
    pub fn get_key(key_ptr: *const u8, key_len: u32, output_ptr: u32) -> u32;
}
pub fn kv_get_key<K: AsRef<str>>(key: K) -> Option<Vec<u8>> {
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
            Some(get_output_data(output_ptr))
        } else {
            None
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