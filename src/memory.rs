use {
    std::{
        convert::TryInto,
        slice,
    },
};


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
        Vec::from_raw_parts(
            value_ptr as *mut u8,
            value_len as usize,
            value_len as usize,
        )
    }
}