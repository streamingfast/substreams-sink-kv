extern crate core;

mod pb;
mod helper;

#[allow(unused_imports)]
use wasmedge_bindgen::*;
use wasmedge_bindgen_macro::*;
use crate::pb::service::{GetMonthRequest, MonthResponse, Month, Block, GetYearRequest, YearResponse};
use crate::pb::blockmeta::{BlockMeta};
use prost::Message;
use substreams_sink_kv::prelude::*;


#[wasmedge_bindgen]
pub fn eth_service_v1_blockmeta_getmonth(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = GetMonthRequest::decode(&v[..]).expect("Failed to decode");
    let store = Store::new();

    let first_key = format!("month:first:{}{}", req.year,req.month);
    let last_key = format!("month:last:{}{}",req.year,req.month);

    let kv_pairs_opt = store.get_many(vec![first_key.to_string(), last_key.to_string()]);

    match kv_pairs_opt {
        Some(kv_pairs) => {
            let out = MonthResponse{
                month: Some(
                    Month{
                        year: req.year,
                        month: req.month,
                        first_block: value_to_block(kv_pairs.pairs[0].value.as_ref()),
                        last_block: value_to_block(kv_pairs.pairs[1].value.as_ref()),
                    }
                )
            };
            Ok(out.encode_to_vec())
        },
        None => {
            Err(format!("key: {}, {} not found",first_key, last_key))
        }
    }

}

#[wasmedge_bindgen]
pub fn eth_service_v1_blockmeta_getyear(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = GetYearRequest::decode(&v[..]).expect("Failed to decode");
    let store = Store::new();

    let start = format!("month:first:{}01", req.year);
    let end = format!("month:first:{}13", req.year);
    let kv_pairs = store.scan(start, end, Some(12 as u32));

    let mut out = YearResponse{ months: vec![] };

    for kv_pair in kv_pairs.pairs {
        out.months.push(Month{
            year: req.year.to_string(),
            month: helper::parse_month(&kv_pair.key),
            first_block: value_to_block(&kv_pair.value),
            last_block: None,
        })
    }

    let start = format!("month:last:{}01", req.year);
    let end = format!("month:last:{}13", req.year);
    let kv_pairs = store.scan(start, end, Some(12 as u32));

    for (i, kv_pair) in kv_pairs.pairs.iter().enumerate() {
        out.months[i].last_block = value_to_block(&kv_pair.value);
    }

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
