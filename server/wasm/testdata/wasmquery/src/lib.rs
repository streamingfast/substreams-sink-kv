mod pb;
use std::str;
#[allow(unused_imports)]
use wasmedge_bindgen::*;
use wasmedge_bindgen_macro::*;
use crate::pb::test::{GetTestRequest, TestPrefixRequest, TestScanRequest, TestGetManyRequest, Tuple, Tuples, OptionalTuples};
use prost::Message;
use substreams_sink_kv::pb::types::KvPair;
use substreams_sink_kv::prelude::*;

#[wasmedge_bindgen]
pub fn sf_test_v1_testservice_testget(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = GetTestRequest::decode(&v[..]).expect("Failed to decode");
    let store = Store::new();

    let optValue = store.get(&req.key);

    if optValue.is_none() {
        return Ok(Tuple{ key: req.key, value: String::from("not found")}.encode_to_vec())
    }

    let kvpair = optValue.unwrap();
    return Ok(to_key_value(&kvpair).encode_to_vec());
}

#[wasmedge_bindgen]
pub fn sf_test_v1_testservice_testgetmany(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = TestGetManyRequest::decode(&v[..]).expect("Failed to decode");
    let store = Store::new();

    let mut keys: Vec<String> = vec![];
    for k in req.keys {
        keys.push(k.to_string())
    }

    let opt_kv_pairs = store.get_many(keys);

    if opt_kv_pairs.is_none() {
        return Ok(OptionalTuples{ error: "Not Found".to_string(), pairs: vec![]}.encode_to_vec())
    }

    let kv_pairs = opt_kv_pairs.unwrap();
    let mut response = OptionalTuples{ pairs: vec![], error: "".to_string()};
    for kv_pair in kv_pairs.pairs {
        response.pairs.push(to_key_value(&kv_pair));
    }
    return Ok(response.encode_to_vec());
}


#[wasmedge_bindgen]
pub fn sf_test_v1_testservice_testprefix(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = TestPrefixRequest::decode(&v[..]).expect("Failed to decode");
    let store = Store::new();

    let kv_pairs = store.prefix(&req.prefix ,req.limit as u32);

    let mut response = Tuples{ pairs: vec![]};
    for kv_pair in kv_pairs.pairs {
        response.pairs.push(to_key_value(&kv_pair));
    }
    return Ok(response.encode_to_vec());
}

#[wasmedge_bindgen]
pub fn sf_test_v1_testservice_testscan(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = TestScanRequest::decode(&v[..]).expect("Failed to decode");
    let store = Store::new();

    let kv_pairs = store.scan(&req.start,&req.exclusive_end, req.limit as u32);

    let mut response = Tuples{ pairs: vec![]};
    for kv_pair in kv_pairs.pairs {
        response.pairs.push(to_key_value(&kv_pair));
    }
    return Ok(response.encode_to_vec());
}

pub fn to_key_value(kv_pair: &KvPair) -> Tuple {
    let output = str::from_utf8(&*kv_pair.value).unwrap();
    Tuple{
        key: kv_pair.key.clone(),
        value: output.to_string(),
    }
}
