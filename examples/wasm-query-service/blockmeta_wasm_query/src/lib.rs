extern crate core;

mod helper;
mod pb;

use crate::pb::blockmeta::BlockMeta;
use crate::pb::service::{Block, GetBlockInfoRequest, GetMonthRequest, GetYearRequest, Month, Months};
use prost::Message;
use substreams_sink_kv::prelude::*;
#[allow(unused_imports)]
use wasmedge_bindgen::*;
use wasmedge_bindgen_macro::*;

#[wasmedge_bindgen]
pub fn eth_service_v1_blockmeta_getblockinfo(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = GetBlockInfoRequest::decode(&v[..]).expect("Failed to decode");
    let store = Store::new();

    let start_date = helper::parse_date(req.start)?;
    let mut end_date = helper::parse_date(req.end)?;
    end_date.incr();

    let start = format!("month:first:{}", start_date.key());
    let end = format!("month:first:{}", end_date.key());
    let kv_pairs = store.scan(start, end, None);

    let mut out = Months { months: vec![] };

    for kv_pair in kv_pairs.pairs {
        out.months.push(Month {
            year: helper::parse_year(&kv_pair.key),
            month: helper::parse_month(&kv_pair.key),
            first_block: value_to_block(&kv_pair.value),
            last_block: None,
        })
    }

    let start = format!("month:last:{}", start_date.key());
    let end = format!("month:last:{}", end_date.key());
    let kv_pairs = store.scan(start, end, None);

    for (i, kv_pair) in kv_pairs.pairs.iter().enumerate() {
        out.months[i].last_block = value_to_block(&kv_pair.value);
    }

    return Ok(out.encode_to_vec());
}

#[wasmedge_bindgen]
pub fn eth_service_v1_blockmeta_getmonth(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = GetMonthRequest::decode(&v[..]).expect("Failed to decode");
    let store = Store::new();

    let first_key = format!("month:first:{}{}", req.year, req.month);
    let last_key = format!("month:last:{}{}", req.year, req.month);

    let kv_pairs_opt = store.get_many(vec![first_key.to_string(), last_key.to_string()]);

    match kv_pairs_opt {
        Some(kv_pairs) => {
            let out = Month {
                year: req.year,
                month: req.month,
                first_block: value_to_block(kv_pairs.pairs[0].value.as_ref()),
                last_block: value_to_block(kv_pairs.pairs[1].value.as_ref()),
            };
            Ok(out.encode_to_vec())
        }
        None => Err(format!("key: {}, {} not found", first_key, last_key)),
    }
}

#[wasmedge_bindgen]
pub fn eth_service_v1_blockmeta_getyear(v: Vec<u8>) -> Result<Vec<u8>, String> {
    let req = GetYearRequest::decode(&v[..]).expect("Failed to decode");
    let store = Store::new();

    let start = format!("month:first:{}01", req.year);
    let end = format!("month:first:{}13", req.year);
    let kv_pairs = store.scan(start, end, Some(12 as u32));

    let mut out = Months { months: vec![] };

    for kv_pair in kv_pairs.pairs {
        out.months.push(Month {
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
        hash: blockmeta.hash,
        parent_hash: blockmeta.parent_hash,
        timestamp: "".to_string(),
    };

    if let Some(tmp) = blockmeta.timestamp {
        blk.timestamp = tmp.to_string()
    }

    return Some(blk);
}
