mod pb;
use std::str;
#[allow(unused_imports)]
use wasmedge_bindgen::*;
use wasmedge_bindgen_macro::*;
use crate::pb::service::{GetMonthRequest, MonthResponse, Month, Year, Block, Error};
use crate::pb::blockmeta::{BlockMeta};
use prost::Message;
use substreams_sink_kv::prelude::*;

#[wasmedge_bindgen]
pub fn eth_service_v1_blockmeta_getmonth(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = GetMonthRequest::decode(&v[..]).expect("Failed to decode");
    let store = Store::new();

    let first_key = format!("month:first:{}{}", req.year,req.month);
    let last_key = format!("month:last:{}{}",req.year,req.month);
    println!("first: {}", first_key);
    println!("last_key: {}", last_key);

    let optValues = store.get_many(vec![first_key.to_string(), last_key.to_string()]);

    if optValues.is_none() {
        let out = MonthResponse{
            error: Some(Error{
                code: 404,
                message: format!("key: {}, {} not found",first_key, last_key).to_string(),
            }),
            month: None
        };
        return Ok(out.encode_to_vec())
    }

    let values = optValues.unwrap();

    let out = MonthResponse{
        error: None,
        month: Some(
            Month{
                year: req.year,
                month: req.month,
                first_block: value_to_block(values.pairs[0].value.as_ref()),
                last_block: value_to_block(values.pairs[1].value.as_ref()),
            }
        )
    };
    return Ok(out.encode_to_vec());
}

pub fn value_to_block(v: &Vec<u8>) -> Option<Block> {
    let blockmeta = BlockMeta::decode(&v[..]).expect("failed to decode blockmeta");
    let mut blk = Block {
        number: blockmeta.number,
        hash: format!("{:02X?}", blockmeta.hash),
        parent_hash: format!("{:02X?}", blockmeta.parent_hash),
        timestamp: "".to_string(),
    };

    if let Some(tmp) = blockmeta.timestamp {
        blk.timestamp = tmp.to_string()
    }

    return Some(blk)
}
