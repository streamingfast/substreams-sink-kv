//! A library for writing Substream Sink Wasm Query handlers.


mod externs;

pub mod helpers;
pub mod pb;
pub mod store;


/// A prelude that makes all store traits available.
///
/// Add the following code to import all traits listed below at once.
///
/// ```
/// use substreams_sink_kv::prelude::*;
/// ```
pub mod prelude {
    pub use crate::store::{
        Store,
        StoreGet,StoreNew,
    };
}